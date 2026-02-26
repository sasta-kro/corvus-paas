package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ========== 3 cleanup helper methods ============

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
