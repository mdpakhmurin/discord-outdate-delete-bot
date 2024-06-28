package cfgloader

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// Config структура для хранения конфигурационных параметров
type Config struct {
	BotToken                          string
	GuildID                           string
	IsLogToFile                       bool
	MaximumOutdateHoursValue          float64
	MinimaOutdatelHoursValue          float64
	IsRemoveCommandsAfterExit         bool
	RemoveInactiveChannelTimeoutHours float64
	RemoveBatchSize                   int
	OldDontRemoveTimeoutHours         float64
}

// Load configuration
func LoadConfig(configPath string) (cfg Config, err error) {
	cfgFile, err := ini.Load(configPath)
	if err != nil {
		return
	}

	err = loadBotSection(cfgFile, &cfg)
	loadLogSection(cfgFile, &cfg)
	loadTimeSectiong(cfgFile, &cfg)

	return
}

// Load [Bot] Section
func loadBotSection(cfgFile *ini.File, cfg *Config) (err error) {
	botSection := cfgFile.Section("Bot")

	cfg.IsRemoveCommandsAfterExit, _ = botSection.Key("IsRemoveCommandsAfterExit").Bool()

	botTokenStr := botSection.Key("BotToken").String()
	// check if bot token exists
	if botTokenStr == "" {
		return fmt.Errorf("failed to read BotToken from config")
	}
	cfg.BotToken = "Bot " + botSection.Key("BotToken").String()

	removeBatchSize, err := botSection.Key("RemoveBatchSize").Int()
	if err != nil {
		removeBatchSize = 30
	}
	cfg.RemoveBatchSize = removeBatchSize

	return nil
}

// Load [Log] Section
func loadLogSection(cfgFile *ini.File, cfg *Config) {
	loggingSection := cfgFile.Section("Logging")

	cfg.IsLogToFile, _ = loggingSection.Key("IsLogToFile").Bool()
}

// Load [Time] Section
func loadTimeSectiong(cfgFile *ini.File, cfg *Config) {
	hoursSection := cfgFile.Section("Time")

	var err error

	cfg.RemoveInactiveChannelTimeoutHours, err = hoursSection.Key("RemoveInactiveChannelTimeoutHours").Float64()
	if err != nil {
		cfg.RemoveInactiveChannelTimeoutHours = 8760 // 1 year
	}

	cfg.MaximumOutdateHoursValue, err = hoursSection.Key("MaximumOutdateHoursValue").Float64()
	if err != nil {
		cfg.MaximumOutdateHoursValue = 720 // 1 month
	}

	cfg.MinimaOutdatelHoursValue, err = hoursSection.Key("MinimalOutdateHoursValue").Float64()
	if err != nil {
		cfg.MinimaOutdatelHoursValue = 0.15 // 9 minutes
	}

	cfg.OldDontRemoveTimeoutHours, err = hoursSection.Key("OldDontRemoveTimeoutHours").Float64()
	if err != nil {
		cfg.OldDontRemoveTimeoutHours = 335 // 14 days
	}
}
