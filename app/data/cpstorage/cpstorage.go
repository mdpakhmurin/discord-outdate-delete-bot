// Global channels properties storage

package cpstorage

// Inititalization and params

import (
	"database/sql"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB      // Variable to store the connection to the database of type sql.DB
	sqlxdb *sqlx.DB     // Variable to store the connection to the database of type sqlx.DB (extended version of sql.DB)
	dbLock sync.RWMutex // Mutex (blocker) to restrict simultaneous access to the database
)

type ChannelPropertiesEntity struct {
	ChannelID            string  `db:"channel_id"`         // Channel ID
	Timeout              float64 `db:"timeout"`            // Time (hours) after which messages are deleted after sending
	LastActivityDateUnix int64   `db:"last_activity_date"` // Date (unixtime) of the last activity in the channel
	NextRemoveDateUnix   int64   `db:"next_remove_date"`   // Date (unixtime) of the next channel check for outdated messages
}

// Initializes the database globally (project).
// Create a new storage or uses an existing one if it exists
func Init(dbDirPath string) {
	openTableConnection(dbDirPath + "/channels.db")
	createChannelsTableIfNotExists()

	sqlxdb = sqlx.NewDb(db, "sqlite3")
}

// Open global (project) connection to DB.
// Open exists db file or create new
func openTableConnection(dbPath string) {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
}

// Create table if it doesn't exists
func createChannelsTableIfNotExists() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			channel_id TEXT PRIMARY KEY,
			timeout REAL,
			last_activity_date INTEGER,
			next_remove_date INTEGER
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}
