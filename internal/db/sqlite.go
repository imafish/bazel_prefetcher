package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver
)

func LoadDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateAndLoadDatabase(dbPath string) (*sql.DB, error) {
	// Check if the database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Create the database file
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	// Load the database
	return LoadDatabase(dbPath)
}
