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
	BotToken                       string
	GuildID                                = ""
	LogToFIle                              = true
	MaximumHoursValue              float64 = 720  // 1 month
	MinimalHoursValue              float64 = 0.05 // 3 minutes
	RemoveCommandsAfterExit                = true
	RemoveInactiveChannelTimeHorus float64 = 720 // 1 month
	Session                        *discordgo.Session
	SharedDataPath                 = "./data"
)

func loadConfig() {
	BotToken = "Bot " + os.Getenv("DISCORD_OUTDATE_DELETE_BOT_TOKEN")
}

func loadSession() {
	var err error
	Session, err = discordgo.New(BotToken)
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

// TODO: Check channs that really may be have messages to delete
func main() {
	if LogToFIle {
		file := setupLogToFile()
		defer file.Close()
	}

	// Open discord session
	err := Session.Open()
	if err != nil {
		log.Fatal("Error when opening a bot session: ", err)
	}
	defer Session.Close()

	registeredCommands := RegisterCommands()
	if RemoveCommandsAfterExit {
		defer RemoveCommands(registeredCommands)
	}

	go RemoveOldMessages()

	waitForExit()
}
