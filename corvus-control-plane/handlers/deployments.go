package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/util"
)

// DeploymentHandler holds the dependencies needed by all deployment endpoints.
// the database is needed for every operation. the logger provides request-scoped context.
type DeploymentHandler struct {
	database *db.Database
	logger   *slog.Logger
}

// NewDeploymentHandler constructs a DeploymentHandler with its required dependencies.
func NewDeploymentHandler(database *db.Database, logger *slog.Logger) *DeploymentHandler {
	return &DeploymentHandler{
		database: database,
		logger:   logger,
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

// GetDeployment handles GET /api/deployments/:id.
// Returns a single deployment by UUID, or 404 if not found.
// in REST api, `:` colon is for a placeholder variable rather than a string (dynamic route segment)
// the colon syntax is just for documentation and convention. In go chi lib, curly braces is used (`/deployments/{id}`)
func (handler *DeploymentHandler) GetDeployment(responseWriter http.ResponseWriter, request *http.Request) {

	// chi.URLParam() extracts the named URL parameter registered in the route pattern.
	// for the route "/api/deployments/{id}", chi.URLParam(r, "id") returns the value
	// that matched the {id} segment of the incoming URL.
	deploymentID := chi.URLParam(request, "id")
	/*
		When a request is made to an endpoint like /api/deployments/happy-dog-1234,
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
// decodes and validates the request body,
// generates a deployment ID and slug,
// persists the record,
// and returns the created deployment with status 201.
// the actual build and container pipeline is NOT triggered here yet (TODO Phase 3).
// this handler only creates the database record and returns it.
func (handler *DeploymentHandler) CreateDeployment(responseWriter http.ResponseWriter, request *http.Request) {

	// --- Decode request body ---

	// json.NewDecoder reads from the request body stream.
	// Decode() populates the target struct and returns an error if:
	//   - the body is not valid JSON
	//   - a field type does not match (e.g. a string where a bool is expected)
	//   - the body is empty
	var requestBody createDeploymentRequest
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "invalid JSON request body", handler.logger)
		return
	}
	// why not make a json decoder helper function like the write function?
	// cuz the function is just 3 lines and the error bubbling logic will be abstracted by 1 layer,
	// so it is just not worth it

	// --- Validate required fields ---
	// why is the validation not a helper function?
	// cuz it is too specific to this deployment fields that other tables/requests (if exists in the future)
	// wont get to use it anyways.
	// same reason for env vars, they are just too specific to a struct?

	// validation is intentionally explicit and readable rather than using a validation library.
	// for a small set of fields, direct checks are clearer and produce more precise error messages than tag-based validators.
	if requestBody.Name == "" {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "name is required", handler.logger)
		return
	}
	// not either of zip or github
	if requestBody.SourceType != models.SourceZip && requestBody.SourceType != models.SourceGitHub {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "source_type must be 'zip' or 'github'", handler.logger)
		return
	}
	// source github but no url
	if requestBody.SourceType == models.SourceGitHub && (requestBody.GitHubURL == nil || *requestBody.GitHubURL == "") {
		writeErrorJsonAndLogIt(responseWriter, http.StatusBadRequest, "github_url is required when source_type is 'github'", handler.logger)
		return
	}

	// --- apply defaults ---

	if requestBody.Branch == "" {
		requestBody.Branch = "main"
	}
	if requestBody.OutputDirectory == "" {
		requestBody.OutputDirectory = "."
	}

	// --- encode env vars ---

	// EnvironmentVariables is received as map[string]string from JSON.
	// the database stores it as a JSON string (SQLite has no map column type).
	// encoding happens here so the model and db layers deal only with *string,
	// not with maps or encoding logic.
	var encodedEnvVars *string
	if len(requestBody.EnvironmentVariables) > 0 {
		envVarsBytes, err := json.Marshal(requestBody.EnvironmentVariables)
		if err != nil {
			handler.logger.Error("failed to encode env vars", "error", err)
			writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to process env vars", handler.logger)
			return
		}

		encodedString := string(envVarsBytes) // convert raw bytes to string
		encodedEnvVars = &encodedString
	} // if no env vars its fine

	// --- generate deployment identifiers ---

	deploymentID := uuid.New().String()
	slug := util.GenerateSlug()

	// the webhook secret is a 32-byte cryptographically random value encoded as hex.
	// 32 bytes = 256 bits of entropy, which is the same strength as an HMAC-SHA256 key.
	// crypto/rand is used (not math/rand) because this value is a security credential:
	// it is used to verify GitHub webhook signatures. math/rand is not suitable for secrets.
	webhookSecret, err := generateWebhookSecret() // helper function
	if err != nil {
		handler.logger.Error("failed to generate webhook secret", "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to generate deployment credentials", handler.logger)
		return
	}

	// --- create/build deployment URL ---

	// the URL is constructed from the slug and set immediately so the client
	// knows the public address before the container is even started.
	// the container may not be live yet (status is "deploying") but the URL is deterministic.
	deploymentURL := "http://" + slug + ".localhost"

	// assemble the deployment model to put into database
	deployment := &models.Deployment{
		ID:                   deploymentID,
		Slug:                 slug,
		Name:                 requestBody.Name,
		SourceType:           requestBody.SourceType,
		GitHubURL:            requestBody.GitHubURL,
		Branch:               requestBody.Branch,
		BuildCommand:         requestBody.BuildCommand,
		OutputDirectory:      requestBody.OutputDirectory,
		EnvironmentVariables: encodedEnvVars,
		Status:               models.StatusDeploying,
		URL:                  &deploymentURL,
		WebhookSecret:        &webhookSecret,
		AutoDeploy:           requestBody.AutoDeploy,
	}

	// --- Write to database (persist to database)  ---
	err = handler.database.InsertDeployment(deployment)
	if err != nil {
		handler.logger.Error("failed to insert deployment", "id", deploymentID, "error", err)
		writeErrorJsonAndLogIt(responseWriter, http.StatusInternalServerError, "failed to create deployment", handler.logger)
		return
	}

	// info logging
	handler.logger.Info("deployment created",
		"id", deploymentID,
		"slug", slug,
		"source_type", requestBody.SourceType,
		"name", requestBody.Name,
	)

	// 201 Created is the correct status for a successful resource creation.
	// 200 OK is for successful reads or updates, not for new resource creation.
	// this responds with 201 AND a json of the full models.Deployment struct that just created
	// (for the client to do frontend logic and display)
	writeJsonAndRespond(responseWriter, http.StatusCreated, deployment)
}

// generateWebhookSecret returns a cryptographically secure random hex string
// suitable for use as an HMAC-SHA256 signing secret.
// 32 random bytes encoded as hex produces a 64-character string.
func generateWebhookSecret() (string, error) {
	// make a 32-byte slice, crypto/rand fills it with random bytes from the OS entropy source
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(secretBytes), nil
}
