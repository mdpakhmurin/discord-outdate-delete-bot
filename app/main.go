package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	session                 *discordgo.Session
	BotToken                string
	GuildID                 = ""
	RemoveCommandsAfterExit = true
)

func loadConfig() {
	BotToken = "Bot " + os.Getenv("DISCORD_OUTDATE_DELETE_BOT_TOKEN")
}

func loadSession() {
	var err error
	session, err = discordgo.New(BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func init() {
	loadConfig()
	loadSession()
	RegisterHandlers()
}

func waitForExit() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press Ctrl+C to exit")
	<-stop
}

func main() {
	// Open discord session
	err := session.Open()
	if err != nil {
		log.Fatal("Error when opening a bot session: ", err)
	}
	defer session.Close()

	registeredCommands := RegisterCommands()

	go RemoveOldMessages()

	waitForExit()

	if RemoveCommandsAfterExit {
		RemoveCommands(registeredCommands)
	}
}
