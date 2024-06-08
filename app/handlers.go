package main

// Handlers

import (
	"fmt"
	"log"
	"math"

	"github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		"set-timeout":    SetTimeoutCommandHandler,
		"info-timeout":   InfoCommandHandler,
		"remove-timeout": RemoveTimeoutCommandHandler,
	}
)

// Registers all handlers
func RegisterHandlers() {
	session.AddHandler(CommandsHandler)
	session.AddHandler(ReadyHandler)
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

func InfoCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	channelID := interaction.ChannelID

	// Get timeout
	hasTimeout, timeout, err := GetChannelTimeout(channelID)
	if err != nil {
		log.Printf("Failed to get timeout %v", err)
		responseToCommand("Failed to get timeout", session, interaction)
		return
	}

	if hasTimeout {
		responseMessage := fmt.Sprintf("All messages sent more than %s ago will be deleted", сonvertFloatHoursToTimeString(timeout))
		responseToCommand(responseMessage, session, interaction)
		return
	} else {
		responseToCommand("Messages are not deleted in this channel", session, interaction)
		return
	}
}

func RemoveTimeoutCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	channelID := interaction.ChannelID

	responseMessage := "Deleting messages in the channel has been stopped"

	// Remove timeout
	err := DeleteChannelTimeout(channelID)
	if err != nil {
		responseMessage = "Failed to stop deletion"
		log.Printf("Failed to stop deletion: %v", err)
	}

	responseToCommand(responseMessage, session, interaction)
}

func SetTimeoutCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	hours := interaction.ApplicationCommandData().Options[0].FloatValue()
	channelID := interaction.ChannelID

	responseMessage := fmt.Sprintf("All messages sent more than %s ago will be deleted", сonvertFloatHoursToTimeString(hours))

	// Save channel timeout
	err := WriteChannelTimeout(channelID, hours)
	if err != nil {
		responseMessage = "Failed to save timeout"
		log.Printf("Failed to save timeout: %v", err)
	}

	responseToCommand(responseMessage, session, interaction)
}

func responseToCommand(message string, session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func сonvertFloatHoursToTimeString(hours float64) string {
	h := int(math.Floor(hours))
	m := int(math.Round((hours - float64(h)) * 60))
	return fmt.Sprintf("%dh %dm", h, m)
}
