package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/build"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/util"
)

// DeploymentHandler holds the dependencies needed by all deployment endpoints.
// the database is needed for every operation. the logger provides request-scoped context.
type DeploymentHandler struct {
	database         *db.Database
	logger           *slog.Logger
	deployerPipeline *build.DeployerPipeline
}

// NewDeploymentHandler constructs a DeploymentHandler with its required dependencies.
func NewDeploymentHandler(
	database *db.Database,
	logger *slog.Logger,
	deployerPipeline *build.DeployerPipeline,
) *DeploymentHandler {

	return &DeploymentHandler{
		database:         database,
		logger:           logger,
		deployerPipeline: deployerPipeline,
	}
}

// createDeploymentRequest defines the shape of the JSON body accepted by POST /api/deployments.
// (client wants to create a new app deployment)
// This is different from the models.Deployment struct, which represents the full deployment record stored in the database.
// fields are validated after decoding (there's validation block inside .CreateDeployment() method)
// All fields use pointer types for optional values so that a missing field
// (absent from JSON) is distinguishable from an explicit empty string "".
// Required fields use plain value types so a missing field decodes to the zero value,
// which the validation block catches and rejects.
type createDeploymentRequest struct {
	// Name is the human-readable label for this deployment (required)
	Name string `json:"name"`

	// SourceType must be "zip" or "github" (required)
	SourceType models.SourceType `json:"source_type"`

	// GitHubURL is the public GitHub repo URL, required only when source_type is "github"
	// `omitempty` will set <null> in the database when field is missing
	GitHubURL *string `json:"github_url,omitempty"`

	// Branch is the git branch to deploy from, defaults to "main" when omitted
	// doesn't matter for zip source type. easier to just make it required than have pointers
	Branch string `json:"branch"`

	// BuildCommand is the shell command to run inside the build container before serving.
	// empty string means no build step (pre-built static site, or a raw zip with no build).
	BuildCommand string `json:"build_command"`

	// OutputDirectory is the subdirectory containing the final static files.
	// defaults to "." (root of the archive or repo).
	OutputDirectory string `json:"output_directory"`

	// EnvironmentVariables is an optional map of environment variables passed to the build container.
	// stored as a JSON string in SQLite. nil means no env vars.
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`

	// AutoDeploy enables automatic redeployment on GitHub push when true.
	// Only relevant for github source type.
	AutoDeploy bool `json:"auto_deploy"`
}

// ListDeployments method handles GET /api/deployments.
// returns all deployments as a JSON array, newest first.
// returns an empty JSON array [] (not null) when no deployments exist,
// because null is harder for frontend clients to handle than an empty array.
func (handler *DeploymentHandler) ListDeployments(responseWriter http.ResponseWriter, request *http.Request) {
	deployments, err := handler.database.ListDeployments()
	if err != nil {
		// using the logger passed in to the handler (not a global logger)
		handler.logger.Error("failed to list deployments", "error", err)
		writeErrorJsonAndLogIt(
			responseWriter,
			http.StatusInternalServerError,
			"failed to retrieve deployments",
			handler.logger,
		)
		return
	}

	// ListDeployments returns nil when the table is empty (append() on a nil slice stays nil).
	// json.Marshal encodes nil slices as JSON null, not [].
	// explicitly converting nil to an empty slice/array ensures the API always returns [].
	if deployments == nil {
		deployments = []*models.Deployment{} // empty list/array of deployments
	}

	writeJsonAndRespond(responseWriter, http.StatusOK, deployments)
}

// GetDeployment handles GET /api/deployments/:uuid.
// Returns a single deployment by UUID, or 404 if not found.
// in REST api, `:` colon is for a placeholder variable rather than a string (dynamic route segment)
// the colon syntax is just for documentation and convention. In go chi lib, curly braces is used (`/deployments/{id}`)
// TODO: endpoint/resource aliasing for `/deployments/{uuid}` with `/deployments/{slug}` so it has better UX (just uuid is just ugly)
func (handler *DeploymentHandler) GetDeployment(responseWriter http.ResponseWriter, request *http.Request) {

	// chi.URLParam() extracts the named URL parameter registered in the route pattern.
	// for the route "/api/deployments/{id}", chi.URLParam(r, "id") returns the value
	// that matched the {id} segment of the incoming URL.
	deploymentID := chi.URLParam(request, "uuid")
	/*
		When a request is made to an endpoint like /api/deployments/happy-dog-1234, (or rn, with uuid)
		the Chi router matches the "happy-dog-1234" segment against the {id} placeholder.
		Calling chi.URLParam(request, "id") inspects the routing context embedded within
		the http.Request object and retrieves that specific matched value. No need for
		manual URL parsing.
	*/

	deployment, err := handler.database.GetDeployment(deploymentID)
	// 2 error checks, for 'record not found' and for actual error
	if errors.Is(err, db.ErrRecordNotFound) {
		writeErrorJsonAndLogIt(responseWriter, http.StatusNotFound, "deployment not found", handler.logger)
		return
	}
	if err != nil {
		handler.logger.Error("failed to get deployment", "id", deploymentID, "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to retrieve deployment", handler.logger)
		return
	}

	writeJsonAndRespond(responseWriter, http.StatusOK, deployment)
}

// CreateDeployment handles POST /api/deployments.
// for source_type "zip" - reads a multipart form upload, validates fields,
// creates the database record, and fires the deployerPipeline in a goroutine.
// for source_type "github": calls deploy github pipeline
// returns 201 immediately with status "deploying".
// the client polls GET /api/deployments/:id to track progress to "live" or "failed".

func (handler *DeploymentHandler) CreateDeployment(responseWriter http.ResponseWriter, request *http.Request) {
	// ===== parse multipart form (needed for the zip upload and json in 1 http request)

	// ParseMultipartForm reads the incoming multipart body into memory up to maxMemory bytes.
	// files larger than maxMemory are spilled to disk automatically by the standard library.
	// 32MB is a reasonable upper limit for a zip containing a pre-built static site.
	// large files (video, raw assets) should be excluded from the zip before uploading.
	const maxMemoryBytes = 32 << 20 // 32MB  TODO maybe handle this better
	errParseMultipartForm := request.ParseMultipartForm(maxMemoryBytes)
	if errParseMultipartForm != nil {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "failed to parse multipart form", handler.logger)
		return
	}

	// ===== Read form fields and validate step by step

	var validatedRequest createDeploymentRequest
	// request.FormValue reads a named field from the parsed multipart form.
	// returns an empty string if the field is absent.

	name := request.FormValue("name")
	if name == "" {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "name is required", handler.logger)
		return
	}
	validatedRequest.Name = name

	rawSourceType := request.FormValue("source_type")
	sourceType := models.SourceType(rawSourceType) // cast type
	if sourceType != models.SourceZip && sourceType != models.SourceGitHub {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "source_type must be 'zip' or 'github'", handler.logger)
		return
	}
	validatedRequest.SourceType = models.SourceType(rawSourceType)

	// only populate the pointer if source is github
	// a nil pointer means "not provided", an empty string means "provided but blank".
	var githubURL *string
	if sourceType == models.SourceGitHub {
		rawGitHubURL := request.FormValue("github_url")
		if rawGitHubURL == "" {
			writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "github_url is required when source_type is 'github'", handler.logger)
			return
		}
		githubURL = &rawGitHubURL
	}
	validatedRequest.GitHubURL = githubURL // empty pointer will get passed if not source github

	branch := request.FormValue("branch")
	if branch == "" {
		branch = "main"
	}
	validatedRequest.Branch = branch

	buildCommand := request.FormValue("build_command") // idk if i can even properly validate build commands
	validatedRequest.BuildCommand = buildCommand

	outputDirectory := request.FormValue("output_directory")
	if outputDirectory == "" {
		outputDirectory = "."
	}
	validatedRequest.OutputDirectory = outputDirectory

	rawEnvironmentVariables := request.FormValue("environment_variables")
	// env vars arrive as a JSON string in the form field.
	// decoding it into a map, then re-encode it as a JSON string for storage.
	// this round-trip validates the JSON and normalises the format.
	// TODO the env var doesn't really need to be in the createDeploymentRequest struct so i gotta do something with it.
	var encodedEnvironmentVariables *string
	if rawEnvironmentVariables != "" {
		var envVarsMap map[string]string
		err := json.Unmarshal([]byte(rawEnvironmentVariables), &envVarsMap)
		if err != nil {
			writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "environment_variables must be a valid JSON object", handler.logger)
			return
		}
		if len(envVarsMap) > 0 {
			envBytes, errMarshal := json.Marshal(envVarsMap)
			if errMarshal != nil {
				handler.logger.Error("failed to encode env vars", "error", errMarshal)
				writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to process environment variables", handler.logger)
				return
			}
			encoded := string(envBytes)
			encodedEnvironmentVariables = &encoded
		}
	}

	// form values are always strings. "true" -> true, anything else -> false.
	rawAutoDeploy := request.FormValue("auto_deploy")
	autoDeploy := rawAutoDeploy == "true"
	validatedRequest.AutoDeploy = autoDeploy

	// ===== handle zip file upload

	// FormFile retrieves the uploaded file for the given form field name.
	// returns the file content as a multipart.File (implements io.Reader)
	// and a FileHeader containing metadata (original filename, size, MIME type).
	// returns an error if the field is absent or the upload failed.
	var uploadedFile multipart.File // multipart.File is an interface that embeds io.Reader, io.ReaderAt, io.Seeker, and io.Closer
	if validatedRequest.SourceType == models.SourceZip {
		var formFileErr error
		uploadedFile, _, formFileErr = request.FormFile("file")
		if formFileErr != nil {
			writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "file is required for zip source type", handler.logger)
			return
		}
		// the uploadedFile is not closed here. If deferred to close, it will close when this CreateDeployment
		// function ends. But the DeployerPipeline is in a go routine so
		// it wont finish yet, thus having closed file errors

	}

	// ===== generate deployment identifiers

	deploymentID := uuid.New().String()
	slug := util.GenerateSlug()

	// the webhook secret is a 32-byte cryptographically random value encoded as hex.
	// 32 bytes = 256 bits of entropy, which is the same strength as an HMAC-SHA256 key.
	// crypto/rand is used (not math/rand) because this value is a security credential.
	// it is used to verify GitHub webhook signatures. math/rand is not suitable for secrets.
	webhookSecret, err := generateWebhookSecret() // helper function
	if err != nil {
		handler.logger.Error("failed to generate webhook secret", "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to generate deployment credentials", handler.logger)
		return
	}

	// ===== create/build deployment URL

	// the URL is constructed from the slug and set immediately so the client
	// knows the public address before the container is even started.
	// the container may not be live yet (status is "deploying") but the URL is deterministic.
	deploymentURL := "http://" + slug + ".localhost"

	// assemble the deployment model to put into database
	deployment := &models.Deployment{
		ID:                   deploymentID,
		Slug:                 slug,
		Name:                 validatedRequest.Name,
		SourceType:           validatedRequest.SourceType,
		GitHubURL:            validatedRequest.GitHubURL,
		Branch:               validatedRequest.Branch,
		BuildCommand:         validatedRequest.BuildCommand,
		OutputDirectory:      validatedRequest.OutputDirectory,
		EnvironmentVariables: encodedEnvironmentVariables,
		Status:               models.StatusDeploying,
		URL:                  &deploymentURL,
		WebhookSecret:        &webhookSecret,
		AutoDeploy:           validatedRequest.AutoDeploy,
	}

	// ===== Writing to database (persist to database)
	err = handler.database.InsertDeployment(deployment)
	if err != nil {
		handler.logger.Error("failed to insert deployment", "id", deploymentID, "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to create deployment", handler.logger)
		return
	}

	// info logging
	handler.logger.Info("deployment record created",
		"id", deploymentID,
		"slug", slug,
		"source_type", validatedRequest.SourceType,
		"name", validatedRequest.Name,
	)

	// =====  Start deployerPipeline asynchronously

	// the deployerPipeline runs in a goroutine so the HTTP handler returns immediately.
	// the client receives 201 with status "deploying" and polls for updates.
	// the goroutine captures the deployerPipeline pointer and deployment by value (safe since deployment is a pointer).
	if validatedRequest.SourceType == models.SourceZip && uploadedFile != nil {
		go handler.deployerPipeline.DeployZipUpload(deployment, uploadedFile)
	}

	if validatedRequest.SourceType == models.SourceGitHub {
		go handler.deployerPipeline.DeployGitHub(deployment)
	}

	// 201 Created is the correct status for a successful resource creation
	// (the record exists, the deployerPipeline is running.)
	// 200 OK is for successful reads or updates, not for new resource creation.
	writeJsonAndRespond(responseWriter, http.StatusCreated, deployment)
	// (response is for the client to do frontend logic and display)
}

// DeleteDeployment handles DELETE /api/deployments/:uuid.
// performs the full teardown sequence:
//   - fetch the deployment record from the database
//   - stop and remove the Docker container
//   - remove the static files from the asset storage root
//   - remove the deployment log file
//   - delete the database record (last, so retries are possible if earlier steps fail)
//
// returns 204 No Content on success (the standard HTTP response for a successful delete
// that has no response body to return).
func (handler *DeploymentHandler) DeleteDeployment(responseWriter http.ResponseWriter, request *http.Request) {
	deploymentID := chi.URLParam(request, "uuid")

	// fetch the deployment to get the slug (needed for container name and file paths)
	deployment, err := handler.database.GetDeployment(deploymentID)
	if errors.Is(err, db.ErrRecordNotFound) {
		writeErrorJsonAndLogIt(responseWriter, http.StatusNotFound, "deployment not found", handler.logger)
		return
	}
	if err != nil {
		handler.logger.Error("failed to get deployment for delete", "id", deploymentID, "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to retrieve deployment", handler.logger)
		return
	}

	handler.logger.Info("deleting deployment",
		"id", deploymentID,
		"slug", deployment.Slug,
	)

	// ===== stop and remove the container.
	// StopAndRemoveContainer is idempotent, returns nil if the container does not exist.
	// this handles the case where the deployment failed before a container was started,
	// or where the container was already manually removed.
	containerName := "deploy-" + deployment.Slug
	context := request.Context()

	err = handler.deployerPipeline.CleanupContainer(context, containerName)
	if err != nil {
		handler.logger.Error("failed to remove container during delete",
			"id", deploymentID,
			"container", containerName,
			"error", err,
		)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to remove deployment container", handler.logger)
		return
	}

	// ===== remove the static files from the asset storage root.
	// os.RemoveAll is idempotent, returns nil if the path does not exist.
	err = handler.deployerPipeline.CleanupFiles(deployment.Slug)
	if err != nil {
		handler.logger.Error("failed to remove deployment files during delete",
			"id", deploymentID,
			"slug", deployment.Slug,
			"error", err,
		)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to remove deployment files", handler.logger)
		return
	}

	// ===== remove the log file (associated with the container)
	// non-fatal if this fails, the deployment is already torn down,
	// a leftover log file is not a functional issue.
	if err := handler.deployerPipeline.CleanupLogFile(deployment.Slug); err != nil {
		handler.logger.Warn("failed to remove deployment log file (non-fatal)",
			"slug", deployment.Slug,
			"error", err,
		)
	}

	// ===== delete the database record (last)
	// if any previous step failed and returned early, the record still exists,
	// allowing the user to retry the delete request.
	if err := handler.database.DeleteDeployment(deploymentID); err != nil {
		handler.logger.Error("failed to delete deployment record",
			"id", deploymentID,
			"error", err,
		)
		writeErrorJsonAndLogIt(
			responseWriter,
			http.StatusInternalServerError,
			"failed to delete deployment record", handler.logger,
		)
		return
	}

	handler.logger.Info("deployment deleted",
		"id", deploymentID,
		"slug", deployment.Slug,
	)

	// 204 No Content = the resource was successfully deleted, there is nothing to return.
	// WriteHeader without Write sends an empty response body.
	responseWriter.WriteHeader(http.StatusNoContent)
}

// RedeployDeployment handles POST /api/deployments/:uuid/redeploy.
// fetches the existing deployment, validates it can be redeployed,
// and triggers the appropriate pipeline in a goroutine.
// returns 202 Accepted immediately (the redeploy runs asynchronously).
// the client polls GET /api/deployments/:uuid to track the status transition
// from "deploying" back to "live" or "failed".
func (handler *DeploymentHandler) RedeployDeployment(responseWriter http.ResponseWriter, request *http.Request) {
	deploymentID := chi.URLParam(request, "uuid")

	deployment, err := handler.database.GetDeployment(deploymentID)
	if errors.Is(err, db.ErrRecordNotFound) {
		writeErrorJsonAndLogIt(responseWriter, http.StatusNotFound, "deployment not found", handler.logger)
		return
	}
	if err != nil {
		handler.logger.Error("failed to get deployment for redeploy", "id", deploymentID, "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to retrieve deployment", handler.logger)
		return
	}

	handler.logger.Info("redeploy requested",
		"id", deploymentID,
		"slug", deployment.Slug,
		"source_type", deployment.SourceType,
	)

	// trigger the pipeline based on source type.
	// zip redeployments use the existing files on disk.
	// GitHub redeployments will re-clone and rebuild
	switch deployment.SourceType {
	case models.SourceZip:
		go handler.deployerPipeline.RedeployExistingZip(deployment)
	case models.SourceGitHub:
		go handler.deployerPipeline.DeployGitHub(deployment)

	default:
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "unknown source type", handler.logger)
		return
	}

	// 202 Accepted: the request has been accepted for processing, but the processing
	// is not complete. the client should poll the deployment status.
	// return the deployment object so the client has the ID and slug to poll with.
	writeJsonAndRespond(responseWriter, http.StatusAccepted, deployment)
}
