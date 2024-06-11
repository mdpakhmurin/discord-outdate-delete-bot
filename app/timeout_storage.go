package main

// Abstraction for working with timeout storage

// TODO: Remove channels after long unising

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	sqlxdb *sqlx.DB
	dbLock sync.RWMutex
)

type ChannelTimeoutEntity struct {
	ChannelID string  `db:"channel_id"`
	Timeout   float64 `db:"timeout"`
}

func init() {
	initDb()
}

func initDb() {
	// open database connection
	var err error
	db, err = sql.Open("sqlite", SharedDataPath+"/timeouts.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create a table if it doesn't exist yet
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS timeouts (channel_id TEXT PRIMARY KEY, timeout REAL)")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	sqlxdb = sqlx.NewDb(db, "sqlite3")
}

// Deletes information about the channel
func DeleteChannelTimeout(channelID string) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("DELETE FROM timeouts WHERE channel_id = ?", channelID)
	return
}

// Remove channels timeouts
func DeleteChannelsTimeouts(channelIDs []string) (err error) {
	if len(channelIDs) == 0 {
		return nil
	}

	// Create a string with placeholders for each channel ID
	placeholders := strings.Repeat(",?", len(channelIDs)-1)

	query := fmt.Sprintf("DELETE FROM timeouts WHERE channel_id IN (?%s)", placeholders)

	// Convert the slice of channel IDs into a slice of empty interfaces to use it with db.Exec
	args := make([]interface{}, len(channelIDs))
	for i, v := range channelIDs {
		args[i] = v
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec(query, args...)
	return
}

// Gets information about all saved channels
func GetAllChannelsTimeout() (timeouts []ChannelTimeoutEntity, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	err = sqlxdb.Select(&timeouts, "SELECT channel_id, timeout FROM timeouts")
	return
}

// Gets timeout of the specified channel
func GetChannelTimeout(channelID string) (hasTimeout bool, timeout float64, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	hasTimeout = true

	row := db.QueryRow("SELECT timeout FROM timeouts WHERE channel_id = ?", channelID)
	err = row.Scan(&timeout)

	if err == sql.ErrNoRows {
		hasTimeout = false
		err = nil
	}

	return
}

// Saves timeout for the specified channel
func WriteChannelTimeout(channelID string, hours float64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("INSERT OR REPLACE INTO timeouts (channel_id, timeout) VALUES (?, ?)", channelID, hours)
	return
}
