package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/cpstorage"
)

func RemoveOldMessages() {
	for {
		channelIdsForRemove, err := cpstorage.GetChannelsIdsWithRemoveDateBeforeMoment(time.Now().Unix())
		if err != nil {
			log.Fatalf("Failed to get channels for remove: %v", err)
		}

		for _, channelId := range channelIdsForRemove {
			isChannelToDelete := false

			channelProperties, err := cpstorage.GetChannelProperties(channelId)
			if err != nil {
				log.Printf("Failed to get channel %s: %v", channelId, err)
			}

			// Get outdate messages.
			messages, err := getChannelMessagesForRemove(channelProperties)

			// Unavailable channels must be deleted
			if err != nil && isErrorChannelUnavailable(err) {
				log.Printf("Channel %s is unavaliable", channelId)
				isChannelToDelete = true
			} else if err != nil {
				log.Printf("Failed to get messages: %v", err)
			}

			// Delete outdate messages
			err = deleteChannelMessages(channelId, messages)
			if err != nil {
				log.Printf("Failed to delete messages: %v", err)
			}

			// Update last activity if there are deleted messages
			if len(messages) > 0 {
				channelProperties.LastActivityDateUnix = time.Now().Unix()
			}

			// Get next remove date in unix format
			nextRemoveDateUnix, err := getNextRemoveDateUnix(len(messages) > 0, channelProperties)
			if err != nil {
				fmt.Printf("Failed to get next remove date: %v", err)
				nextRemoveDateUnix = 0
			}
			channelProperties.NextRemoveDateUnix = nextRemoveDateUnix

			// Inactive channels must be deleted
			if isChannelInactive(channelProperties) {
				isChannelToDelete = true
			}

			// Update channel properties
			err = cpstorage.UpdateChannelLastActivityDate(channelId, channelProperties.LastActivityDateUnix)
			if err != nil {
				log.Printf("Failed to update channel last activity %s: %v", channelId, err)
			}
			err = cpstorage.UpdateChannelNextRemoveDate(channelId, channelProperties.NextRemoveDateUnix)
			if err != nil {
				log.Printf("Failed to update channel next remove date %s: %v", channelId, err)
			}

			// Delete channel
			if isChannelToDelete {
				err = cpstorage.DeleteChannelProperties(channelId)
				if err != nil {
					log.Printf("Failed to delete channel %s: %v", channelId, err)
				}
			}
		}

		time.Sleep(3 * time.Second)
	}
}

// Get next date for removing in channel
func getNextRemoveDateUnix(isMessagesToRemoveExists bool, channelProperties *cpstorage.ChannelPropertiesEntity) (nextRemoveDateUnix int64, err error) {
	// if messages were deleted, perform removing again
	if isMessagesToRemoveExists {
		return 0, nil
	}

	// get a message that will be deleted next in the future
	messagesAfterOutdate, err := getChannelMessagesAfterOutdateTime(1, channelProperties)
	if err != nil {
		return 0, err
	}

	// if no message - remove time will be after the channel timeout time
	if len(messagesAfterOutdate) == 0 {
		nextRemoveDate := time.Now().Add(time.Duration(channelProperties.Timeout * float64(time.Hour)))
		return nextRemoveDate.Unix(), nil
	}

	// get message sending time
	newMessageTimestamp, err := discordgo.SnowflakeTimestamp(messagesAfterOutdate[0].ID)
	if err != nil {
		return 0, err
	}

	// Next remove time is equal to the time the message was sent + timeout time
	nextRemoveDate := newMessageTimestamp.Add(time.Duration(channelProperties.Timeout * float64(time.Hour)))
	return nextRemoveDate.Unix(), nil
}

// Check if there has been chat activity for too long
func isChannelInactive(channelProperties *cpstorage.ChannelPropertiesEntity) (isInactive bool) {
	timeScienceLastActivity := time.Since(time.Unix(channelProperties.LastActivityDateUnix, 0))
	return timeScienceLastActivity.Hours() > Config.RemoveInactiveChannelTimeoutHours
}

// Check is error of getting messages from channel is access problem
func isErrorChannelUnavailable(discordMessageGetError error) (isUnavailable bool) {
	if errD, ok := discordMessageGetError.(*discordgo.RESTError); ok {
		if errD.Message != nil {
			switch errD.Message.Code {
			case discordgo.ErrCodeUnknownChannel, discordgo.ErrCodeMissingAccess, discordgo.ErrCodeMissingPermissions:
				return true
			}
		}
	}
	return false
}

// Get messages that must be removed
func getChannelMessagesForRemove(channelProperties *cpstorage.ChannelPropertiesEntity) (messages []*discordgo.Message, err error) {
	outdateSnwoflakeId := getChannelOutdateTimeInSnowflakeIdFormat(channelProperties)

	messages, err = Session.ChannelMessages(channelProperties.ChannelID, Config.RemoveBatchSize, outdateSnwoflakeId, "", "")
	if err != nil {
		return messages, err
	}

	// Too old messages cannot be deleted. Bad work if use old id in ChannelMessages
	messages = excludeTooOldMessages(messages)

	// Pinned messages will not be deleted
	messages = excludePinnedMessages(messages)

	// First message in thread can not be deleted
	messages = excludeThreadStartMessages(messages)

	return messages, nil
}

// Get messages that were sent after outdate time
func getChannelMessagesAfterOutdateTime(messagesNumber int, channelProperties *cpstorage.ChannelPropertiesEntity) (messages []*discordgo.Message, err error) {
	outdateSnowflakeId := getChannelOutdateTimeInSnowflakeIdFormat(channelProperties)
	messages, err = Session.ChannelMessages(channelProperties.ChannelID, messagesNumber, "", outdateSnowflakeId, "")

	return
}

// Get time when messages became outdated and converts it to snowflake format
func getChannelOutdateTimeInSnowflakeIdFormat(channelProperties *cpstorage.ChannelPropertiesEntity) (snowflakeId string) {
	outdateTimestamp := time.Now().Add(-time.Duration(channelProperties.Timeout * float64(time.Hour)))
	outdateSnowflakeId := TimestampToSnowflakeId(outdateTimestamp)

	return outdateSnowflakeId
}

// Get time when messages became too old and can not be deleted
func getTooOldTimeInSnoflakeIdFormat() (snowflakeId string) {
	tooOldTimeStamp := time.Now().Add(-time.Hour * time.Duration(Config.OldDontRemoveTimeoutHours))
	tooOldSnowflakeId := TimestampToSnowflakeId(tooOldTimeStamp)

	return tooOldSnowflakeId
}

// Filter pinned messages
func excludePinnedMessages(messagesAfterOutdate []*discordgo.Message) (filteredMessages []*discordgo.Message) {
	for _, message := range messagesAfterOutdate {
		if !message.Pinned {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return filteredMessages
}

// Filter thread start message (first message in thread)
func excludeThreadStartMessages(messagesAfterOutdate []*discordgo.Message) (filteredMessages []*discordgo.Message) {
	for _, message := range messagesAfterOutdate {
		if message.Type != discordgo.MessageTypeThreadStarterMessage {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return filteredMessages
}

// Filter too old message that cannot be deleted by API
func excludeTooOldMessages(messagesAfterOutdate []*discordgo.Message) (filteredMessages []*discordgo.Message) {
	tooOldSnoflakeId := getTooOldTimeInSnoflakeIdFormat()

	for _, message := range messagesAfterOutdate {
		if message.ID > tooOldSnoflakeId {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return filteredMessages
}

// Delete messages in channel with their threads
func deleteChannelMessages(channelID string, messages []*discordgo.Message) (err error) {
	// delete messages
	messageIDs := make([]string, len(messages))
	for i, message := range messages {
		messageIDs[i] = message.ID
	}

	err = Session.ChannelMessagesBulkDelete(channelID, messageIDs)
	if err != nil {
		return err
	}

	// delete threads
	for _, message := range messages {
		if message.Thread != nil {
			_, err = Session.ChannelDelete(message.Thread.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
