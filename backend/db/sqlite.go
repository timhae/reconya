package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectToSQLite initializes and returns a SQLite connection
func ConnectToSQLite(dbPath string) (*sql.DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for SQLite: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	log.Println("Connected to SQLite database")
	return db, nil
}

// InitializeSchema creates all the necessary tables if they don't exist
func InitializeSchema(db *sql.DB) error {
	// Create networks table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS networks (
		id TEXT PRIMARY KEY,
		cidr TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("failed to create networks table: %w", err)
	}

	// Create devices table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		ipv4 TEXT NOT NULL,
		mac TEXT,
		vendor TEXT,
		status TEXT NOT NULL,
		network_id TEXT,
		hostname TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		last_seen_online_at TIMESTAMP,
		port_scan_started_at TIMESTAMP,
		port_scan_ended_at TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create devices table: %w", err)
	}

	// Create ports table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS ports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL,
		number TEXT NOT NULL,
		protocol TEXT NOT NULL,
		state TEXT NOT NULL,
		service TEXT NOT NULL,
		FOREIGN KEY (device_id) REFERENCES devices(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create ports table: %w", err)
	}

	// Create event_logs table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS event_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		description TEXT NOT NULL,
		device_id TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("failed to create event_logs table: %w", err)
	}

	// Create system_status table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS system_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		network_id TEXT,
		public_ip TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create system_status table: %w", err)
	}

	// Create local_device table for system_status
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS local_devices (
		system_status_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		ipv4 TEXT NOT NULL,
		mac TEXT,
		vendor TEXT,
		status TEXT NOT NULL,
		hostname TEXT,
		PRIMARY KEY (system_status_id),
		FOREIGN KEY (system_status_id) REFERENCES system_status(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create local_devices table: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}