package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	session                 *discordgo.Session
	BotToken                string
	GuildID                 = ""
	LogToFIle               = true
	RemoveCommandsAfterExit = true
	SharedDataPath          = "./data"
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

func setupLogToFile() (file *os.File) {
	// Creating log file
	file, err := os.OpenFile(SharedDataPath+"/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error setting up log to file: ", err)
	}

	// Creating a MultiWriter that will write to a file and to the console
	multi := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multi)

	return
}

func main() {
	if LogToFIle {
		file := setupLogToFile()
		defer file.Close()
	}

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
