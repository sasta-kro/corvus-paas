package build

// pipeline.go orchestrates the full deployment lifecycle for a single deployment.
// it is the bridge between the HTTP handler (which accepts the request and returns/ends immediately)
// and the infrastructure layer (docker package, filesystem operations).
// all steps run inside a goroutine so the HTTP handler returns 202 without blocking.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/docker"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/util"
)

// DeployerPipeline holds the dependencies needed to run a deployment.
// constructed once in main.go and passed to the handler via handlers.RouterDependencies.
// Each Deploy() call runs independently, the DeployerPipeline itself holds no per-deployment state.
type DeployerPipeline struct {
	database     *db.Database
	dockerClient *docker.Client
	logger       *slog.Logger

	// assetStorageRoot is the base directory on the host where static files are stored.
	// each deployment gets its own subdirectory `<assetStorageRoot>/<slug>/`
	// this subdirectory is bind-mounted into the Nginx container (each project has only access to its assets)
	assetStorageRoot string

	// logRoot is the base directory where per-deployment log files are written.
	// Each deployment log is here `<logRoot>/<slug>.log`
	// TODO logs are written even in v1 so v2 log streaming does not require changing the build system, only adding a read endpoint.
	logRoot string

	// traefikNetwork is the Docker network name (corvus-paas-network) passed to the Nginx container
	// so Traefik can route traffic to it.
	traefikNetwork string
}

// DeployerPipelineConfig groups the configuration values DeployerPipeline needs.
// mirrors the relevant fields from config.Config so the pipeline
// does not import the config package (keeps the dependency graph clean).
type DeployerPipelineConfig struct {
	AssetStorageRoot string
	LogRoot          string
	TraefikNetwork   string
}

// NewDeployerPipeline constructs a DeployerPipeline with its required dependencies.
func NewDeployerPipeline(
	database *db.Database,
	dockerClient *docker.Client,
	logger *slog.Logger,
	config DeployerPipelineConfig,
) *DeployerPipeline {
	return &DeployerPipeline{
		database:         database,
		dockerClient:     dockerClient,
		logger:           logger,
		assetStorageRoot: config.AssetStorageRoot,
		logRoot:          config.LogRoot,
		traefikNetwork:   config.TraefikNetwork,
	}
}

// DeployZipUpload runs the full zip deployment pipeline for the given deployment.
// It is designed to be called as a goroutine from the HTTP handler.
// uploadedFile is an io.Reader over the raw zip bytes from the multipart upload from the user
//
// pipeline steps:
//   - open log file for this deployment
//   - set status to "deploying" (already set at creation, but refreshed here for redeploys)
//   - write uploaded zip bytes to a temp file on disk
//   - extract the zip to a temp working directory
//   - copy the output subdirectory to the asset storage root (to <assetStorageRoot>/<thisDeploymentDir>/)
//   - stop and remove any existing container for this slug (handles redeployment)
//   - start nginx container with the asset storage directory bind-mounted
//   - update status to "live"
//   - clean up temp files
func (deployerPipeline *DeployerPipeline) DeployZipUpload(deployment *models.Deployment, uploadedFile io.Reader) {
	// a background context is used for the deployerPipeline goroutine.
	// the HTTP request context would be cancelled the moment the handler returns,
	// which would cancel all Docker SDK calls mid-flight.
	// the deployerPipeline must outlive the HTTP request.
	deployContext := context.Background()

	// opening the log file for the current deployment (each deployment has its own log file)
	// all deployerPipeline steps write to this log so (TODO) build output is preserved for v2 streaming.
	logFile, errOpenLogFile := deployerPipeline.openLogFile(deployment.Slug)
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

	// deployerPipelineLog() is a helper that generates two log (simultaneously), a raw text entry in
	// a specific deployment log file and a structured log entry in the application’s standard output.
	// Used throughout the deployerPipeline to record each step's outcome.
	deployerPipelineLog := func(format string, args ...any) {
		line := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, args...))

		// prefixes the slog with this pipeline name and slug of the deployment
		deployerPipeline.logger.Info("deployerPipeline", "slug", deployment.Slug, "msg", fmt.Sprintf(format, args...))
		if logFile != nil {
			logFile.WriteString(line) // nolint:errcheck -- log write failures are non-fatal (can ignore ig)
		}
	}

	// failedDeploymentStatusUpdateAndLogIt() is a helper that updates the status to "failed" and logs the reason.
	// Called at any step that cannot be recovered from.
	failedDeploymentStatusUpdateAndLogIt := func(reason string, err error) {
		deployerPipelineLog("FAILED: %s: %v", reason, err)

		// this just updates the failed deployment's status to 'failed' in the db
		dbErr := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusFailed)
		if dbErr != nil {
			deployerPipeline.logger.Error("failed to update status to failed",
				"id", deployment.ID,
				"error", dbErr,
			) // lmao cant do anything about it if the status change fails
		}
	}

	deployerPipelineLog("deployerPipeline started for deployment %q (slug: %s)", deployment.Name, deployment.Slug)

	// ===== Set status as deploying
	// status was set to "deploying" at record creation. refreshing here again
	// handles the redeploy case where a previous run left the status as "live" or "failed".
	errUpdateStatus := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusDeploying)
	if errUpdateStatus != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to update status to deploying", errUpdateStatus)
		return
	}

	// ===== Write the uploaded zip bytes to a temp file on disk.
	// os.CreateTemp() is a build in lib function creates a new FILE in the OS temp directory with a unique name.
	// the file is used as the source for zip extraction and deleted after extraction.
	// this is just created anywhere for now, in the next step, this file will be put in a proper temp working dir
	tmpFileForZipExtraction, errCreateTempFile := os.CreateTemp("", "corvus-upload-*.zip") // `*` is where the random string will be
	if errCreateTempFile != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to create temp file for zip upload", errCreateTempFile)
		return
	}
	// defer removal of the temp zip file so it is cleaned up on any exit path.
	// the file is closed inside the copy block below before extraction begins.
	defer os.Remove(tmpFileForZipExtraction.Name())

	deployerPipelineLog("writing uploaded zip to temp file: %s", tmpFileForZipExtraction.Name())

	// io.Copy streams the uploaded bytes from the request body into the temp file.
	// this avoids loading the entire zip into memory.
	_, errCopyUploadedZipFileToDisk := io.Copy(tmpFileForZipExtraction, uploadedFile)
	if errCopyUploadedZipFileToDisk != nil {
		tmpFileForZipExtraction.Close()
		failedDeploymentStatusUpdateAndLogIt("failed to write uploaded zip to disk", errCopyUploadedZipFileToDisk)
		return
	}
	// close the file before passing its path to the zip extractor.
	// the extractor opens it fresh for reading. Leaving it open for writing
	// would cause a file descriptor conflict on some OS/filesystem combinations.
	tmpFileForZipExtraction.Close()

	// ===== Extracting the zip to a temp working directory.
	// the working directory name includes the deployment ID for traceability.
	tempWorkingDir := filepath.Join(os.TempDir(), "corvus-build-"+deployment.ID)
	defer os.RemoveAll(tempWorkingDir) // clean up the working directory on any exit path

	deployerPipelineLog("extracting zip to working directory: %s", tempWorkingDir)
	errExtractingZipUpload := ExtractZipUpload(tmpFileForZipExtraction.Name(), tempWorkingDir)
	if errExtractingZipUpload != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to extract zip archive", errExtractingZipUpload)
		return
	}
	deployerPipelineLog("zip extracted successfully")

	// ===== Resolve the output directory within the extracted content
	// OutputDirectory is user-provided (eg. "dist", "build", ".").
	// filepath.Join() also handles the "." case correctly: Join(tempWorkingDir, ".") == tempWorkingDir.
	sourceCodeDirectory := filepath.Join(tempWorkingDir, deployment.OutputDirectory)

	// Verifying if the output directory actually exists inside the extracted content.
	// a wrong OutputDirectory value is a very common user error and should produce
	// a clear failure message rather than a confusing Docker bind-mount error.
	_, errOutputDirDoesntExist := os.Stat(sourceCodeDirectory)
	if errors.Is(errOutputDirDoesntExist, os.ErrNotExist) {
		failedDeploymentStatusUpdateAndLogIt(
			fmt.Sprintf("output directory %q not found inside the zip archive", deployment.OutputDirectory),
			errOutputDirDoesntExist,
		)
		return
	}

	// ===== Copying the output directory to the asset storage root.
	// the asset storage root is the stable location bind-mounted into the Nginx container.
	// working directories are ephemeral (temp). the asset storage root persists across deploys.
	destDirInAssetStorageRoot := filepath.Join(deployerPipeline.assetStorageRoot, deployment.Slug)

	deployerPipelineLog("copying output directory to asset storage root: %s -> %s", sourceCodeDirectory, destDirInAssetStorageRoot)
	errCopySourceCodeDir := util.CopyDirectory(sourceCodeDirectory, destDirInAssetStorageRoot)
	if errCopySourceCodeDir != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to copy output directory to asset storage root", errCopySourceCodeDir)
		return
	}
	deployerPipelineLog("files copied to asset storage root")

	// ===== Stop and remove any existing container for this slug
	// this is a no-op for new deployments (no container exists yet).
	// for redeployments, it replaces the currently running container
	containerName := "deploy-" + deployment.Slug
	deployerPipelineLog("stopping existing container if present: %s", containerName)
	errStopAndRemoveContainer := deployerPipeline.dockerClient.StopAndRemoveContainer(deployContext, containerName)
	if errStopAndRemoveContainer != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to remove existing container", errStopAndRemoveContainer)
		return
	}

	// ===== Starting the Nginx container
	deployerPipelineLog("starting nginx container: %s", containerName)
	errCreateAndStartNginxContainer := deployerPipeline.dockerClient.CreateAndStartNginxContainer(deployContext, docker.NginxContainerConfig{
		ContainerName:       containerName,
		Slug:                deployment.Slug,
		HostSourceDirectory: destDirInAssetStorageRoot,
		TraefikNetwork:      deployerPipeline.traefikNetwork,
	})
	if errCreateAndStartNginxContainer != nil {
		failedDeploymentStatusUpdateAndLogIt("failed to start nginx container", errCreateAndStartNginxContainer)
		return
	}
	deployerPipelineLog("nginx container started successfully")

	// ===== Updating container status to live
	errUpdateStatusToLive := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusLive)
	if errUpdateStatusToLive != nil {
		// the container is running but the DB update failed.
		// log the error but do not fail the deployment, the site is actually live.
		// the status inconsistency will be visible in the API response.
		deployerPipeline.logger.Error("container is live but failed to update status to live",
			"id", deployment.ID,
			"slug", deployment.Slug,
			"error", errUpdateStatusToLive,
		)
		return
	}

	deployerPipelineLog("deployment complete. site is live at http://%s.localhost", deployment.Slug)
	// dw about the url being http and https since this is just for internal routing between traefik and docker
	deployerPipeline.logger.Info("deployment live",
		"id", deployment.ID,
		"slug", deployment.Slug,
		"url", "http://"+deployment.Slug+".localhost",
	)
}

// openLogFile creates or opens the log file for a deployment (each deployment has its own log file).
// the log directory is created if it does not exist.
// the file is opened in append mode so redeployments add to the existing log
// rather than overwriting it, preserving the full deployment history in one file.
func (deployerPipeline *DeployerPipeline) openLogFile(slug string) (*os.File, error) {
	err := os.MkdirAll(deployerPipeline.logRoot, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	logPath := filepath.Join(deployerPipeline.logRoot, slug+".log")
	// os.O_APPEND: writes go to the end of the file.
	// os.O_CREATE: create the file if it does not exist.
	// os.O_WRONL: open for writing only.
	// 0644: owner read/write, group and others read-only.
	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
