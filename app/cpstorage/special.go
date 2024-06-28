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

// Get channels IDs with remove date before specified date (unix time)
// These channels may contain outdated messages for removing
func GetChannelsIdsWithRemoveDateBeforeMoment(momentUnixTime int64) (channelIDs []string, err error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	query := `
        SELECT channel_id FROM channels
        WHERE next_remove_date < ?
    `
	err = sqlxdb.Select(&channelIDs, query, momentUnixTime)

	return
}
