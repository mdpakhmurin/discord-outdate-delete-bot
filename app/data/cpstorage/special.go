package cpstorage

// Special CRUD operations

// Get channels with remove date before specified date (unix time)
// These channels may contain outdated messages for removing
func GetChannelsWithRemoveDateBeforeMoment(momentUnixTime int64) (channels []*ChannelPropertiesEntity, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	query := `
        SELECT * FROM channels
        WHERE next_remove_date < ?
    `
	err = sqlxdb.Select(&channels, query, momentUnixTime)

	return
}

// Update last activity date (unix time) for channel
func UpdateChannelLastActivity(channelID string, lastActivityUnixTime int64) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, err = db.Exec("UPDATE channels SET last_activity_date = ? WHERE channel_id = ?", lastActivityUnixTime, channelID)
	return
}
