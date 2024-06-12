package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func RemoveOldMessages() {
	for {
		// Get timeouts for each channel
		channelsTimeouts, err := GetAllChannelsTimeout()
		if err != nil {
			log.Fatalf("Failed to get query timeouts: %v", err)
		}

		var unavailableСhannelsToDelete []string

		for _, channelTimeout := range channelsTimeouts {
			channelID := channelTimeout.ChannelID

			// Get outdate messages
			messages, err := getOutdateChannelMessages(30, channelTimeout)
			if err != nil {
				// Is the message unavailable due to access issues
				if errD, ok := err.(*discordgo.RESTError); ok {
					if errD.Message != nil {
						switch errD.Message.Code {
						case discordgo.ErrCodeUnknownChannel, discordgo.ErrCodeMissingAccess, discordgo.ErrCodeMissingPermissions:
							unavailableСhannelsToDelete = append(unavailableСhannelsToDelete, channelID)
						}
					}
				} else {
					log.Printf("Failed to get messages: %v", err)
				}
				continue
			}
			// Pinned messages will not be deleted
			messages = excludePinnedMessages(messages)

			// Delete outdate messages
			err = deleteChannelMessages(channelID, messages)
			if err != nil {
				log.Printf("Failed to delete messages: %v", err)
			}
		}

		// Delete unavailable channels
		err = DeleteChannelsTimeouts(unavailableСhannelsToDelete)
		if err != nil {
			log.Printf("Failed to delete channels: %v", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func getOutdateChannelMessages(messagesNumber int, channelTimeout ChannelTimeoutEntity) (messages []*discordgo.Message, err error) {
	lastTimestampForRemove := time.Now().Add(-time.Duration(channelTimeout.Timeout * float64(time.Hour)))
	lastIdForRmove := TimestampToSnowflakeId(lastTimestampForRemove)

	messages, err = session.ChannelMessages(channelTimeout.ChannelID, messagesNumber, lastIdForRmove, "", "")

	return
}

func excludePinnedMessages(messages []*discordgo.Message) (unpinnedMessages []*discordgo.Message) {
	for _, message := range messages {
		if !message.Pinned {
			unpinnedMessages = append(unpinnedMessages, message)
		}
	}
	return
}

func deleteChannelMessages(channelID string, messages []*discordgo.Message) (err error) {
	messageIDs := make([]string, len(messages))
	for i, message := range messages {
		messageIDs[i] = message.ID
	}
	err = session.ChannelMessagesBulkDelete(channelID, messageIDs)
	return err
}
