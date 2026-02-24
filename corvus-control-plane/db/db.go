// Package db manages the SQLite database connection and schema migrations.
// it exposes a Database struct that wraps *sql.DB and is passed via dependency
// injection to any layer that needs database access.
package db

import (
	"database/sql" // standard lib for SQL acess. provides DB connection pool and query execution methods
	"fmt"
	"log/slog"
	"os"            // to change file access permission to modify database file
	"path/filepath" // to create parent directory for the database file if it doesn't exist

	// the underscore import registers the go-sqlite3 driver with database/sql
	// without this import, sql.Open("sqlite3", ...) returns "unknown driver" error.
	// the package is never referenced directly in code, only its init() side effect is needed.
	// It tells Go: "I'm not going to call any functions from this package directly,
	// but I want you to run its init() function."
	// That init() function registers the string "sqlite3" into the global sql package.
	_ "github.com/mattn/go-sqlite3"
)

/*
Database acts as a specialized wrapper for the database layer.
wrapping rather than embedding keeps the public surface area intentional.
only methods defined on this struct are exposed to callers.
if the underlying driver changes (e.g. Postgres), only this file changes.

*sql.DB: A pointer to the standard library's database handle. It

	   manages a "connection pool," which is a set/list of active connections
	   maintained for reuse to improve performance and resource management.
		(doesn't have to make a new connection everytime)

Wrapping vs. Embedding: This struct uses wrapping (encapsulation).

	By making the *sql.Database field private (lowercase), it prevents
	external packages (callers) from accessing raw database methods
	directly. Callers are restricted to the high-level API defined
	specifically for the CorvusPaas domain logic.

Exposed Surface Area: This design pattern ensures the "public

	surface" of the database package remains small and intentional.
	It decouples the application logic from the underlying driver,
	allowing for potential migrations (eg, SQLite to PostgreSQL)
*/
type Database struct {
	connection *sql.DB
	logger     *slog.Logger
}

// migrate() is a method attached to Database. It runs the schema DDL against the database.
// Basically, it just creates empty tables and columns based on the schema query/definition
// Why not pass in the schema as an argument?
// Because the schema is a constant defined in this file, and it is not expected to change at runtime.
// It is tightly coupled with the Database struct and its migration logic, so keeping it as a constant
// within the same file keeps the code organized and encapsulated.
// If the schema were to be passed in as an argument, the caller of the migrate() function
// would need to know about the schema definition and manage it. In turn, the OpenDatabase() function will
// have to know about schema, in turn, main() or another caller have to know
// which would break the separation of concerns.
func (database *Database) migrate() error {
	_, err := database.connection.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema migration (create tables & columns): %w", err)
	}
	return nil
}

/*
schema is the SQL DDL that defines the deployments table.
It uses IF NOT EXISTS so it is safe to run on every startup, no run if the table exists, creates it if not.
this is a minimal migration strategy appropriate for a single-node PoC (proof of concept)
For a production system with multiple schema versions, a proper migration
library (e.g. golang-migrate) would be used instead.
*/

const schema = `
CREATE TABLE IF NOT EXISTS deployments (
    id             TEXT PRIMARY KEY,
    slug           TEXT UNIQUE NOT NULL,
    name           TEXT NOT NULL,
    source_type    TEXT NOT NULL,
    github_url     TEXT,
    branch         TEXT NOT NULL DEFAULT 'main',
    build_cmd      TEXT NOT NULL DEFAULT '',
    output_dir     TEXT NOT NULL DEFAULT '.',
    env_vars       TEXT,
    status         TEXT NOT NULL,
    url            TEXT,
    webhook_secret TEXT,
    auto_deploy    INTEGER NOT NULL DEFAULT 0,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

/*
OpenDatabase opens the SQLite database at the given file path, runs the schema
migration (from the query schema), and returns a ready-to-use *Database.
The directory for the database file is created if it does not exist,
so the caller function does not need to pre-create the path on disk.
*/
func OpenDatabase(dbPath string, logger *slog.Logger) (*Database, error) {
	// create the parent directory of the db file if it does not exist.
	// os.MkdirAll is a not run if the directory already exists.
	dir := filepath.Dir(dbPath)

	// 0755 represents standard Unix directory permissions in octal.
	// 0 is prefix for octal.
	// then it goes 7 for owner (read/write/execute), 5 for group (read/execute), 5 for others (read/execute).
	// now the Go process has permission to manage the database file within the directory.
	err := os.MkdirAll(dir, 0755)
	if err != nil { // error on creating directory
		// %q = quoted string, %w = wrap original error inside the new error.
		// this allows the caller to check for specific error types using errors.Is() or errors.As()
		return nil, fmt.Errorf("failed to create database directory %q: %w", dir, err)
	}

	// sql.Open does not actually open a dbConnection. it only validates the arguments
	// and prepares the pool. the real dbConnection is established on the first query
	// so it is recommended to test the connection right after opening (done here with migration call later)
	// "sqlite3" is the driver name registered by the go-sqlite3 init() function.
	dbConnection, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %q: %w", dbPath, err)
	}

	// SQLite does not support concurrent writes from multiple connections.
	// setting MaxOpenConns to 1 prevents "database is locked" errors that occur
	// when the dbConnection pool opens multiple connections and they write simultaneously (error)
	dbConnection.SetMaxOpenConns(1)

	// init database struct with the opened connection and logger
	database := &Database{
		connection: dbConnection,
		logger:     logger,
	}

	// Trying to create tables & columns (schema migration) immediately after opening the connection
	// if this fails, the application cannot function, so the error is returned
	// to the caller (main.go) which will log it and close the application
	// since the app is useless without a working database, it is better to fail fast here
	// The migration uses "IF NOT EXISTS" logic, ensuring  that the operation is safe to
	// run on every application startup without destroying existing data or causing errors
	err = database.migrate()
	if err != nil {
		return nil, fmt.Errorf("database migration (table & column creation, DDL) failed: %w", err)
	}

	logger.Info("database opened and schema migrated", "path", dbPath)
	return database, nil
}

// CloseDatabase releases the database connection pool. (closes connection)
// this should be deferred in main.go immediately after Open returns successfully.
func (database *Database) CloseDatabase() error {
	return database.connection.Close()
}
