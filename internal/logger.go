package internal

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// InitializeLogger initializes the logger with the given log level and other settings.
func InitializeLogger(logLevel string) (*log.Logger, error) {
	if logLevel == "" {
		logLevel = "info" // Default log level
	}

	// Create a new instance of the logger
	logger := log.New()

	// Set log output to stdout
	logger.SetOutput(os.Stdout)

	// Set JSON formatter
	logger.SetFormatter(&log.JSONFormatter{
		TimestampFormat:   "",
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		DataKey:           "",
		FieldMap:          nil,
		CallerPrettyfier:  nil,
		PrettyPrint:       false,
	})

	// Parse and set log level
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	logger.SetLevel(level)

	return logger, nil
}
