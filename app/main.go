package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/cfgloader"
	"github.com/mdpakhmurin/discord-outdate-delete-bot/cpstorage"
)

var (
	Session        *discordgo.Session
	SharedDataPath = "./data"
	Config         cfgloader.Config
)

func init() {
	loadConfig()
	loadSession()
	RegisterHandlers()
}

// Load gloabal config
func loadConfig() {
	var err error

	// Try to load config
	Config, err = cfgloader.LoadConfig(SharedDataPath + "/config.ini")
	if err != nil {
		log.Fatalf("Failed to load config, %v", err)
	}

	// Beauty print loaded config
	cfgJSON, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	log.Printf("Config loaded: %s\n", cfgJSON)
}

// Conntect to bot. Load session
func loadSession() {
	var err error
	Session, err = discordgo.New(Config.BotToken)
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

	if Config.IsLogToFile {
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
	if Config.IsRemoveCommandsAfterExit {
		defer RemoveCommands(registeredCommands)
	}

	go RemoveOldMessages()

	waitForExit()
}
