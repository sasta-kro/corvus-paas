// Package docker wraps the Docker SDK client and provides high-level functions
// for the operations the corvus control plane needs:
// starting containers (with Nginx),
// stopping them, and later running ephemeral build containers.
// all Docker SDK calls are isolated here so no other package imports the Docker SDK directly.
// if the Docker interaction strategy changes (eg, switching from SDK to raw socket calls),
// only this package changes.
package docker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	dockerSDKclient "github.com/docker/docker/client"
	/* The syntax 'aliasName "path/to/package"' assigns an alias or local identifier to an imported package.
	 * This override is utilized to:
	 * 1. Resolve naming collisions when multiple imports share the same package name.
	 * 2. Improve code readability by replacing generic names (eg, 'client') with descriptive ones (eg, 'dockerSDKclient').
	 * 3. Provide a stable local name even if the remote package path is long or
	 * differently named than the internal package declaration.
	 *
	 * All exported types and functions from the package are accessed via the alias.
	 * Example: dockerSDKclient.NewClientWithOpts()
	 */)

// DockerClient (docker.DockerClient) is a custom struct that wraps the Docker SDK client with a logger.
// the SDK client itself manages the connection to the Docker daemon over the Unix socket.
// it is safe to share a single DockerClient across goroutines cuz the SDK handles concurrency internally.
type DockerClient struct {
	sdk    *dockerSDKclient.Client
	logger *slog.Logger
}

// NewClient `docker.NewClient()` constructs a Docker DockerClient (my custom defined struct),
// then connects to the Docker daemon using the
// default socket path (/var/run/docker.sock), and performs a ping to verify
// the connection is live before returning.
// returning an error here should cause main.go to exit immediately cuz
// if the Docker daemon is unreachable, the platform cannot function.
func NewClient(logger *slog.Logger) (*DockerClient, error) {
	// > client.NewClientWithOpts is a constructor that initializes the SDK client.
	// > client.FromEnv reads $DOCKER_HOST, $DOCKER_TLS_VERIFY, $DOCKER_CERT_PATH env variables from
	// the OS environment. When those are not set (local dev, direct socket),
	// it falls back to the default Unix socket. ("unix:///var/run/docker.sock")
	// > `client.WithAPIVersionNegotiation()` automatically negotiates the highest API version
	// both the client and the daemon support. without this, a version mismatch between
	// the SDK and the installed Docker daemon causes every API call to fail.
	sdkClient, err := dockerSDKclient.NewClientWithOpts(
		dockerSDKclient.FromEnv,
		dockerSDKclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		// In Go, a foundational rule of error handling is to either
		// handle an error (which may include logging it) or return it, but never both.
		// this is so that there won't be duplicate error logs when the caller handles it too
		return nil, fmt.Errorf("failed to create docker sdk client: %w", err)
	}

	// creating the custom client (which is just a wrapper for sdk client and logger)
	corvusDockerClient := &DockerClient{
		sdk:    sdkClient,
		logger: logger,
	}
	// a `defer corvusDockerClient.sdk.Close()` is not placed here cuz or else if will
	// immediately close after a DockerClient struct/obj is created (which is silly)

	/*
		The 'context' package is the standard Go mechanism for controlling
		the lifecycle, cancellation, and deadlines of operations.

		Because the Docker SDK performs I/O operations across a Unix socket
		or network, these calls can hang indefinitely if the Docker Daemon
		becomes unresponsive (crash, hang)

		'context.Background()' returns a non-nil, empty Context. It acts as
		the "Root" of the context tree. It has no deadline and is never
		canceled. Passing this root context to Docker API calls is like saying
		"I am aware I need to control this connection, and I have made the conscious decision to let it run forever."
		Usually, the plain root isn't passed, a timer is usually added.
	*/
	var backgroundContext context.Context = context.Background()

	// ping the docker daemon immediately to fail fast if Docker is not running.
	// a 5-second timeout is enough for a local socket response.
	// if this times out, Docker is either not running or the socket path is wrong.
	//
	// `context.WithTimeout` allocates internal resources (timers) to track the deadline.
	// It returns context, and CancelFunc (cancelPing) which stops the timer and frees these resources.
	//
	// To prevent memory leaks by cleaning up the context resources as soon
	//    as the operation finishes, rather than waiting for the full timeout duration to elapse.
	// Using 'defer' ensures the cleanup function runs automatically when the
	//    parent function exits. This guarantees resources are freed on success, error,
	//    or panic, without duplicating the cleanup call at every exit path.
	pingContext, cancelPingContextTimer := context.WithTimeout(backgroundContext, 5*time.Second)
	defer cancelPingContextTimer()
	// this is for if the ping finishes earlier than 5 seconds.
	// This makes Go stop the internal counter and the context to not waste the resources.

	// ping() is a simple custom wrapper command that just uses the sdk client's ping function
	err = corvusDockerClient.ping(pingContext)
	if err != nil {
		return nil, fmt.Errorf("docker daemon unreachable: %w", err)
		// serious error cuz this is the whole point of the platform
		// In Go, a foundational rule of error handling is to either
		// handle an error (which may include logging it) or return it, but never both.
		// this is so that there won't be duplicate error logs when the caller handles it too
	}

	logger.Info("docker client connected", "host", sdkClient.DaemonHost())
	return corvusDockerClient, nil
}

// --- helper functions

// ping (`docker.DockerClient.ping()`) sends a lightweight ping request to the Docker daemon.
// used at startup to verify connectivity before the server begins accepting requests.
func (dockerClient *DockerClient) ping(context context.Context) error {
	_, err := dockerClient.sdk.Ping(context) // ping values are not needed so, ignored with `_`
	if err != nil {
		return fmt.Errorf("docker ping failed: %w", err)
	}
	return nil // ping success
}

// Close releases the underlying Docker SDK client connection.
// should be deferred in main.go or a caller immediately after NewClient returns successfully.
func (dockerClient *DockerClient) Close() error {
	// damn, this is a smart way to do it. calling the wrapped function on a return block.
	// this makes the wrapper acts like it is directly returning the error from the wrapped function.
	return dockerClient.sdk.Close()
}
