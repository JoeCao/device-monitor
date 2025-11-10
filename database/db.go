package database

import (
	"database/sql"
	"device-monitor-go/config"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB() error {
	// Ensure database directory exists
	dbPath := config.AppConfig.DatabasePath
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set database configurations
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	DB = db

	// Initialize schema
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS device_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id VARCHAR(100) NOT NULL,
		session_id VARCHAR(100) NOT NULL UNIQUE,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		duration INTEGER DEFAULT 0,
		status VARCHAR(20) DEFAULT 'running',
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_device_id ON device_sessions(device_id);
	CREATE INDEX IF NOT EXISTS idx_session_id ON device_sessions(session_id);
	CREATE INDEX IF NOT EXISTS idx_status ON device_sessions(status);
	CREATE INDEX IF NOT EXISTS idx_start_time ON device_sessions(start_time);

	-- IoT data points table (kept for compatibility but not used for storage)
	CREATE TABLE IF NOT EXISTS iot_data_points (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id VARCHAR(100) NOT NULL,
		point_name VARCHAR(100) NOT NULL,
		point_value REAL,
		unit VARCHAR(20),
		timestamp DATETIME NOT NULL,
		raw_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES device_sessions(session_id)
	);

	CREATE INDEX IF NOT EXISTS idx_iot_session_id ON iot_data_points(session_id);
	CREATE INDEX IF NOT EXISTS idx_iot_point_name ON iot_data_points(point_name);
	CREATE INDEX IF NOT EXISTS idx_iot_timestamp ON iot_data_points(timestamp);
	`

	_, err := DB.Exec(schema)
	return err
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Transaction helper
func WithTx(fn func(*sqlx.Tx) error) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// NullTime helper for handling nullable timestamps
type NullTime struct {
	sql.NullTime
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", nt.Time.Format("2006-01-02T15:04:05.000Z"))), nil
}