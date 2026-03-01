package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

// buildImage is the Docker image used for ephemeral build containers.
// node:20-alpine covers the majority of static site generators:
// React/Vite, Next.js (static export), Vue, Svelte, Astro, and any
// npm/yarn/pnpm-based build toolchain.
// In future (v2), this could be made configurable per deployment so users can
// specify a custom build image (eg, python, kotlin, go, etc).
const buildImage string = "node:20-alpine"

// RunEphemeralBuildContainerConfig holds the parameters for RunEphemeralBuildContainer()
// grouping them in a struct keeps the function signature stable as
// more options are added (eg, memory limits, custom image in v2).
type RunEphemeralBuildContainerConfig struct {
	// ContainerName is the Docker container name.
	// used "build-<slug>" to distinguish from "deploy-<slug>" serving containers.
	ContainerName string

	// BuildCommand is the shell command to execute inside the container.
	// passed to `sh -c` so shell operators (&&, ||, pipes) are interpreted.
	// example: "npm ci && npm run build"
	BuildCommand string

	// HostSourceDirectory is the absolute path on the host filesystem
	// containing the cloned repository. This directory is bind-mounted
	// read-write into the container at /workspace so the build process
	// can read source files and write output (e.g. dist/ folder) back
	// to the same directory on the host.
	HostSourceDirectory string

	// EnvironmentVariables is a list of KEY=VALUE strings passed to the
	// container as environment variables. The build process sees them as
	// normal env vars (eg NODE_ENV=production).
	// nil or empty slice means no extra env vars.
	EnvironmentVariables []string

	// LogWriter receives the combined stdout and stderr output from the
	// build process. Typically the deployment log file on disk.
	LogWriter io.Writer
}

// RunEphemeralBuildContainer creates and runs an ephemeral Docker container that
// executes a build command inside a bind-mounted source directory.
//
// Lifecycle:
//
//	Pull the build image if not already cached locally
//	Create the container with the source directory bind-mounted at /workspace
//	Start the container (the build command begins executing)
//	Wait for the container to exit
//	Read all container logs and write them to the LogWriter
//	Remove the container (deferred, runs on both success and failure)
//	Check the exit code: 0 = success, non-zero = build failure
//
// The container runs without a TTY, so Docker multiplexes stdout and stderr
// using its 8-byte header protocol. stdcopy.StdCopy is used to demultiplex
// the stream into clean text for the log file.
func (dockerClient *DockerClient) RunEphemeralBuildContainer(
	buildContext context.Context,
	config RunEphemeralBuildContainerConfig,
) error {
	// ===== pull image if not already present
	pullError := dockerClient.pullImageIfNotPresent(buildContext, buildImage)
	if pullError != nil {
		return fmt.Errorf("failed to pull build image %q: %w", buildImage, pullError)
	}

	// ===== container config
	// the build command is wrapped in `sh -c` so that shell operators
	// (&&, ||, ;, pipes) in the user-provided command string are interpreted
	// by the shell rather than treated as literal arguments.
	// WorkingDir is set to /workspace so relative paths in the build command
	// (eg "npm run build" looking for package.json) resolve correctly.
	containerInternalConfig := &container.Config{
		Image:      buildImage,
		Cmd:        []string{"sh", "-c", config.BuildCommand},
		WorkingDir: "/workspace",
		Env:        config.EnvironmentVariables,

		// this fixes the permission error that doesn't allow deleting a temp folder in /tmp
		// This happens because the build container ran as root inside the container.
		// When npm ci && npm run build wrote the dist/ folder, the files were created
		// owned by root (UID 0) on the host filesystem. The Go backend runs as
		// sasta-docker or whatever the user is, which cannot delete root-owned files.
		// This makes the build process run as the same user that owns the Go process,
		// so all output files are owned by sasta-docker and the cleanup can delete them.
		User: fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}

	// the source directory is bind-mounted read-write so the build process
	// can write output files (e.g. dist/, build/, out/) back to the host.
	// this is different from the Nginx serving container which uses read-only mounts.
	// the build output must persist on the host after the container is removed,
	// because the pipeline copies it to the asset storage root next.
	containerHostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   config.HostSourceDirectory,
				Target:   "/workspace",
				ReadOnly: false, // build process writes output
			},
		},
	}

	// ===== create container
	createResponse, createError := dockerClient.sdk.ContainerCreate(
		buildContext,
		containerInternalConfig,
		containerHostConfig,
		nil, // no networking config needed, build container does not need Traefik routing
		nil, // platform: nil = host native architecture
		config.ContainerName,
	)
	if createError != nil {
		return fmt.Errorf("failed to create build container %q: %w", config.ContainerName, createError)
	}

	dockerClient.logger.Info("build container created",
		"container_id", createResponse.ID[:12],
		"container_name", config.ContainerName,
		"build_command", config.BuildCommand,
	)

	// ===== defer container removal
	// the ephemeral container must be removed after use, regardless of whether
	// the build succeeded or failed. deferring removal here guarantees cleanup
	// even if an error causes an early return below.
	// Force: true handles the edge case where the container is somehow still running.
	defer func() {
		removeError := dockerClient.sdk.ContainerRemove(
			buildContext,
			createResponse.ID,
			container.RemoveOptions{Force: true},
		)
		if removeError != nil {
			// log the removal failure but do not override the original error.
			// a leftover container is not ideal but is not a deployment-breaking issue.
			dockerClient.logger.Warn("failed to remove build container (non-fatal)",
				"container_name", config.ContainerName,
				"error", removeError,
			)
		} else {
			dockerClient.logger.Info("build container removed", "container_name", config.ContainerName)
		}
	}() // anonymous function end

	// ===== starting the build container
	startError := dockerClient.sdk.ContainerStart(
		buildContext,
		createResponse.ID,
		container.StartOptions{},
	)
	if startError != nil {
		return fmt.Errorf("failed to start build container %q: %w", config.ContainerName, startError)
	}

	dockerClient.logger.Info("build container started (building code...)", "container_name", config.ContainerName)

	// ===== wait for container to exit
	/* Why wait?
	ContainerStart  -->  build command begins running (takes 10-60 seconds)
	ContainerWait   -->  blocks the pipeline/pauses Go code here until the build command finishes (basically waitForContainer() )
	                     returns the exit code (0 = success, non-zero = failure)
	ContainerLogs   -->  reads all output (safe because the container already exited)
	ContainerRemove -->  cleans up (deferred)
	*/
	// ContainerWait returns two channels:
	//   - statusChannel: receives the exit status when the container stops
	//   - errorChannel: receives an error if the wait itself fails (ege container removed externally)
	// the "condition" arg/parameter specifies what event to wait for.
	// `container.WaitConditionNotRunning` means "wait until the container is no longer running."
	statusChannel, errorChannel := dockerClient.sdk.ContainerWait(
		buildContext,
		createResponse.ID,
		container.WaitConditionNotRunning,
	)

	// select blocks until one of the channels produces a value.
	// either the container exits normally (statusChannel) or an error occurs (errorChannel).
	var exitCode int64
	select {
	case waitError := <-errorChannel:
		if waitError != nil {
			return fmt.Errorf("error waiting for build container %q: %w", config.ContainerName, waitError)
		}
	case waitStatus := <-statusChannel:
		exitCode = waitStatus.StatusCode
		dockerClient.logger.Info("build container exited",
			"container_name", config.ContainerName,
			"exit_code", exitCode,
		)
	}

	// ===== read container logs
	// logs are read after the container exits to ensure all output has been flushed.
	// ShowStdout and ShowStderr capture both streams.
	// since the container runs without a TTY, Docker multiplexes stdout and stderr
	// using an 8-byte header protocol per frame. stdcopy.StdCopy demultiplexes
	// the stream into clean text.
	logReadCloser, logError := dockerClient.sdk.ContainerLogs(
		buildContext,
		createResponse.ID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		},
	)
	if logError != nil {
		// log reading failure is not fatal to the deployment itself,
		// but it means the user will not see build output in the log file.
		dockerClient.logger.Warn("failed to read build container logs (non-fatal)",
			"container_name", config.ContainerName,
			"error", logError,
		)
	} else {
		defer logReadCloser.Close()
		// stdcopy.StdCopy takes two writers: one for stdout, one for stderr.
		// passing the same writer for both merges the streams into one
		// chronological log in the deployment log file.
		_, copyError := stdcopy.StdCopy(config.LogWriter, config.LogWriter, logReadCloser)
		if copyError != nil {
			dockerClient.logger.Warn("failed to copy build container logs to log file (non-fatal)",
				"container_name", config.ContainerName,
				"error", copyError,
			)
		}
	}

	// ===== check exit code
	// exit code 0 = build succeeded, non-zero = build failed.
	// a non-zero exit code means the build command itself failed
	// (eg, npm install found missing dependencies, a syntax error in the code, a test failure).
	// the specific error details are already written to the log file via the stdout/stderr capture above.
	if exitCode != 0 {
		return fmt.Errorf("build command exited with code '%d' in container %q", exitCode, config.ContainerName)
	}

	return nil
}
