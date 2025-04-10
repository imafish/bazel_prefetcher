package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Item struct {
	ID           int64     `json:"id" db:"id"`
	Size         int64     `json:"size"`
	Path         string    `json:"path"`
	Url          string    `json:"url"`
	Hash         string    `json:"hash"`
	UrlHash      string    `json:"url_hash"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

type ItemTable struct {
	db *sql.DB
}

func NewItemTable(db *sql.DB) *ItemTable {
	return &ItemTable{db: db}
}

func (t *ItemTable) Create() error {
	// Check if the table already exists
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='items'`
	row := t.db.QueryRow(query)

	var tableName string
	err := row.Scan(&tableName)
	if err == nil && tableName == "items" {
		// Table already exists, no need to create it
		return nil
	}

	// Create the table if it does not exist
	query = `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		size INTEGER,
		path TEXT,
		url TEXT,
		hash TEXT,
		url_hash TEXT,
		downloaded_at DATETIME
	)`
	_, err = t.db.Exec(query)
	return err
}

func (t *ItemTable) Drop() error {
	query := `DROP TABLE IF EXISTS items`
	_, err := t.db.Exec(query)
	return err
}

func (t *ItemTable) Insert(item *Item) error {
	// Set DownloadedAt to the current time
	item.DownloadedAt = time.Now()

	query := `INSERT INTO items (size, path, url, hash, url_hash, downloaded_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := t.db.Exec(query, item.Size, item.Path, item.Url, item.Hash, item.UrlHash, item.DownloadedAt)
	if err != nil {
		return err
	}
	item.ID, err = result.LastInsertId()
	return err
}

func (t *ItemTable) CreateOrUpdate(item *Item) error {
	// Check if the item exists by URL
	existingItem, err := t.GetByUrl(item.Url)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existingItem != nil {
		// Update the existing item and set DownloadedAt to the current time
		item.DownloadedAt = time.Now()
		query := `UPDATE items SET 
				  size = ?, 
				  path = ?, 
				  hash = ?, 
				  url_hash = ?, 
				  downloaded_at = ? 
				  WHERE url = ?`
		_, err = t.db.Exec(query, item.Size, item.Path, item.Hash, item.UrlHash, item.DownloadedAt, item.Url)
		return err
	}

	// Insert a new item if it does not exist
	return t.Insert(item)
}

func (t *ItemTable) GetByID(id int64) (*Item, error) {
	query := `SELECT id, size, path, url, hash, url_hash, downloaded_at FROM items WHERE id = ?`
	row := t.db.QueryRow(query, id)

	var item Item
	err := row.Scan(&item.ID, &item.Size, &item.Path, &item.Url, &item.Hash, &item.UrlHash, &item.DownloadedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (t *ItemTable) GetByUrl(url string) (*Item, error) {
	query := `SELECT id, size, path, url, hash, url_hash, downloaded_at FROM items WHERE url = ?`
	row := t.db.QueryRow(query, url)

	var item Item
	err := row.Scan(&item.ID, &item.Size, &item.Path, &item.Url, &item.Hash, &item.UrlHash, &item.DownloadedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (t *ItemTable) DeleteByID(id int64) error {
	query := `DELETE FROM items WHERE id = ?`
	_, err := t.db.Exec(query, id)
	return err
}

func (t *ItemTable) GetAll() ([]Item, error) {
	query := `SELECT id, size, path, url, hash, url_hash, downloaded_at FROM items`
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Size, &item.Path, &item.Url, &item.Hash, &item.UrlHash, &item.DownloadedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (t *ItemTable) DebugPrintAll() error {
	items, err := t.GetAll()
	if err != nil {
		return err
	}

	for _, item := range items {
		// Print each item in a readable format
		fmt.Printf("ID: %d, Size: %d, Path: %s, URL: %s, Hash: %s, URL Hash: %s, Downloaded At: %s\n",
			item.ID, item.Size, item.Path, item.Url, item.Hash, item.UrlHash, item.DownloadedAt)
	}

	return nil
}
