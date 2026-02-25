package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// Why need Nginx? Why not a raw Alpine image?
// A raw Alpine image is strictly a minimal operating system. It does not
// contain a web server process to listen for HTTP requests or serve files.
// A container started from a raw Alpine image would immediately exit or just ignore all web traffic.
//
// The 'nginx:alpine' image packages the Alpine OS with the Nginx web server preconfigured.
// Nginx provides the necessary long-running foreground process that listens
// on port 80, parses incoming HTTP requests routed by Traefik, reads the
// bind-mounted static files from disk, and returns them to the requester.

// nginxImage is the Docker image used for every per-deployment web server.
// nginx:alpine is chosen over nginx:latest because it is significantly smaller
// (~40MB vs ~180MB), has a minimal attack surface, and has everything needed
// to serve static files over HTTP.
const nginxImage string = "nginx:alpine"

// NginxContainerConfigArgs holds the parameters/args the caller passes to CreateAndStartNginxContainer().
// Grouping them in a struct rather than as individual function arguments keeps
// the function signature stable as more options are added (eg custom nginx config in future implementations).
type NginxContainerConfigArgs struct {
	// the Docker container name. convention: "deploy-<slug>"
	ContainerName string

	// Slug is the deployment slug used to construct Traefik routing labels
	// and the public URL. example: "happy-dog-3f9a"
	Slug string

	// HostSourceDirectory is the absolute path on the host (VM) filesystem
	// that contains the static files to serve. this path is bind-mounted
	// read-only into the Nginx container at /usr/share/nginx/html.
	// it must exist on disk before CreateAndStartNginxContainer() is called.
	HostSourceDirectory string

	// TraefikNetwork is the Docker network name that both Traefik and
	// this container must be on for Traefik to proxy traffic to it.
	TraefikNetwork string
}

// ---
// Why are the Create & Start container functions coupled into 1 function?
// Also for Stop & Remove container, why not split them to separate functions?
//
// Cuz of "Immutable Infrastructure Principle"
//
// In Platform-as-a-Service (PaaS) architectures, containers are treated as
// ephemeral (disposable) entities rather than persistent virtual machines.
// The standard operational pattern for restarts or updates is to completely
// destroy the existing container and create a new one from scratch, rather
// than pausing and resuming.
//
// Zero Configuration Drift: Destroying the container clears any leftover
//    temporary files, altered memory states, or crashed background processes.
//    Starting fresh guarantees the environment is in a predictable, clean state.
// Statelessness Enforcement: Forcing complete destruction ensures the
//    application does not mistakenly rely on the container's internal filesystem
//    for persistent storage.
// Industry Standard: This coupling of Stop/Remove followed by Create/Start
//    matches the behavior of production platforms (eg, Heroku, Vercel) when
//    a deployment is restarted.
// ---

// CreateAndStartNginxContainer pulls the nginx:alpine image IF not already present (handled by Docker daemon),
// creates a new Nginx container with the given static files bind-mounted,
// attaches Traefik routing labels, connects it to the Traefik network,
// and starts it.
// This is the function that makes a deployment "live": once this returns without error,
// Traefik has already picked up the new routing rule and the site is reachable.
// args: need context cuz the underlying docker sdk methods require context, config struct is just to organize actual args
func (dockerClient *Client) CreateAndStartNginxContainer(context context.Context, config NginxContainerConfigArgs) error {
	// --- pull image (with helper func) ---

	// ImagePull returns a stream of JSON progress events (one line per layer).
	// the stream MUST be fully consumed and closed before calling ContainerCreate.
	// if the stream is not drained, the pull is not guaranteed to complete,
	// and ContainerCreate may fail with "image not found".
	createContainerError := dockerClient.pullImageIfNotPresent(context, nginxImage) // logs are in the helper func
	if createContainerError != nil {
		return fmt.Errorf("failed to pull nginx image: %w", createContainerError)
	}

	// --- build container config (args for `docker create` or docker-compose config equivalent) ---

	// container.Config describes the container's internal runtime properties like
	// which image to use, environment variables, exposed ports, etc.
	// this is the config for "inside the container" view.
	// Why a pointer variable with `*`, and not a direct struct value? why need to add `&` address?
	// Because the Docker SDK functions (ContainerCreate) expect a pointer to the config struct (only accepts *container.Config)
	// and using a pointer allows modifying the config in place if needed (not necessary here, but common in more complex scenarios).
	containerInternalConfig := &container.Config{
		Image: nginxImage,

		// Labels are key-value metadata attached to the container.
		// Traefik watches the Docker socket and reads these labels to
		// automatically configure routing rules. when this container starts,
		// Traefik picks up the labels and begins routing <slug>.localhost to it.
		// no Traefik config file reload is required. this is the "Netlify magic".
		Labels: traefikLabels(config.Slug), // helper func

		// Why not set a Cmd field here?
		// The 'nginx:alpine' image inherently knows how to start its own web server process.
		// Explicitly setting the 'Cmd' field in the SDK overrides this default behavior, which would break the Nginx startup.
		//
		// Build vs. Serve: This specific container acts solely as a static file server.
		//    User build commands (eg, 'npm run build') require heavy language runtimes
		//    and will be executed by separate, ephemeral "build containers" earlier in
		//    the deployment pipeline.
	}

	/*
		'container.HostConfig' defines host‑specific runtime settings that live
		outside the container (port bindings, volume mounts, CPU/memory limits,
		network mode, restart policy, privilege flags).

		This is intentionally separate from 'container.Config' to reflect the
		Docker Engine API design:
		- container.Config = portable "inside the container" settings (Image, Cmd, Entrypoint, Env, WorkingDir, etc)
		- container.HostConfig = host‑dependent "outside the container" settings

		The Docker CLI and docker‑compose.yaml hide this split in a single user‑friendly
		API, but they still internally split those into these two structs. The Go
		SDK exposes the real Engine API for clarity and precision in code.
	*/
	containerHostConfig := &container.HostConfig{
		// Mounts is the list of bind mounts for this container.
		// a bind mount maps a host directory (outside container) into a container path (inside).
		// Type: `mount.TypeBind` this is a host directory (outside), not a Docker-managed volume.
		// Source: the host directory containing the static files (written by the build pipeline).
		// Target: the path inside the container where Nginx looks for files to serve.
		// ReadOnly: true. the container only needs to read the files, not write them.
		//   locking it read-only is a defence-in-depth measure: even if Nginx/container were compromised,
		//   it could not modify the source files on the host.
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   config.HostSourceDirectory,
				Target:   "/usr/share/nginx/html",
				ReadOnly: true,
			},
		},

		// RestartPolicy controls what Docker does when the container exits.
		// "unless-stopped" means: restart automatically on crash or host reboot,
		// but do not restart if the container was explicitly stopped (eg, on delete).
		// this keeps deployments alive across VM reboots without requiring
		// any external process manager.
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// network.NetworkingConfig connects the container to a named Docker network at creation time.
	// connecting at creation (not after start) avoids a race condition where Traefik
	// discovers the container before it is on the network and tries to proxy to
	// an address it cannot reach. The container and Traefik must share the same Docker network for Traefik
	// to see the container's internal IP address and proxy traffic to it.
	// Race Conditions: Traefik reacts to container start events instantly. Defining the network at creation time (rather than after startup)
	// ensures the container is already on the correct network when Traefik attempts to route traffic to it.
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			config.TraefikNetwork: {},
		},
	}
	// Map Syntax: The 'EndpointsConfig' field uses a map[string]*EndpointSettings
	//    because Docker allows a container to join multiple networks at once.
	// Using 'networkName: {}' creates a map entry with the
	//    specified name and assigns default network settings. This is sufficient
	//    for standard proxying where Traefik only needs to reach the container IP.

	// Default Behavior: When the platform parameter is nil, the Docker daemon
	// automatically selects the image layer matching the host's native
	// architecture (eg, linux/amd64 for linux/amd64 and macos/arm for macos/arm).
	// In a swarm or cluster containing mixed hardware (eg, both Raspberry Pi ARM
	//nodes and Intel x86 nodes), specifying the platform ensures the scheduler
	//places the container on a node capable of running that specific binary.
	var platform *v1.Platform = nil

	// --- Creating the container (`docker create` equivalent) ---
	createResponse, createContainerError := dockerClient.sdk.ContainerCreate(
		context,
		containerInternalConfig,
		containerHostConfig,
		networkingConfig,
		platform,
		config.ContainerName,
	)
	if createContainerError != nil {
		return fmt.Errorf("failed to create nginx container %q: %w", config.ContainerName, createContainerError)
	}

	dockerClient.logger.Info("nginx container created",
		"container_id", createResponse.ID[:12], // first 12 chars is the conventional short ID
		"container_name", config.ContainerName,
		"slug", config.Slug,
	)

	// --- start container  (like `docker start`) ---
	// ContainerStart transitions the container from "created" to "running".
	// after this call returns without error, the container is live and
	// Traefik will begin routing requests to it.
	startError := dockerClient.sdk.ContainerStart(
		context,
		createResponse.ID,
		container.StartOptions{},
		// start options are default (empty) most of the time. only used for advanced features like checkpoint/restore
	)
	if startError != nil {
		return fmt.Errorf("failed to start nginx container %q: %w", config.ContainerName, startError)
	}

	dockerClient.logger.Info("nginx container started",
		"container_name", config.ContainerName,
		"slug", config.Slug,
		"url", "http://"+config.Slug+".localhost",
	)
	// Why not https here?
	// In a containerized environment, the connection between the reverse proxy (Traefik) and the
	// application server (Nginx) occurs over a private, isolated Docker network. Encrypting traffic
	// within this "trusted" zone is generally unnecessary and adds significant CPU overhead.
	// Standard practice dictates that the gateway handles the encryption, while the internal jump to
	// the specific container remains over http.
	// Development Simplicity: Modern browsers treat '.localhost' as a secure
	//    origin, allowing development to proceed over HTTP without the friction
	//    of managing local self-signed certificates.
	//
	// Separation of Concerns: The Nginx container is responsible for file
	//    delivery. Security layers, such as TLS/SSL, are delegated to the
	//    infrastructure layer (Traefik).

	return nil
}

// StopAndRemoveContainer stops and removes a container by name.
// used when a deployment is deleted or before a redeploy replaces the old container.
// the function can handle when a container do not exist. If no container with
// the given name is found, it returns nil (NOT an error), because the desired
// state (container gone) is already satisfied.
func (dockerClient *Client) StopAndRemoveContainer(context context.Context, containerName string) error {
	// ContainerList with a name filter finds containers by name.
	// filters.NewArgs (from docker) builds the filter argument the SDK expects.
	// "name" filter matches containers whose name contains the given string,
	// so "deploy-happy-dog" also matches "deploy-happy-dog-2" if it exists.
	// the exact match is verified below by checking the full name in the list result.

	// create a single filter criteria for the container name. `filters.Arg` returns a KeyValuePair struct.
	nameCriteria := filters.Arg("name", containerName)

	// initialize the Args collection with the specific criteria.
	// NewArgs is variadic and can accept multiple KeyValuePair arguments.
	listFilters := filters.NewArgs(nameCriteria)

	// this is like `docker ps` i guess
	containers, containerListError := dockerClient.sdk.ContainerList(
		context,
		container.ListOptions{
			All:     true, // include stopped containers, not just running ones
			Filters: listFilters,
		},
	)
	if containerListError != nil {
		return fmt.Errorf("failed to list containers to find %q: %w", containerName, containerListError)
	}

	// find the exact container by name.
	// Docker prefixes container names with "/" internally (eg, "/deploy-happy-dog-3f9a").
	// the comparison includes the prefix to avoid false partial matches.
	targetName := "/" + containerName
	var targetContainerID string // declaration

	// The initial Docker API filter ("name=...") returns partial matches.
	// This nested loop ensures an exact match and stops searching immediately
	// once the correct container is identified.
	for _, listedContainer := range containers { // `_,` syntax is to discard the index
		// A single Docker container can have an array of multiple names (cuz legacy & alias stuff).
		// Iterate through all names attached to this specific container.
		for _, name := range listedContainer.Names {
			if name == targetName {
				targetContainerID = listedContainer.ID // found
				break
			}
		}
		// if something is already found, no need to continue checking
		if targetContainerID != "" {
			break
		}
	}
	/* Python Equivalent for this nested for-loop

	for listed_container in containers:
	    for name in listed_container['Names']:  # Assuming Names is a list in a dictionary
	        if name == target_name:
	            target_container_id = listed_container['ID']
	            break
	    if target_container_id:
	        break
	*/

	// container not found = desired state is already achieved
	if targetContainerID == "" {
		dockerClient.logger.Info("container not found, nothing to remove", "name", containerName)
		return nil
	}

	// --- Stop container ---
	// ContainerStop sends SIGTERM to the container process, giving it time to shut down gracefully.
	// if it does not exit within the timeout, Docker sends SIGKILL.
	// 10 seconds is generous for an Nginx container serving static files.
	stopTimeout := 10
	stopError := dockerClient.sdk.ContainerStop(
		context,
		targetContainerID,
		container.StopOptions{
			Timeout: &stopTimeout,
		},
		// usually no options needed, but can specify a custom timeout, custom signal (SIGINT)
	)
	if stopError != nil {
		return fmt.Errorf("failed to stop container %q: %w", containerName, stopError)
	}

	// --- Remove container ---
	// ContainerRemove deletes the container and its writable layer.
	// RemoveVolumes: false. Do not remove any named volumes attached to the container.
	//   (there are none here, but this is the safe default).
	// Force: false. The container was already stopped above, force is usually not needed.
	removeError := dockerClient.sdk.ContainerRemove(
		context,
		targetContainerID,
		container.RemoveOptions{
			RemoveVolumes: false, //  (`-v` flag in `docker rm`)
			Force:         false, // (`-f` flag in `docker rm`),
		},
	)
	if removeError != nil {
		return fmt.Errorf("failed to remove container %q: %w", containerName, removeError)
	}

	dockerClient.logger.Info("container stopped and removed", "name", containerName)
	return nil
}

// pullImageIfNotPresent (helper function) pulls a Docker image if it is not already present in the local image cache.
// The check for whether to download or not is handled by the Docker daemon.
// the pull response is a stream of JSON progress lines that must be fully consumed.
// discarding the output with io.Discard avoids storing progress text in memory,
// since the caller only needs to know if the pull succeeded or failed, not the progress detail.
// in v2, TODO: this stream can be forwarded to the deployment log file for visibility.
func (dockerClient *Client) pullImageIfNotPresent(context context.Context, imageName string) error {
	dockerClient.logger.Info("pulling docker image", "image", imageName)

	/*
		Docker SDK client's `.ImagePull()` sends a request to the Docker daemon to download
		the specified image from a container registry (Docker Hub by default).

		Return Value: A stream (io.ReadCloser) containing JSON progress logs. The Go program must
		read that stream to completion and must close the stream when done, or the Daemon might hang or leak resources.
		- Reader: Allows the program to consume the download progress data (do something with it).
		- Closer: Must be explicitly closed after reading to release the
		underlying network connection and prevent resource leaks.

		`image.PullOptions{}` : Configuration struct for the pull operation.
		Empty here because no authentication, platform override, nor
		special behavior is required for this public image pull.

		If the pull fails (eg, network error, invalid image name), the
		error is captured in `pullError` and handled.
	*/

	// ImagePull returns a io.ReadCloser struct that streams pull progress as newline-delimited JSON.
	// the pull is not complete until the stream is fully read and closed.
	imagePullResponseStream, pullError := dockerClient.sdk.ImagePull(
		context,
		imageName,
		image.PullOptions{},
	)
	if pullError != nil {
		// error returned to the caller to deal with
		return fmt.Errorf("failed to initiate image pull for %q: %w", imageName, pullError)
	}
	defer imagePullResponseStream.Close() // deferred so that the ReaderCloser stream is closed when this func ends

	/*
		The Docker Daemon sends image pull progress as a live stream of text? bytes?.
		The stream must be drained, or the daemon can block when
		the pipe buffer becomes full.

		An easy way to deal with it is just dumping it to standard output (aka dumping it to terminal)
		using `io.Copy(os.Stdout, imagePullResponseReader)` where
		`io.Copy` reads the entire stream and writes it to stdout (prints it to the console)
		Any io.Writer can be used instead like writing to file or logs (file, logger, MultiWriter).

		The stream must be closed after reading to release the
		underlying file descriptor (basically like ID by the OS for resources being used) and avoid resource leaks.
	*/
	// Here, io.Copy drains the stream into io.Discard (a write sink that discards all bytes).
	// this is required to cuz without draining, the HTTP response body is not fully consumed and the
	// Docker daemon may not finish writing all image layers to disk.
	_, err := io.Copy(io.Discard, imagePullResponseStream)
	if err != nil {
		return fmt.Errorf("failed to stream image pull response for %q: %w", imageName, err)
	}

	dockerClient.logger.Info("docker image pulled/downloaded and ready", "image", imageName)
	return nil
}

// traefikLabels returns the Docker container labels that instruct Traefik
// to route HTTP traffic for a specific slug to this container.
// Traefik watches the Docker socket and reacts to label changes in real time.
// no config file reload is needed. labels are the config.
//
// label breakdown:
//   - traefik.enable=true                      -- opt this container into Traefik routing
//     (required because exposedByDefault: false in traefik.yml)
//   - traefik.http.routers.<slug>.rule          -- match requests where the Host header equals <slug>.localhost
//   - traefik.http.services.<slug>.loadbalancer -- tell Traefik which port inside the container to proxy to
func traefikLabels(slug string) map[string]string {
	return map[string]string{
		"traefik.enable":                                              "true",
		"traefik.http.routers." + slug + ".rule":                      "Host(`" + slug + ".localhost`)",
		"traefik.http.services." + slug + ".loadbalancer.server.port": "80",
	}
}

// containerAge is a helper used in log output to show how long a container has been running.
// time.Since(start) computes the duration from a Unix timestamp to now and rounds to seconds.
func containerAge(startedAt time.Time) string {
	return time.Since(startedAt).Round(time.Second).String()
}
