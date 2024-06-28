package main

// Commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands []*discordgo.ApplicationCommand

func initCommands() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "info-timeout",
			Description: "Shows timeout set in the channel",
		},
		{
			Name:        "remove-timeout",
			Description: "Stop removing message int channel",
		},
		{
			Name:        "set-timeout",
			Description: "Bot deletes messages older than specified hours in the channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "hours",
					Description: "message lifetime in hours",
					MinValue:    &Config.MinimaOutdatelHoursValue,
					MaxValue:    Config.MaximumOutdateHoursValue,
					Required:    true,
				},
			},
		},
	}
}

// Register commands in the bot (and local)
func RegisterCommands() (registeredCommands []*discordgo.ApplicationCommand) {
	initCommands()
	log.Println("Registering commands...")

	for _, command := range commands {
		registered_command, err := Session.ApplicationCommandCreate(Session.State.User.ID, Config.GuildID, command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", command.Name, err)
		}
		registeredCommands = append(registeredCommands, registered_command)
	}

	return
}

// Remove (unregisted) commands in the bot
func RemoveCommands(commandsForRemoving []*discordgo.ApplicationCommand) {
	log.Println("Removing commands...")

	for _, command := range commandsForRemoving {
		err := Session.ApplicationCommandDelete(Session.State.User.ID, Config.GuildID, command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
}
