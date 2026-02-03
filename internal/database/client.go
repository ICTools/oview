package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Client provides database operations
type Client struct {
	masterDB *sql.DB
	dsn      string
}

// NewClient creates a new database client
func NewClient(dsn string) (*Client, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		masterDB: db,
		dsn:      dsn,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.masterDB != nil {
		return c.masterDB.Close()
	}
	return nil
}

// DatabaseExists checks if a database exists
func (c *Client) DatabaseExists(dbName string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err := c.masterDB.QueryRow(query, dbName).Scan(&exists)
	return exists, err
}

// CreateDatabase creates a new database
func (c *Client) CreateDatabase(dbName string) error {
	exists, err := c.DatabaseExists(dbName)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists
	}

	query := fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)
	_, err = c.masterDB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

// UserExists checks if a user exists
func (c *Client) UserExists(username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)"
	err := c.masterDB.QueryRow(query, username).Scan(&exists)
	return exists, err
}

// CreateUser creates a new database user
func (c *Client) CreateUser(username, password string) error {
	exists, err := c.UserExists(username)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists
	}

	query := fmt.Sprintf(`CREATE USER "%s" WITH PASSWORD '%s'`, username, password)
	_, err = c.masterDB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GrantAccess grants all privileges on a database to a user
func (c *Client) GrantAccess(dbName, username string) error {
	// First grant database-level privileges
	queries := []string{
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s"`, dbName, username),
	}

	for _, query := range queries {
		if _, err := c.masterDB.Exec(query); err != nil {
			return fmt.Errorf("failed to grant access: %w", err)
		}
	}

	// Now connect to the database and grant schema/table privileges
	dbDSN := replaceDatabaseInDSN(c.dsn, dbName)
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}
	defer db.Close()

	schemaQueries := []string{
		fmt.Sprintf(`GRANT ALL ON SCHEMA public TO "%s"`, username),
		fmt.Sprintf(`GRANT ALL ON ALL TABLES IN SCHEMA public TO "%s"`, username),
		fmt.Sprintf(`GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO "%s"`, username),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "%s"`, username),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "%s"`, username),
	}

	for _, query := range schemaQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to grant schema access: %w", err)
		}
	}

	return nil
}

// EnableExtension enables a Postgres extension in a database
func (c *Client) EnableExtension(dbName, extension string) error {
	// Connect to the specific database
	dbDSN := replaceDatabaseInDSN(c.dsn, dbName)
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", extension)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to enable extension %s: %w", extension, err)
	}

	return nil
}

// CreateSchema creates the RAG schema in a database with the specified embedding dimension
func (c *Client) CreateSchema(dbName string, embeddingDim int) error {
	// Connect to the specific database
	dbDSN := replaceDatabaseInDSN(c.dsn, dbName)
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}
	defer db.Close()

	// Execute schema SQL with the specified embedding dimension
	schemaSQL := GetSchemaSQL(embeddingDim)
	_, err = db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetConnection gets a connection to a specific database
func (c *Client) GetConnection(dbName string) (*sql.DB, error) {
	dbDSN := replaceDatabaseInDSN(c.dsn, dbName)
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database %s: %w", dbName, err)
	}

	return db, nil
}

// replaceDatabaseInDSN replaces the database name in a DSN string
func replaceDatabaseInDSN(dsn, newDB string) string {
	// Simple replacement for postgres:// DSNs
	// Format: postgres://user:pass@host:port/database?params
	// Find the database part (after last / before ?)
	lastSlash := -1
	questionMark := len(dsn)

	for i, c := range dsn {
		if c == '/' {
			lastSlash = i
		}
		if c == '?' {
			questionMark = i
			break
		}
	}

	if lastSlash == -1 {
		return dsn
	}

	return dsn[:lastSlash+1] + newDB + dsn[questionMark:]
}
