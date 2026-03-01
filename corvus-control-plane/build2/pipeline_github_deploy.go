package build2

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/docker"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// DeployGitHub is the full pipeline for deploying from a public GitHub repository.
// It clones the repo, optionally runs a build command in an ephemeral container,
// then hands off to the shared deployToNginx helper for the serving steps.
//
// Called as a goroutine from the handler `go pipeline.DeployGitHub(deployment)`
// Uses context.Background() because the HTTP request context is already done
// by the time this goroutine runs.
func (deployerPipeline *DeployerPipeline) DeployGitHub(deployment *models.Deployment) {
	// a new background context is used for the deployerPipeline goroutine.
	// the HTTP request context would be cancelled the moment the handler returns,
	// which would cancel all Docker SDK calls mid-flight.
	// the deployerPipeline must outlive the HTTP request.
	deployContext := context.Background()

	// ===== opening log file and create pipeline logger (same pattern as DeployZipUpload) ---
	logFile, errOpenLogFile := deployerPipeline.openLogFileForCurrentDeployment(deployment.Slug)
	if errOpenLogFile != nil {
		// non-fatal, the pipeline continues without a log file rather than
		// failing the deployment over a logging issue (consistent with DeployZipUpload)
		deployerPipeline.logger.Error("failed to open deployment log file",
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

	// logWriter is the io.Writer passed to cloneGitHubRepo() and RunEphemeralBuildContainer()
	// for capturing git/build output. If the log file failed to open, io.Discard
	// is used instead of nil to avoid a nil writer panic when git or Docker
	// tries to write to it.
	var logWriter io.Writer
	if logFile != nil {
		logWriter = logFile
	} else {
		logWriter = io.Discard
	}

	// ===== Set status as deploying
	pipelineLogger.logInfo("starting github deployment pipeline")
	statusError := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusDeploying)
	if statusError != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to set status to deploying", statusError)
		return
	}

	// ===== Cloning the repo
	// GitHubURL is a *string (pointer) because it is optional for zip deployments (nil-able)
	// for github deployments, the handler validates it is non-nil before inserting.
	if deployment.GitHubURL == nil {
		pipelineLogger.logFailureAndUpdateStatus("github_url is nil",
			fmt.Errorf("missing github_url on deployment %q", deployment.ID),
		)
		return
	}

	// the temp working directory path is generated without creating the directory.
	// git clone creates the destination directory itself. If os.MkdirTemp were used,
	// the directory would already exist and git clone would fail with "destination path already exists."
	// this matches the pattern used in DeployZipUpload where filepath.Join(os.TempDir(), ...)
	// generates a path and lets the subsequent operation create it.
	// it would look like `/<tmp folder>/corvus-build-.../<the cloned repo>`
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

	pipelineLogger.logInfo("cloning repository: %s (branch: %s)", *deployment.GitHubURL, deployment.Branch)
	cloneError := cloneGitHubRepo(*deployment.GitHubURL, deployment.Branch, tempWorkingDir, logWriter)
	if cloneError != nil {
		pipelineLogger.logFailureAndUpdateStatus("git clone failed", cloneError)
		return
	}
	pipelineLogger.logInfo("clone complete")

	// ===== running build command (if provided)
	if deployment.BuildCommand != "" {
		pipelineLogger.logInfo("running build command: %s", deployment.BuildCommand)

		// decode environment variables from JSON string to []string{"KEY=VALUE", ...}
		envVarsList, envDecodeError := decodeEnvVarsToSlice(deployment.EnvironmentVariables)
		if envDecodeError != nil {
			pipelineLogger.logFailureAndUpdateStatus("failed to decode environment variables", envDecodeError)
			return
		}

		buildContainerName := "building-" + deployment.Slug
		buildConfig := docker.RunEphemeralBuildContainerConfig{
			ContainerName:        buildContainerName,
			BuildCommand:         deployment.BuildCommand,
			HostSourceDirectory:  tempWorkingDir,
			EnvironmentVariables: envVarsList,
			LogWriter:            logWriter,
		}

		buildError := deployerPipeline.dockerClient.RunEphemeralBuildContainer(deployContext, buildConfig)
		if buildError != nil {
			pipelineLogger.logFailureAndUpdateStatus("build failed", buildError)
			return
		}
		pipelineLogger.logInfo("build complete")
	} else {
		pipelineLogger.logInfo("no build command specified, skipping build step")
	}

	// ===== Handing it off to the shared deployToNginxHelper
	// tempWorkingDir now contains the cloned (and possibly built) source files.
	// deployToNginx resolves the output directory, copies to asset storage,
	// stops any existing container, starts the nginx container, and sets status to live.
	deployerPipeline.deployToNginx(
		deployContext,
		deployment,
		tempWorkingDir,
		pipelineLogger,
	)
}
