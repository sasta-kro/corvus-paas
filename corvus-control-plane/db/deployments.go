package db

// deployments.go contains all SQL query functions for the deployments table.
// each function is a method on *Database and operates on a single table - deployments (for now)
// raw SQL syntax is used intentionally as it keeps the query layer explicit,
// avoids ORM magic, and makes the SQL readable and auditable.

// Raw SQL vs. ORM (Object-Relational Mapping)?
// Raw SQL: Involves writing explicit SQL statements (SELECT, INSERT,
//    etc.) as strings within the Go code. This provides maximum
//    transparency and allows for precise performance optimization
//    since the developer controls every instruction sent to the database.
//
// ORM Magic: Refers to libraries that automatically generate SQL
//    based on code objects (structs). While faster to write initially,
//    it creates an abstraction layer that can hide bugs, lead to
//    inefficient database access patterns, and complicate debugging.
//
// By using raw database/sql, the CorvusPaas
//    logic remains "auditable." A developer can read the file and
//    immediately understand the database schema and interaction
//    without needing to know the specific internal rules of an ORM library.

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// ErrRecordNotFound is returned by GetDeployment when no row matches the given ID.
// callers should check for this sentinel error to distinguish "not found" (404)
// from a real database error (500, internal server error).
var ErrRecordNotFound = errors.New("deployment not found")

// InsertDeployment writes a new deployment row to the database.
// the deployment struct MUST have ID, Slug, and Status already populated
// by the caller (handler or pipeline) before calling this function.
// > Why not put this function in the db.go file?
// Because db.go is meant for general database connection and migration (DB DDL init) logic,
// while deployments.go is specifically for the deployments table/relation and its functions.
// If another table is added like "users",
// then a new file users.go would be created with functions related to the users table.
// > Why pointer to *models.Deployment instead of value models.Deployment?
// Accepting *models.Deployment prevents unnecessary memory allocation (copying a large struct)
// and allows the function to mutate the original struct (eg, setting CreatedAt timestamps) so the
// caller retains the updated data.
func (database *Database) InsertDeployment(deployment *models.Deployment) error {
	// backticks ` are equivalent to """ multi line strings in python.
	// Parameterized Queries (?): The question marks act as secure placeholders
	//    for values. The database driver binds variables to these placeholders
	//    at execution time. This strictly separates the SQL command from the
	//    user-provided data, completely eliminating the risk of SQL injection.
	query := ` 
		INSERT INTO deployments (
			id, slug, name,
			source_type, github_url, branch,
			build_cmd, output_dir, env_vars, 
			status, url, webhook_secret, 
			auto_deploy, created_at, updated_at
		) VALUES (
			?, ?, ?, -- these are parameter placeholders, PostgresSQL uses $1, $2, $3
			?, ?, ?, 
			?, ?, ?, 
			?, ?, ?,
			?, ?, ?
		)
	`

	timeNow := time.Now().UTC()
	// .Now() returns the time from the computer's system clock.
	// Storing timestamps in UTC prevents "time drift" issues when the application server
	// and the database server are in different geographic regions with different timezones.

	// Setting CreatedAt and UpdatedAt inside the InsertDeployment function ensures that the caller is
	// not responsible for managing record metadata, leading to a cleaner API and consistent
	// timestamping across all database entries.
	deployment.CreatedAt = timeNow
	deployment.UpdatedAt = timeNow

	_, err := database.connection.Exec(query, // takes query and args... for placeholder parameters
		deployment.ID,
		deployment.Slug,
		deployment.Name,
		deployment.SourceType,
		deployment.GitHubURL, // *string, nil inserts NULL
		deployment.Branch,
		deployment.BuildCmd,
		deployment.OutputDir,
		deployment.EnvVars, // *string, nil inserts NULL
		deployment.Status,
		deployment.URL,           // *string, nil inserts NULL
		deployment.WebhookSecret, // *string, nil inserts NULL
		deployment.AutoDeploy,    // bool, driver converts to 0/1
		deployment.CreatedAt,
		deployment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert deployment %q: %w", deployment.ID, err)
	}
	return nil
}

// GetDeployment fetches a single deployment row by its UUID.
// returns ErrRecordNotFound if no row matches, which callers map to HTTP 404.
func (database *Database) GetDeployment(id string) (*models.Deployment, error) {
	query := `
		SELECT
			id, slug, name, source_type, github_url, branch,
			build_cmd, output_dir, env_vars, status, url,
			webhook_secret, auto_deploy, created_at, updated_at
		FROM deployments
		WHERE id = ?
	`

	// QueryRow is used for single-row queries. (Query() is for multiple rows.)
	// it returns a *sql.Row which has a Scan() method to read the data.
	row := database.connection.QueryRow(query, id)
	// QueryRow defers the "not found" check until Scan is invoked. If the database returns
	// an empty set, Scan returns sql.ErrNoRows, which is then mapped to the domain-specific sentinel error.

	deployment, err := scanDeployment(row)
	if errors.Is(err, sql.ErrNoRows) {
		// sql.ErrNoRows is the standard error returned by Scan() when no row matches the query.
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %q: %w", id, err)
	}
	return deployment, nil
}

// ListDeployments returns all deployment rows ordered by creation time descending (DESC)
// (newest first), matching the expected dashboard sort order.
func (database *Database) ListDeployments() ([]*models.Deployment, error) {
	query := `
		SELECT
			id, slug, name, source_type, github_url, branch,
			build_cmd, output_dir, env_vars, status, url,
			webhook_secret, auto_deploy, created_at, updated_at
		FROM deployments
		ORDER BY created_at DESC
	`

	rows, err := database.connection.Query(query) // Query() returns multiple rows
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// rows.Close() releases the database connection back to the pool.
	// must be deferred immediately after checking the Query error,
	// not before, because rows is nil if err != nil.
	// Automatic Cleanup (QueryRow): The *sql.Row type returned by QueryRow
	//    automatically releases its database connection back to the pool
	//    as soon as the Scan() method is called.
	//
	// Manual Cleanup (Query): The *sql.Rows type returned by Query
	//    maintains an active connection to allow for iteration. This
	//    connection must be explicitly released via rows.Close().
	//    Failure to close rows results in a "connection leak," eventually
	//    exhausting the connection pool and causing the application to hang.
	defer rows.Close()

	// TODO: continue checking the code from here.

	var deployments []*models.Deployment
	for rows.Next() {
		deployment, err := scanDeployment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}
		deployments = append(deployments, deployment)
	}

	// rows.Err() returns any error that occurred during iteration.
	// this is separate from the scan error above and must be checked
	// after the loop, not inside it.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment rows: %w", err)
	}

	return deployments, nil
}

// UpdateStatus sets the status and updated_at timestamp for a deployment.
// this is the most frequent write operation: called at each state transition
// in the deployment pipeline (deploying -> live | failed).
func (database *Database) UpdateStatus(id string, status models.DeploymentStatus) error {
	query := `UPDATE deployments SET status = ?, updated_at = ? WHERE id = ?`

	result, err := database.conn.Exec(query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update status for deployment %q: %w", id, err)
	}

	// RowsAffected returns 0 if no row matched the WHERE clause,
	// meaning the ID does not exist. this prevents silent no-ops.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected for deployment %q: %w", id, err)
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// UpdateURL sets the public URL for a deployment once the container is live.
// called by the pipeline after the Nginx container starts successfully.
func (database *Database) UpdateURL(id string, url string) error {
	query := `UPDATE deployments SET url = ?, updated_at = ? WHERE id = ?`

	result, err := database.conn.Exec(query, url, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update url for deployment %q: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected for deployment %q: %w", id, err)
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// DeleteDeployment removes a deployment row by ID.
// the caller is responsible for stopping the container and removing files
// before calling this function. the Database row is the last thing deleted.
func (database *Database) DeleteDeployment(id string) error {
	query := `DELETE FROM deployments WHERE id = ?`

	result, err := database.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment %q: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected for deployment %q: %w", id, err)
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// scanner is an interface satisfied by both *sql.Row and *sql.Rows.
// this allows scanDeployment to work with both QueryRow (single row)
// and Query (multiple rows) without duplicating the scan logic.
// Implicit Interfaces (Duck Typing): In Go, interfaces are satisfied
//
//	implicitly. Because both *sql.Row (returned by QueryRow) and
//	*sql.Rows (returned by Query) possess a Scan(dest ...any) error
//	method, they both automatically satisfy this interface.
type scanner interface {
	Scan(dest ...any) error
}

// scanDeployment reads and converts/serializes a single database row into a Deployment struct.
// all pointer fields (GitHubURL, EnvVars, URL, WebhookSecret) are scanned
// into their pointer types directly; database/sql sets them to nil for NULL columns.
func scanDeployment(row scanner) (*models.Deployment, error) {
	var deployment models.Deployment

	// The Scan() method requires memory addresses
	//    (pointers, via the & operator) for the destination variables. It
	//    reads the raw database columns sequentially and overwrites the
	//    memory addresses of the struct fields with the parsed Go types.
	err := row.Scan(
		&deployment.ID,
		&deployment.Slug,
		&deployment.Name,
		&deployment.SourceType,
		&deployment.GitHubURL, // scans NULL -> nil *string
		&deployment.Branch,
		&deployment.BuildCmd,
		&deployment.OutputDir,
		&deployment.EnvVars, // scans NULL -> nil *string
		&deployment.Status,
		&deployment.URL,           // scans NULL -> nil *string
		&deployment.WebhookSecret, // scans NULL -> nil *string
		&deployment.AutoDeploy,    // scans INTEGER 0/1 -> bool
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}
