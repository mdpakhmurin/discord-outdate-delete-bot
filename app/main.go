package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/data/cpstorage"
)

var (
	BotToken                          string
	GuildID                                   = ""
	IsLogToFIle                               = true
	MaximumHoursValue                 float64 = 720  // 1 month
	MinimalHoursValue                 float64 = 0.05 // 3 minutes
	IsRemoveCommandsAfterExit                 = true
	RemoveInactiveChannelTimeoutHorus float64 = 720 // 1 month
	Session                           *discordgo.Session
	SharedDataPath                    = "./data"
)

func init() {
	loadConfig()
	loadSession()
	RegisterHandlers()
}

// Load configuration
func loadConfig() {
	BotToken = "Bot " + os.Getenv("DISCORD_OUTDATE_DELETE_BOT_TOKEN")
}

// Conntect to bot. Load session
func loadSession() {
	var err error
	Session, err = discordgo.New(BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

// Wait ctrl+c to exit
func waitForExit() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press Ctrl+C to exit")
	<-stop
}

// Set up logging to file
func setupLogToFile(logFilePath string) (file *os.File) {
	// Creating log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error setting up log to file: ", err)
	}

	// Creating a MultiWriter that will write to a file and to the console
	multi := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multi)

	return
}

func main() {
	cpstorage.Init(SharedDataPath)

	if IsLogToFIle {
		file := setupLogToFile(SharedDataPath + "/log.txt")
		defer file.Close()
	}

	// Open discord session
	err := Session.Open()
	if err != nil {
		log.Fatal("Error when opening a bot session: ", err)
	}
	defer Session.Close()

	registeredCommands := RegisterCommands()
	if IsRemoveCommandsAfterExit {
		defer RemoveCommands(registeredCommands)
	}

	go RemoveOldMessages()

	waitForExit()
}
