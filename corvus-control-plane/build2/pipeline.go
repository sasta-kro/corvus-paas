package build2

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/docker"
)

// DeployerPipeline holds the dependencies needed to run a deployment.
// constructed once in main.go and passed to the handler via handlers.RouterDependencies.
// Each Deploy() call runs independently, the DeployerPipeline itself holds no per-deployment state.
type DeployerPipeline struct {
	database     *db.Database
	dockerClient *docker.DockerClient
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
	dockerClient *docker.DockerClient,
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

// openLogFileForCurrentDeployment creates or opens the log file for a deployment (each deployment has its own log file).
// the log directory is created if it does not exist.
// the file is opened in append mode so redeployments add to the existing log
// rather than overwriting it, preserving the full deployment history in one file.
func (deployerPipeline *DeployerPipeline) openLogFileForCurrentDeployment(slug string) (*os.File, error) {
	err := os.MkdirAll(deployerPipeline.logRoot, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	logPath := filepath.Join(deployerPipeline.logRoot, slug+".log")
	// os.O_APPEND: writes go to the end of the file.
	// os.O_CREATE: create the file if it does not exist.
	// os.O_WRONLY: open for writing only.
	// 0644: owner read/write, group and others read-only.
	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
