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
	ChannelID        string  `db:"channel_id"`
	Timeout          float64 `db:"timeout"`
	LastActivityDate int64   `db:"last_activity_date"`
	NextRemoveDate   int64   `db:"next_remove_date"`
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

// Get channels with remove date before specified date
// These channels may contain outdated messages for removing
func GetChannelsWithRemoveDateBeforeMoment(momentUnixtime int64) (channels []*ChannelPropertiesEntity, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	query := `
        SELECT * FROM channels
        WHERE next_remove_date < ?
    `
	err = sqlxdb.Select(&channels, query, momentUnixtime)

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

	_, err = db.Exec(`
		INSERT OR REPLACE INTO channels
			(channel_id, timeout, last_activity_date, next_remove_date)
			VALUES (?, ?, ?, ?)`,
		channelProperties.ChannelID,
		channelProperties.Timeout,
		channelProperties.LastActivityDate,
		channelProperties.NextRemoveDate)
	return
}

// Save channel properties
func WriteChannelsProperties(channelsProperties []*ChannelPropertiesEntity) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	// Use transaction to save all array as one action
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			transaction.Rollback()
		} else {
			err = transaction.Commit()
		}
	}()

	// prepared write statement
	statement, err := transaction.Prepare(`
        INSERT OR REPLACE INTO channels
            (channel_id, timeout, last_activity_date, next_remove_date)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer statement.Close()

	// insert each of channels properties in prepared statement
	for _, channelProperties := range channelsProperties {
		_, err = statement.Exec(
			channelProperties.ChannelID,
			channelProperties.Timeout,
			channelProperties.LastActivityDate,
			channelProperties.NextRemoveDate,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update last activity for channel with unix time
func UpdateChanneLastActivity(channelID string, lastActivityUnixTime int64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("UPDATE channels SET last_activity_date = ? WHERE channel_id = ?", lastActivityUnixTime, channelID)
	return
}
