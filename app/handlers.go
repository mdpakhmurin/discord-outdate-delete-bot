package main

// Handlers

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/cpstorage"
)

var (
	// Map of command handlers
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		"set-timeout":    SetTimeoutCommandHandler,
		"info-timeout":   InfoCommandHandler,
		"remove-timeout": RemoveTimeoutCommandHandler,
	}
)

// Registers all handlers
func RegisterHandlers() {
	Session.AddHandler(CommandsHandler)
	Session.AddHandler(ReadyHandler)
}

// Triggered at startup
func ReadyHandler(session *discordgo.Session, event *discordgo.Ready) {
	log.Println("Bot has been successfully launched")
}

// Triggered when the user sends a command
func CommandsHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	// Redirects to the handler of the corresponding command handler
	if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
		handler(session, interaction)
	}
}

// Info timeout command handler
func InfoCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	channelID := interaction.ChannelID

	// Get channelProperties
	channelProperties, err := cpstorage.GetChannelProperties(channelID)

	if err != nil {
		log.Printf("Failed to get timeout %v", err)
		responseToCommand("Failed to get timeout", session, interaction)
	} else if channelProperties == nil {
		responseToCommand("Messages are not deleted in this channel", session, interaction)
	} else {
		responseMessage := fmt.Sprintf("All messages sent more than %s ago will be deleted", сonvertFloatHoursToTimeString(channelProperties.Timeout))
		responseToCommand(responseMessage, session, interaction)

		err := cpstorage.UpdateChannelLastActivity(channelID, time.Now().Unix())
		if err != nil {
			log.Printf("Failed to update channel last activity %v: ", err)
		}
	}
}

// Remove timeout command handler
func RemoveTimeoutCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	channelID := interaction.ChannelID

	responseMessage := "Deleting messages in the channel has been stopped"

	// Remove timeout
	err := cpstorage.DeleteChannelProperties(channelID)
	if err != nil {
		responseMessage = "Failed to stop deletion"
		log.Printf("Failed to stop deletion: %v", err)
	}

	responseToCommand(responseMessage, session, interaction)
}

// Set timeout command handler
func SetTimeoutCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	hours := interaction.ApplicationCommandData().Options[0].FloatValue()
	channelID := interaction.ChannelID

	responseMessage := fmt.Sprintf("All messages sent more than %s ago will be deleted", сonvertFloatHoursToTimeString(hours))

	channelProperties := cpstorage.ChannelPropertiesEntity{
		ChannelID:            channelID,
		Timeout:              hours,
		LastActivityDateUnix: time.Now().Unix(),
		NextRemoveDateUnix:   0, // Channel must be checked now
	}

	// Save channel properties
	err := cpstorage.WriteChannelProperties(&channelProperties)
	if err != nil {
		responseMessage = "Failed to save timeout"
		log.Printf("Failed to save timeout: %v", err)
	}

	responseToCommand(responseMessage, session, interaction)
}

// Universal way to responsd to command
func responseToCommand(message string, session *discordgo.Session, interaction *discordgo.InteractionCreate) (err error) {
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   64, // 64 - Ephemeral messages. These messages are visible only to the user who called the command
		},
	})
	return
}

// Format hours (float) to "Xh Ym" form
func сonvertFloatHoursToTimeString(hours float64) string {
	h := int(math.Floor(hours))
	m := int(math.Round((hours - float64(h)) * 60))
	return fmt.Sprintf("%dh %dm", h, m)
}
