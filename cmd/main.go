package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oceanplexian/go-openai-proxy/internal"
	log "github.com/sirupsen/logrus"
)

const (
	ReadTimeout  = 10 * time.Second
	WriteTimeout = 10 * time.Second
)

func main() {
	var configPath, cliListeners, logLevel, certFile, keyFile string

	var useTLS bool

	flag.StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")
	flag.StringVar(&cliListeners, "listeners", "", "Comma-separated list of listeners to override config")
	flag.StringVar(&logLevel, "logLevel", "", "Log level")
	flag.StringVar(&certFile, "certFile", "", "Path to the certificate file")
	flag.StringVar(&keyFile, "keyFile", "", "Path to the key file")
	flag.BoolVar(&useTLS, "useTLS", false, "Whether to use TLS")
	flag.Parse()

	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Couldn't load configuration: ", err)
	}

	cfg.UseTLS = useTLS

	if certFile != "" {
		cfg.CertFile = certFile
	}

	if logLevel != "" {
		cfg.LogConfig.LogLevel = logLevel
	}

	if keyFile != "" {
		cfg.KeyFile = keyFile
	}

	if cliListeners != "" {
		overrideListeners(cfg, cliListeners)
	}

	// Initialize the logger in logger.go
	logger, err := internal.InitializeLogger(&cfg.LogConfig)

	if err != nil {
		fmt.Println("Failed to initialize logger:", err) //nolint
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		internal.Response(cfg, logger, w, r)
	})

	startListeners(cfg, logger)
}

func overrideListeners(cfg *internal.Config, cliListeners string) {
	const ExpectedParts = 2

	cfg.Listeners = []internal.Listener{}

	for _, cliListener := range strings.Split(cliListeners, ",") {
		parts := strings.Split(cliListener, ":")
		isValidFormat := len(parts) == ExpectedParts

		if !isValidFormat {
			log.Error("Invalid listener format, skipping: ", cliListener)
			continue
		}

		newListener := internal.Listener{Interface: parts[0], Port: parts[1]}
		cfg.Listeners = append(cfg.Listeners, newListener)
	}
}

// startListener uses a logger from the context for logging.
func startListener(cfg *internal.Config, logger *log.Logger, listener internal.Listener) {
	address := fmt.Sprintf("%s:%s", listener.Interface, listener.Port)

	defaultTimeout := 10 * time.Second

	server := &http.Server{
		Addr:         address,
		Handler:      nil,
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	}

	if cfg.UseTLS {
		logger.WithFields(log.Fields{"address": address}).Info("Starting TLS listener")
		logger.WithFields(log.Fields{"certFile": cfg.CertFile, "keyFile": cfg.KeyFile}).Info("Loading certificates")

		err := server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			logger.WithFields(log.Fields{"address": address, "error": err}).Error("ListenAndServeTLS")
		}
	} else {
		logger.WithFields(log.Fields{"address": address}).Info("Starting listener")

		err := server.ListenAndServe()
		if err != nil {
			logger.WithFields(log.Fields{"address": address, "error": err}).Error("ListenAndServe")
		}
	}
}

// startListeners starts all listeners and uses the context for logging.
func startListeners(cfg *internal.Config, logger *log.Logger) {
	var listenerWaitGroup sync.WaitGroup
	for _, listener := range cfg.Listeners {
		listenerWaitGroup.Add(1)

		listenerFunc := func(listener internal.Listener) {
			defer listenerWaitGroup.Done()
			startListener(cfg, logger, listener)
		}
		go listenerFunc(listener)
	}

	listenerWaitGroup.Wait()
}
