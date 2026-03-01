package build2

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/docker"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// DeployZipUpload runs the full zip deployment pipeline for the given deployment.
// It is designed to be called as a goroutine from the HTTP handler.
// uploadedFile is an io.Reader over the raw zip bytes from the multipart upload from the user
//
// pipeline steps:
//   - open log file for this deployment
//   - set status to "deploying"
//   - write uploaded zip bytes to a temp file on disk
//   - extract the zip to a temp working directory
//   - hand off to deployToNginx (shared steps: validate output dir, copy to asset storage, start nginx)
func (deployerPipeline *DeployerPipeline) DeployZipUpload(
	deployment *models.Deployment,
	uploadedFile io.ReadCloser,
) {
	// a background context is used for the deployerPipeline goroutine.
	// the HTTP request context would be cancelled the moment the handler returns,
	// which would cancel all Docker SDK calls mid-flight.
	// the deployerPipeline must outlive the HTTP request.
	deployContext := context.Background()

	// opening the log file for the current deployment (each deployment has its own log file)
	// all deployerPipeline steps write to this log so (TODO) build output is preserved for v2 streaming.
	logFile, errOpenLogFile := deployerPipeline.openLogFileForCurrentDeployment(deployment.Slug)
	if errOpenLogFile != nil {
		// if the log file cannot be opened, log the error to the structured logger only.
		// this is not fatal, the deployerPipeline continues without a log file rather than
		// failing the deployment over a logging issue.
		deployerPipeline.logger.Error("failed to open deployment log file",
			"slug", deployment.Slug,
			"error", errOpenLogFile,
		)
	}
	if logFile != nil { // close only if log file opened
		defer logFile.Close()
	}

	// the uploaded file is closed here (not in the HTTP handler) because the handler
	// returns immediately after launching this goroutine. A defer in the handler would
	// close the file while this goroutine is still reading from it.
	defer uploadedFile.Close()

	// setting up the helper logger struct to log to both slog and log file
	pipelineLogger := &deployerPipelineLogger{
		pipeline:   deployerPipeline,
		deployment: deployment,
		logFile:    logFile,
	}

	pipelineLogger.logInfo("Pipeline started for zip deployment %q (slug: %s)", deployment.Name, deployment.Slug)

	// ===== Set status as deploying
	// status was set to "deploying" at record creation. refreshing here again
	// handles the redeploy case where a previous run left the status as "live" or "failed".
	errUpdateStatus := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusDeploying)
	if errUpdateStatus != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to update status to deploying", errUpdateStatus)
		return
	}

	// ===== Write the uploaded zip bytes to a temp zip file on disk

	// os.CreateTemp() creates a new file in the OS default temp directory with a unique name.
	// the uploaded zip bytes are written to this file creating a .zip file, then the file is passed to
	// ExtractZipUpload which reads it for extraction. The file is removed via defer after extraction completes.
	tempZipFileForExtraction, errCreateTempZipFile := os.CreateTemp("", "corvus-upload-*.zip") // `*` is where the random string will be
	if errCreateTempZipFile != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to create temp zip file for zip upload", errCreateTempZipFile)
		return
	}
	// defer removal of the temp zip file so it is cleaned up on any exit path.
	// the file is closed inside the copy block below before extraction begins.
	defer os.Remove(tempZipFileForExtraction.Name())

	pipelineLogger.logInfo("writing uploaded zip to temp file: %s", tempZipFileForExtraction.Name())

	// io.Copy streams the uploaded bytes from the request body into the temp zip file.
	// this avoids loading the entire zip into memory.
	_, errCopyUploadedZipFileToDisk := io.Copy(tempZipFileForExtraction, uploadedFile)
	if errCopyUploadedZipFileToDisk != nil {
		tempZipFileForExtraction.Close()
		pipelineLogger.logFailureAndUpdateStatus("failed to write uploaded zip to disk", errCopyUploadedZipFileToDisk)
		return
	}
	// close the file before passing its path to the zip extractor.
	// the extractor opens it fresh for reading. Leaving it open for writing
	// would cause a file descriptor conflict on some OS/filesystem combinations.
	tempZipFileForExtraction.Close()

	// ===== Extracting the zip to a temp working directory
	// the working directory name includes the deployment ID for traceability.
	tempWorkingDir := filepath.Join(os.TempDir(), "corvus-build-"+deployment.ID)
	defer func() { // clean up the working directory on any exit path
		removeError := os.RemoveAll(tempWorkingDir)
		if removeError != nil {
			deployerPipeline.logger.Warn("failed to remove temp build directory (non-fatal)",
				"path", tempWorkingDir,
				"error", removeError,
			)
		}
	}()

	pipelineLogger.logInfo("extracting zip to working directory: %s", tempWorkingDir)
	errExtractingZipUpload := ExtractZipUpload(tempZipFileForExtraction.Name(), tempWorkingDir)
	if errExtractingZipUpload != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to extract zip archive", errExtractingZipUpload)
		return
	}
	pipelineLogger.logInfo("zip extracted successfully")

	deployerPipeline.deployToNginx(
		deployContext,
		deployment,
		tempWorkingDir,
		pipelineLogger,
	)
}

// RedeployExistingZip re-creates the Nginx container for an existing deployment
// using the files already present in the asset storage root.
// used for zip redeployments where the original upload no longer exists,
// and the only copy of the static files is the deployed directory.
// for github source type, the full clone+build pipeline runs instead (Phase 4).
//
// This method doesn't use deployToNginx helper because it does not copy files (they already exist),
// so it only shares the container stop/start/status update, which is only three steps
// and not worth extracting into a separate method.
func (deployerPipeline *DeployerPipeline) RedeployExistingZip(deployment *models.Deployment) {
	redeployContext := context.Background()

	logFile, errOpenLogFile := deployerPipeline.openLogFileForCurrentDeployment(deployment.Slug)
	if errOpenLogFile != nil {
		deployerPipeline.logger.Error("failed to open deployment log file for redeploy",
			"slug", deployment.Slug,
			"error", errOpenLogFile,
		)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	// setting up the helper logger struct to log to both slog and log file
	pipelineLogger := &deployerPipelineLogger{
		pipeline:   deployerPipeline,
		deployment: deployment,
		logFile:    logFile,
	}
	pipelineLogger.logInfo("redeploy started for deployment %q (slug: %s)", deployment.Name, deployment.Slug)

	// set status to deploying
	if err := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusDeploying); err != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to update status to deploying", err)
		return
	}

	// (zip only) verify the extracted zip files still exist on disk
	deploymentDir := filepath.Join(deployerPipeline.assetStorageRoot, deployment.Slug)
	if _, err := os.Stat(deploymentDir); os.IsNotExist(err) {
		pipelineLogger.logFailureAndUpdateStatus("deployment files not found on disk, cannot redeploy", err)
		return
	}

	// stop and remove the old container
	containerName := "deploy-" + deployment.Slug
	pipelineLogger.logInfo("stopping existing container: %s", containerName)
	errRemoveContainer := deployerPipeline.dockerClient.StopAndRemoveContainer(redeployContext, containerName)
	if errRemoveContainer != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to remove existing container", errRemoveContainer)
		return
	}

	// start a new container pointing to the same files
	pipelineLogger.logInfo("starting nginx container: %s", containerName)
	errStartNginxContainer := deployerPipeline.dockerClient.CreateAndStartNginxContainer(
		redeployContext,
		docker.NginxContainerConfig{
			ContainerName:       containerName,
			Slug:                deployment.Slug,
			HostSourceDirectory: deploymentDir,
			TraefikNetwork:      deployerPipeline.traefikNetwork,
		},
	)
	if errStartNginxContainer != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to start nginx container", errStartNginxContainer)
		return
	}
	pipelineLogger.logInfo("nginx container started successfully")

	// update status to live
	if err := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusLive); err != nil {
		deployerPipeline.logger.Error("container is live but failed to update status after redeploy",
			"id", deployment.ID,
			"slug", deployment.Slug,
			"error", err,
		)
		return
	}

	pipelineLogger.logInfo("redeploy complete. site is live at http://%s.localhost", deployment.Slug)
	deployerPipeline.logger.Info("redeploy live",
		"id", deployment.ID,
		"slug", deployment.Slug,
		"url", "http://"+deployment.Slug+".localhost",
	)
}
