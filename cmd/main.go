package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/oceanplexian/go-openai-proxy/internal"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {

	// Command-line overrides
	var configPath, cliListeners string
	var certFile, keyFile string
	var logLevel string
	var useTLS bool

	flag.StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")
	flag.StringVar(&cliListeners, "listeners", "", "Comma-separated list of listeners to override config (format: iface:port,iface:port,...)")
	flag.StringVar(&logLevel, "logLevel", "", "Log level (debug, info, warn, error, dpanic, panic, fatal)")
	flag.StringVar(&certFile, "certFile", "", "Path to the certificate file")
	flag.StringVar(&keyFile, "keyFile", "", "Path to the key file")
	flag.BoolVar(&useTLS, "useTLS", true, "Whether to use TLS")

	flag.Parse()

	// Load configuration
	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		panic("Couldn't load configuration")
	}
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if certFile != "" {
		cfg.CertFile = certFile
	}
	if keyFile != "" {
		cfg.KeyFile = keyFile
	}
	// TLS is enabled by default, but can be disabled with a flag for testing
	if !useTLS {
		cfg.UseTLS = useTLS
	}
	fmt.Println("Loaded log level from config:", cfg.LogLevel)

	// Validate upstreams and collect priorities
	prioritySet := make(map[int]bool)
	for name, upstream := range cfg.Upstreams {
		if upstream.Type != "azure" && upstream.Type != "openai" {
			fmt.Printf("Invalid type for upstream %s: %s\n", name, upstream.Type)
			return
		}
		if upstream.Type == "azure" && upstream.URL == "" {
			fmt.Printf("Missing URL for Azure upstream %s\n", name)
			return
		}
		if upstream.Type == "openai" && upstream.URL != "" {
			fmt.Printf("URL should not be provided for OpenAI upstream %s\n", name)
			return
		}
		if upstream.Priority <= 0 {
			fmt.Printf("Invalid priority for upstream %s: %d\n", name, upstream.Priority)
			return
		}
		if _, exists := prioritySet[upstream.Priority]; exists {
			fmt.Printf("Duplicate priority for upstream %s: %d\n", name, upstream.Priority)
			return
		}
		prioritySet[upstream.Priority] = true
	}

	// If listeners are provided via CLI, override the ones from the config
	if cliListeners != "" {
		cfg.Listeners = []internal.Listener{}
		for _, cliListener := range strings.Split(cliListeners, ",") {
			parts := strings.Split(cliListener, ":")
			if len(parts) != 2 {
				fmt.Println("Invalid listener format, skipping:", cliListener)
				continue
			}
			cfg.Listeners = append(cfg.Listeners, internal.Listener{Interface: parts[0], Port: parts[1]})
		}
	}

	// Initialize logger
	logger, err := internal.InitializeLogger(cfg.LogLevel)
	if err != nil {
		panic("Couldn't initialize logger")
	}
	defer logger.Sync()

	// Create context with logger and config
	ctx := context.WithValue(context.Background(), "logger", logger)
	ctx = context.WithValue(ctx, "config", cfg)
	logger.Debug("Added config to context: ", zap.Any("config", ctx.Value("config")))

	// Register HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		internal.Response(ctx, w, r)
	})

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Could not get hostname", zap.Error(err))
	} else {
		logger.Info("Hostname", zap.String("hostname", hostname))
	}

	// Start listeners based on configuration
	var wg sync.WaitGroup
	for _, listener := range cfg.Listeners {
		wg.Add(1)
		go func(listener internal.Listener) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%s", listener.Interface, listener.Port)
			logger.Info("Starting listener", zap.String("address", address))

			if cfg.UseTLS {
				err := http.ListenAndServeTLS(address, cfg.CertFile, cfg.KeyFile, nil)
				if err != nil {
					logger.Error("ListenAndServeTLS: ", zap.String("address", address), zap.Error(err))
				}
			} else {
				err := http.ListenAndServe(address, nil)
				if err != nil {
					logger.Error("ListenAndServe: ", zap.String("address", address), zap.Error(err))
				}
			}
		}(listener)
	}

	wg.Wait()
}
