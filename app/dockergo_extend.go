package main

// Common functions for working with discord objects. Extending discordgo

import (
	"strconv"
	"time"
)

// Converts a timestamp to a snowflakeid.
// Please note that this ID is not real and it is only used to indicate the time
func TimestampToSnowflakeId(timestamp time.Time) (ID string) {
	snowflakeID := (timestamp.UnixNano()/1000000 - 1420070400000) << 22
	ID = strconv.FormatInt(snowflakeID, 10)
	return ID
}
