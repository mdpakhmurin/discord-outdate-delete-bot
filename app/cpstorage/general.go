package cpstorage

// General CRUD operations

import (
	"database/sql"
	"fmt"
)

// Delete channel properies
func DeleteChannelProperties(channelID string) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("DELETE FROM channels WHERE channel_id = ?", channelID)
	return
}

// Delete channels properties
func DeleteChannelsProperties(channelIDs []string) (err error) {
	if len(channelIDs) == 0 {
		return nil
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	// Use transaction to delete all array as one action
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rolbackErr := transaction.Rollback()
			if rolbackErr != nil {
				err = fmt.Errorf("failed to rollback transcation %v, after write error: %v", rolbackErr, err)
			}
		} else {
			err = transaction.Commit()
		}
	}()

	// prepared delete statement
	statement, err := transaction.Prepare(`
		DELETE FROM channels 
		WHERE channel_id = ?
    `)
	if err != nil {
		return err
	}
	defer statement.Close()

	// insert each of channels ids in prepared statement
	for _, channelId := range channelIDs {
		_, err = statement.Exec(
			channelId,
		)
		if err != nil {
			return err
		}
	}

	return nil
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

	if err == sql.ErrNoRows {
		err = nil
		channelProperties = nil
	}

	return
}

// Save channel properties
func WriteChannelProperties(channelProperties *ChannelPropertiesEntity) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec(`
		INSERT OR REPLACE INTO channels
			(channel_id, timeout, last_activity_date, next_remove_date)
			VALUES (?, ?, ?, ?)`,
		channelProperties.ChannelID,
		channelProperties.Timeout,
		channelProperties.LastActivityDateUnix,
		channelProperties.NextRemoveDateUnix)

	return
}

// Save channels properties
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
			rolbackErr := transaction.Rollback()
			if rolbackErr != nil {
				err = fmt.Errorf("failed to rollback transcation %v, after write error: %v", rolbackErr, err)
			}
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
			channelProperties.LastActivityDateUnix,
			channelProperties.NextRemoveDateUnix,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update last activity date (unix time) for channel
func UpdateChannelLastActivityDate(channelID string, lastActivityUnixTime int64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("UPDATE channels SET last_activity_date = ? WHERE channel_id = ?", lastActivityUnixTime, channelID)
	return
}

// Update last activity date (unix time) for channel
func UpdateChannelNextRemoveDate(channelID string, nextRemoveDateUnixTime int64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("UPDATE channels SET next_remove_date = ? WHERE channel_id = ?", nextRemoveDateUnixTime, channelID)
	return
}
