package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func RemoveOldMessages() {
	for {
		// Get timeouts for each channel
		channelsProperties, err := GetAllChannelsProperties()
		if err != nil {
			log.Fatalf("Failed to get query timeouts: %v", err)
		}

		var channelsForRemove []string

		for i := range channelsProperties {
			channelProperties := channelsProperties[i]
			channelID := channelProperties.ChannelID

			// Get outdate messages
			messages, err := getOutdateChannelMessages(30, channelProperties)
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
				channelProperties.LastActivity = time.Now().Unix()
			}

			// Inactive channels must be deleted
			if isChannelInactive(channelProperties) {
				channelsForRemove = append(channelsForRemove, channelID)
			}
		}

		// Update channels properties
		WriteChannelsProperties(channelsProperties)

		// Delete unavailable channels
		err = DeleteChannelsProperties(channelsForRemove)
		if err != nil {
			log.Printf("Failed to delete channels: %v", err)
		}

		time.Sleep(1 * time.Second)
	}
}

// Check if there has been chat activity for too long
func isChannelInactive(channelProperties *ChannelPropertiesEntity) (isInactive bool) {
	timeScienceLastActivity := time.Since(time.Unix(channelProperties.LastActivity, 0))
	return timeScienceLastActivity.Hours() > RemoveInactiveChannelTimeHorus
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
func getOutdateChannelMessages(messagesNumber int, channelTimeout *ChannelPropertiesEntity) (messages []*discordgo.Message, err error) {
	lastTimestampForRemove := time.Now().Add(-time.Duration(channelTimeout.Timeout * float64(time.Hour)))
	lastIdForRmove := TimestampToSnowflakeId(lastTimestampForRemove)

	messages, err = Session.ChannelMessages(channelTimeout.ChannelID, messagesNumber, lastIdForRmove, "", "")

	return
}

// Filter pinned messages
func excludePinnedMessages(messages []*discordgo.Message) (unpinnedMessages []*discordgo.Message) {
	for _, message := range messages {
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
