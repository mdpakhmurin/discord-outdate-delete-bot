package main

// Abstraction for working with channels properties storage

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

type ChannelPropertiesEntity struct {
	ChannelID    string  `db:"channel_id"`
	Timeout      float64 `db:"timeout"`
	LastActivity int64   `db:"last_activity"`
}

func init() {
	initDb()
}

func initDb() {
	openTableConnection()
	createTableIfNotExists()

	sqlxdb = sqlx.NewDb(db, "sqlite3")
}

// Open global connection to DB
func openTableConnection() {
	var err error
	db, err = sql.Open("sqlite", SharedDataPath+"/channels.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
}

// Create table if it doesn't exists
func createTableIfNotExists() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS channels (channel_id TEXT PRIMARY KEY, timeout REAL, last_activity INTEGER)")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// Delete channel properies
func DeleteChannelProperties(channelID string) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("DELETE FROM channels WHERE channel_id = ?", channelID)
	return
}

// Remove channels properties
func DeleteChannelsProperties(channelIDs []string) (err error) {
	if len(channelIDs) == 0 {
		return nil
	}

	// Create a string with placeholders for each channel ID
	placeholders := strings.Repeat(",?", len(channelIDs)-1)

	query := fmt.Sprintf("DELETE FROM channels WHERE channel_id IN (?%s)", placeholders)

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

// Get all channels properties
func GetAllChannelsProperties() (channelsProperties []*ChannelPropertiesEntity, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	err = sqlxdb.Select(&channelsProperties, "SELECT * FROM channels")
	return
}

// Get channel properties
func GetChannelProperties(channelID string) (channelProperties *ChannelPropertiesEntity, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	channelProperties = &ChannelPropertiesEntity{}
	err = sqlxdb.Get(channelProperties, "SELECT * FROM channels WHERE channel_id = $1", channelID)

	return
}

// Save channels properties
func WriteChannelProperties(channelProperties *ChannelPropertiesEntity) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("INSERT OR REPLACE INTO channels (channel_id, timeout, last_activity) VALUES (?, ?, ?)",
		channelProperties.ChannelID,
		channelProperties.Timeout,
		channelProperties.LastActivity)
	return
}

// Save channel properties
func WriteChannelsProperties(channelProperties []*ChannelPropertiesEntity) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	for _, timeout := range channelProperties {
		_, err = db.Exec("INSERT OR REPLACE INTO channels (channel_id, timeout, last_activity) VALUES (?, ?, ?)",
			timeout.ChannelID,
			timeout.Timeout,
			timeout.LastActivity)
	}
	return
}

// Update last activity for channel with unix time
func UpdateChanneLastActivity(channelID string, lastActivityUnixTime int64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("UPDATE channels SET last_activity = ? WHERE channel_id = ?", lastActivityUnixTime, channelID)
	return
}
