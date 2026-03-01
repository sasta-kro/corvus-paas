// Package models defines the data structures (structs) shared across the application.
// this package has no imports from other internal packages, making it the
// foundation of the dependency graph. other packages (db, handlers, build) import from model.go
package models

import "time"

/*
DeploymentStatus and SourceType are both string under the hood, but giving them their own type
means the Go compiler will reject `deployment.Status = "typo"` if "typo" is not one of
the declared constants. Plain string fields give no such protection (type safety)
*/

// DeploymentStatus represents the current lifecycle state of a deployment.
// using a named string type instead of plain string enforces that only valid
// status values are used at compile time when combined with the constants below.
type DeploymentStatus string

// SourceType represents where the deployment's source files originate from.
type SourceType string

/*
Gemini said
In Go, the const (...) syntax is a Constant Block. It is used like an ENUM in other languages.
Used like `models.StatusDeploying` or `models.SourceZip` where the `models` is the package name.
It is used to group related constants together under a single declaration,
rather than repeating the const keyword on every line.
basically instead of this:
```
const StatusDeploying DeploymentStatus = "deploying"
const StatusLive DeploymentStatus = "live"
const StatusFailed DeploymentStatus = "failed"
```
it can be written as:
*/
const (
	// StatusDeploying means the pipeline is actively running (cloning, building, starting container)
	StatusDeploying DeploymentStatus = "deploying"

	// StatusLive means the container is running and the site is reachable
	StatusLive DeploymentStatus = "live"

	// StatusFailed means the pipeline encountered an error and did not complete
	StatusFailed DeploymentStatus = "failed"
)

const (
	// SourceZip means the deployment source is a user-uploaded zip file containing the static site files
	SourceZip SourceType = "zip"

	// SourceGitHub means the deployment source is a public GitHub repository URL
	SourceGitHub SourceType = "github"
)

/*
Deployment is the central data model for the application.
it maps 1:1 to the deployments table in SQLite and is the struct
passed between the database layer, the pipeline, and the HTTP handlers.

`json` struct tags control how the Go struct is serialized/converted to JSON in HTTP responses.
`db` struct tags are used by the database layer for column name mapping from the Go struct field name.
`omitempty` on pointer fields means the key is omitted from JSON output when the value is nil,
which keeps API responses clean for fields that are not always populated.
*/
type Deployment struct {
	/*
		Pointer fields (*string) are for optional data.
		GitHubURL is `*string` not `string` because a zip deployment has no GitHub URL at all.
		Go do not allow nil values for plain strings, so Go forces to store an empty string.
		Empty string shouldn't be the default cuz it should be "main" or something
	*/

	// ID is a UUID v4, generated at creation time, used as the primary key
	ID string `json:"id" db:"id"`

	// Slug is the URL-safe identifier used in the deployment's public URL.
	// example: "graceful-fox" -> http://graceful-fox.localhost
	Slug string `json:"slug" db:"slug"`

	// Name is the human-readable label the user assigns to the deployment
	Name string `json:"name" db:"name"`

	// SourceType is either "zip" or "github"
	SourceType SourceType `json:"source_type" db:"source_type"`

	// GitHubURL is the public repo URL, only populated for github source type
	// why POINTER? it allows the field to be nil unlike a value string, which defaults to "" (empty string)
	GitHubURL *string `json:"github_url,omitempty" db:"github_url"`

	// Branch is the git branch to clone and build from, should default to "main"
	Branch string `json:"branch" db:"branch"`

	// BuildCommand is the shell command run inside the build container before serving.
	// empty string means no build step (pre-built static site).
	// example: "npm ci && npm run build"
	BuildCommand string `json:"build_command" db:"build_command"`

	// OutputDirectory is the directory inside the repo or extracted zip that contains
	// the final static files to serve. defaults to "." (root of the archive).
	// example: "dist", "build", "out"
	OutputDirectory string `json:"output_directory" db:"output_directory"`

	// EnvironmentVariables is a JSON-encoded key-value map of environment variables
	// passed into the build container. stored as a string in SQLite.
	// example: {"NODE_ENV":"production"}
	// nil means no env vars were provided
	EnvironmentVariables *string `json:"environment_variables,omitempty" db:"environment_variables"`

	// Status is the current lifecycle state of the deployment
	Status DeploymentStatus `json:"status" db:"status"`

	// URL is the fully qualified URL where the deployment is reachable.
	// set after the container starts successfully.
	// example: "http://graceful-fox.localhost"
	// slug is the "nickname" part of the URL, but URL is the full address including protocol and domain
	URL *string `json:"url,omitempty" db:"url"`

	// WebhookSecret is the HMAC-SHA256 signing secret for GitHub webhook verification.
	// generated at deployment creation time, returned once to the user, never logged
	WebhookSecret *string `json:"webhook_secret,omitempty" db:"webhook_secret"`

	// AutoDeploy controls whether a push to the configured branch triggers a rebuild.
	// stored as INTEGER 0/1 in SQLite (SQLite has no native boolean type).
	AutoDeploy bool `json:"auto_deploy" db:"auto_deploy"`

	// ExpiresAt is the timestamp when this deployment should be automatically
	// cleaned up (container stopped, files removed, DB row deleted).
	// nil means the deployment does not expire.
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`

	// CreatedAt is set once at row insertion time
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is refreshed on every status transition
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
