package main

// Commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	minimalHoursValue = 0.05 // 3 minutes
	commands          = []*discordgo.ApplicationCommand{
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
					MinValue:    &minimalHoursValue,
					Required:    true,
				},
			},
		},
	}
)

func RegisterCommands() (registeredCommands []*discordgo.ApplicationCommand) {
	log.Println("Registering commands...")

	for _, command := range commands {
		registered_command, err := session.ApplicationCommandCreate(session.State.User.ID, GuildID, command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", command.Name, err)
		}
		registeredCommands = append(registeredCommands, registered_command)
	}

	return
}

func RemoveCommands(commandsForRemoving []*discordgo.ApplicationCommand) {
	log.Println("Removing commands...")

	for _, command := range commandsForRemoving {
		err := session.ApplicationCommandDelete(session.State.User.ID, GuildID, command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
}
