package build

// pipeline_cleanup.go contains cleanup methods on DeployerPipeline used by the
// DeleteDeployment handler. These methods live here (not in the handler) because
// they need access to dockerClient, assetStorageRoot, and logRoot which belong
// to the pipeline struct.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// TeardownDeployment runs the full teardown sequence for a deployment:
// stop container, remove files, remove log, delete DB row.
// Used by both the DELETE handler and the expiration cleanup loop.
// Returns an error if any critical step fails (container or file removal).
// Log file removal failure is non-fatal and only logged.
func (deployerPipeline *DeployerPipeline) TeardownDeployment(
	teardownContext context.Context,
	deployment *models.Deployment,
) error {
	containerName := "deploy-" + deployment.Slug

	if err := deployerPipeline.CleanupContainer(teardownContext, containerName); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// ===== remove the static files from the asset storage root.
	// os.RemoveAll is idempotent, returns nil if the path does not exist.
	if err := deployerPipeline.CleanupFiles(deployment.Slug); err != nil {
		return fmt.Errorf("failed to remove files: %w", err)
	}

	// ===== remove the log file (associated with the container)
	// non-fatal if this fails, the deployment is already torn down,
	// a leftover log file is not a functional issue.
	if err := deployerPipeline.CleanupLogFile(deployment.Slug); err != nil {
		deployerPipeline.logger.Warn("failed to remove log file (non-fatal)",
			"slug", deployment.Slug,
			"error", err,
		)
	}

	// ===== delete the database record (last)
	// if any previous step failed and returned early, the record still exists,
	// allowing the user to retry the delete request.
	if err := deployerPipeline.database.DeleteDeployment(deployment.ID); err != nil {
		return fmt.Errorf("failed to delete deployment record: %w", err)
	}

	return nil
}

// ========== 3 cleanup helper methods

// CleanupContainer stops and removes the Docker container with the given name.
// this is a thin wrapper around dockerClient.StopAndRemoveContainer that exists
// so the handler package does not need to import the docker package directly.
// the handler calls pipeline.CleanupContainer(), not docker.StopAndRemoveContainer().
func (deployerPipeline *DeployerPipeline) CleanupContainer(ctx context.Context, containerName string) error {
	return deployerPipeline.dockerClient.StopAndRemoveContainer(ctx, containerName)
}

// CleanupFiles removes the deployment's static files directory from the asset storage root.
// the path is constructed from the slug: <assetStorageRoot>/<slug>/
// os.RemoveAll returns nil if the path does not exist, making this idempotent.
func (deployerPipeline *DeployerPipeline) CleanupFiles(slug string) error {
	deploymentDir := filepath.Join(deployerPipeline.assetStorageRoot, slug)
	if err := os.RemoveAll(deploymentDir); err != nil {
		return fmt.Errorf("failed to remove deployment directory %q: %w", deploymentDir, err)
	}
	deployerPipeline.logger.Info("deployment files removed", "path", deploymentDir)
	return nil
}

// CleanupLogFile removes the deployment's log file from the log root.
// the path is constructed from the slug: <logRoot>/<slug>.log
// os.Remove returns an error if the file does not exist, but the caller
// treats this as non-fatal (a missing log file is not a problem).
func (deployerPipeline *DeployerPipeline) CleanupLogFile(slug string) error {
	logPath := filepath.Join(deployerPipeline.logRoot, slug+".log")
	if err := os.Remove(logPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove log file %q: %w", logPath, err)
	}
	deployerPipeline.logger.Info("deployment log file removed", "path", logPath)
	return nil
}
