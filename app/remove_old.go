package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/data/cpstorage"
)

func RemoveOldMessages() {
	for {
		// Get timeouts for each channel
		channelsProperties, err := cpstorage.GetChannelsWithRemoveDateBeforeMoment(time.Now().Unix())
		if err != nil {
			log.Fatalf("Failed to get query timeouts: %v", err)
		}

		var channelsForRemove []string

		for i := range channelsProperties {
			channelProperties := channelsProperties[i]
			channelID := channelProperties.ChannelID

			// Get outdate messages
			messages, err := getChannelMessagesBeforeOutdateTime(30, channelProperties)
			// Unavailable channels must be deleted
			if err != nil && isErrorChannelUnavailable(err) {
				log.Printf("Channel %s is unavaliable", channelID)
				channelsForRemove = append(channelsForRemove, channelID)
			} else if err != nil {
				log.Printf("Failed to get messages: %v", err)
			}

			// Pinned messages will not be deleted
			messages = excludePinnedMessages(messages)

			// Delete outdate messages
			err = deleteChannelMessages(channelID, messages)
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
				channelsForRemove = append(channelsForRemove, channelID)
			}
		}

		// Update channels properties
		cpstorage.WriteChannelsProperties(channelsProperties)

		// Delete unavailable channels
		err = cpstorage.DeleteChannelsProperties(channelsForRemove)
		if err != nil {
			log.Printf("Failed to delete channels: %v", err)
		}

		time.Sleep(1 * time.Second)
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
	return timeScienceLastActivity.Hours() > RemoveInactiveChannelTimeoutHorus
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

// Get messages that are already outdated
func getChannelMessagesBeforeOutdateTime(messagesNumber int, channelProperties *cpstorage.ChannelPropertiesEntity) (messages []*discordgo.Message, err error) {
	outdateSnwoflakeId := getChannelOutdateTimeInSnowflakeFormat(channelProperties)
	messages, err = Session.ChannelMessages(channelProperties.ChannelID, messagesNumber, outdateSnwoflakeId, "", "")

	return
}

// Get messages that were sent after outdate time
func getChannelMessagesAfterOutdateTime(messagesNumber int, channelProperties *cpstorage.ChannelPropertiesEntity) (messages []*discordgo.Message, err error) {
	outdateSnowflakeId := getChannelOutdateTimeInSnowflakeFormat(channelProperties)
	messages, err = Session.ChannelMessages(channelProperties.ChannelID, messagesNumber, "", outdateSnowflakeId, "")

	return
}

// Get time when messages became outdated and converts it to snowflake format
func getChannelOutdateTimeInSnowflakeFormat(channelProperties *cpstorage.ChannelPropertiesEntity) (snowflakeId string) {
	outdateTimestamp := time.Now().Add(-time.Duration(channelProperties.Timeout * float64(time.Hour)))
	outdateSnowflakeId := TimestampToSnowflakeId(outdateTimestamp)

	return outdateSnowflakeId
}

// Filter pinned messages
func excludePinnedMessages(messagesAfterOutdate []*discordgo.Message) (unpinnedMessages []*discordgo.Message) {
	for _, message := range messagesAfterOutdate {
		if !message.Pinned {
			unpinnedMessages = append(unpinnedMessages, message)
		}
	}
	return
}

// Delete messages in channel
func deleteChannelMessages(channelID string, messages []*discordgo.Message) (err error) {
	messageIDs := make([]string, len(messages))
	for i, message := range messages {
		messageIDs[i] = message.ID
	}
	err = Session.ChannelMessagesBulkDelete(channelID, messageIDs)
	return err
}
