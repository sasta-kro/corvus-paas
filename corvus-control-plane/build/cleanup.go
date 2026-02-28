package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

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

/*

Why put these 3 cleanup helper functions here?
The three cleanup methods need access to three pieces of data:

Method - Needs
 `CleanupContainer` - `dockerClient`
 `CleanupFiles` - `assetStorageRoot`
 `CleanupLogFile` - `logRoot`

All three of those fields live on the `DeployerPipeline` struct.
The `DeploymentHandler` struct does not have them and should not have them.

If the cleanup logic were written inline in the handler, the handler would need to know:
- The Docker client (to stop containers)
- The asset storage root path (to construct `/srv/corvus-paas/deployments/<slug>` and call `os.RemoveAll`)
- The log root path (to construct `/srv/corvus-paas/logs/<slug>.log` and call `os.Remove`)

That means either passing all three as additional fields on `DeploymentHandler`,
or importing the `docker` package directly into the handler. Both break the current architecture
where the handler only knows about the database, the logger, and the pipeline.
The handler asks the pipeline to do infrastructure work.
The handler never touches the filesystem or Docker directly.

The method receiver is `*DeployerPipeline` because that struct already holds `dockerClient`,
`assetStorageRoot`, and `logRoot`. Whether the methods live in `pipeline.go` or `cleanup.go`
within the `build` package does not matter to the compiler or the architecture. Both files are
`package build`, so `cleanup.go` can define methods on `DeployerPipeline` exactly the same way
`pipeline.go` does. Splitting them into a separate file is a readability choice, and a reasonable
one since cleanup is a distinct concern from the deploy pipeline flow.
*/
