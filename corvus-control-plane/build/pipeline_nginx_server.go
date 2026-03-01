package build

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/docker"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/util"
)

// deployToNginx handles the shared deployment steps (at the tail end) that are identical
// for all source types (both zip & GitHub). It is called after the source-specific logic
// (zip extraction or git clone + build) has produced a directory of
// static files at contentRoot.
//
// steps performed:
//   - validate the output directory exists within sourceCodeDirectory
//   - copy the output subdirectory to the asset storage root (to <assetStorageRoot>/<thisDeploymentDir>/)
//   - stop and remove any existing container for this slug (handles redeployment)
//   - start nginx container with the asset storage directory bind-mounted
//   - update status to "live"
//
// Returns true if the deployment reached "live" status, false if any step failed.
// all logging and status updates are handled internally via the pipelineLogger.
func (deployerPipeline *DeployerPipeline) deployToNginx(
	deployContext context.Context,
	deployment *models.Deployment,
	contentRoot string,
	pipelineLogger *deployerPipelineLogger,
) bool {

	// ===== Resolve the output directory within the content root
	// contentRoot is the temp directory containing the extracted zip or cloned repo.
	// it is a single deployment's temp working directory (e.g., /tmp/corvus-build-<uuid>/)
	// OutputDirectory is user-provided (e.g., "dist", "build", ".").
	// filepath.Join handles the "." case correctly: Join(contentRoot, ".") == contentRoot.
	outputDirectory := filepath.Join(contentRoot, deployment.OutputDirectory)

	// Verifying if the output directory actually exists
	// a wrong OutputDirectory value is a very common user error and should produce
	// a clear failure message rather than a confusing Docker bind-mount error.
	_, errStat := os.Stat(outputDirectory)
	if errors.Is(errStat, os.ErrNotExist) {
		pipelineLogger.logFailureAndUpdateStatus(
			fmt.Sprintf("output directory %q not found in source content (zip/github)", deployment.OutputDirectory),
			errStat,
		)
		return false
	}
	if errStat != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to stat output directory", errStat)
		return false
	}

	// ===== Copying the output directory to the asset storage root.
	// the asset storage root is the stable location bind-mounted into the Nginx container.
	// working directories are ephemeral (temp). the asset storage root persists across deploys.
	destDirInAssetStorageRoot := filepath.Join(deployerPipeline.assetStorageRoot, deployment.Slug)

	pipelineLogger.logInfo("copying output directory to asset storage root: %s -> %s", outputDirectory, destDirInAssetStorageRoot)
	errCopySourceCodeDir := util.CopyDirectory(outputDirectory, destDirInAssetStorageRoot)
	if errCopySourceCodeDir != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to copy output directory to asset storage root", errCopySourceCodeDir)
		return false
	}
	pipelineLogger.logInfo("files copied to asset storage root")

	// ===== Stop and remove any existing container for this slug

	// this should be a no-op for new deployments (no container exists yet).
	// for GitHub redeploys, this replaces the currently running container.
	// StopAndRemoveContainer is idempotent, returns nil if the container does not exist.
	containerName := "deploy-" + deployment.Slug
	pipelineLogger.logInfo("stopping existing container if present: %s", containerName)
	errStopAndRemoveContainer := deployerPipeline.dockerClient.StopAndRemoveContainer(deployContext, containerName)
	if errStopAndRemoveContainer != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to remove existing container", errStopAndRemoveContainer)
		return false
	}

	// ===== Starting the Nginx container
	pipelineLogger.logInfo("starting nginx container: %s", containerName)
	errCreateAndStartNginxContainer := deployerPipeline.dockerClient.CreateAndStartNginxContainer(deployContext, docker.NginxContainerConfig{
		ContainerName:       containerName,
		Slug:                deployment.Slug,
		HostSourceDirectory: destDirInAssetStorageRoot,
		TraefikNetwork:      deployerPipeline.traefikNetwork,
	})
	if errCreateAndStartNginxContainer != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to start nginx container", errCreateAndStartNginxContainer)
		return false
	}
	pipelineLogger.logInfo("nginx container started successfully")

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
		return false
	}

	pipelineLogger.logInfo("deployment complete. site is live at http://%s.localhost", deployment.Slug)
	// dw about the url being http and https since this is just for internal routing between traefik and docker
	deployerPipeline.logger.Info("deployment live",
		"id", deployment.ID,
		"slug", deployment.Slug,
		"url", "http://"+deployment.Slug+".localhost",
	)

	return true
}
