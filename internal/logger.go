package internal

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

var (
	ErrInvalidFileSettings = fmt.Errorf("invalid file settings for log output")
	ErrInvalidLogLevel     = fmt.Errorf("failed to parse log level")
	ErrInvalidLogOutput    = fmt.Errorf("invalid log output setting")
)

// InitializeLogger initializes the logger with the given log level and other settings.
func InitializeLogger(cfg *LogConfig) (*log.Logger, error) {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	// Parse log level
	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, ErrInvalidLogLevel
	}

	logger.SetLevel(level)

	switch cfg.LogOutput {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "file":
		if cfg.Filepath == "" || cfg.Filesize == 0 || cfg.FileCount == 0 {
			return nil, ErrInvalidFileSettings
		}

		logger.SetOutput(rotatorr.NewMust(&rotatorr.Config{
			FileSize: cfg.Filesize,
			Filepath: cfg.Filepath,
			Rotatorr: &timerotator.Layout{FileCount: cfg.FileCount},
		}))
	default:
		return nil, ErrInvalidLogOutput
	}

	return logger, nil
}
