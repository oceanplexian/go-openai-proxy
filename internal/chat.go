package internal

import (
	"context"
	"errors"
	"io"
	"math"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

// CreateChatCompletionStream creates a chat completion stream based on the given upstreams and messages,
// by Default it will use the upstream with the lowest "priority number" and send requests to that one.
func CreateChatCompletionStream(
	cfg *Config,
	logger *log.Logger,
	upstreams map[string]Upstream,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) (<-chan string, string) {
	lowestPriority := math.MaxInt
	var selectedUpstream Upstream

	var selectedUpstreamName string // Added to keep track of the selected upstream name

	for name, upstream := range cfg.Upstreams { // Note: We're using cfg.Upstreams here
		if upstream.Priority < lowestPriority {
			lowestPriority = upstream.Priority
			selectedUpstream = upstream
			selectedUpstreamName = name // Store the name of the selected upstream
		}
	}

	var channel <-chan string

	switch selectedUpstream.Type {
	case "azure":
		channel = CreateAzureChatCompletionStream(cfg, logger, selectedUpstream.APIKey, selectedUpstream.URL, messages, maxTokens)
	case "openai":
		channel = CreateOpenAIChatCompletionStream(cfg, logger, selectedUpstream.APIKey, messages, maxTokens)
	default:
		logger.Error("Invalid upstream type")
		return nil, ""
	}

	return channel, selectedUpstreamName // Return the channel and the selected upstream name
}

// New version of CreateChatCompletionStream
func CreateOpenAIRequest(
	cfg *Config,
	logger *log.Logger,
	requestType string,
	messages []openai.ChatCompletionMessage,
	prompt string,
	maxTokens int,
) (<-chan string, string) {

	// Determine the appropriate upstream based on priority
	lowestPriority := math.MaxInt
	var selectedUpstream Upstream
	var selectedUpstreamName string

	for name, upstream := range cfg.Upstreams {
		if upstream.Priority < lowestPriority {
			lowestPriority = upstream.Priority
			selectedUpstream = upstream
			selectedUpstreamName = name
		}
	}

	responseChannel := make(chan string)
	go func() {
		defer close(responseChannel)
		logger.WithFields(log.Fields{"selectedUpstreamType": selectedUpstream.Type, "requestType": requestType}).Debug("I'm a bug bug bug")
		switch requestType {
		case "chat":
			if selectedUpstream.Type == "azure" {
				for content := range CreateAzureChatCompletionStream(cfg, logger, selectedUpstream.APIKey, selectedUpstream.URL, messages, maxTokens) {
					responseChannel <- content
				}
			} else if selectedUpstream.Type == "openai" {
				// Assuming you've named the OpenAI chat function "CreateOpenAIChatCompletionStream"
				for content := range CreateOpenAIChatCompletionStream(cfg, logger, selectedUpstream.APIKey, messages, maxTokens) {
					responseChannel <- content
				}
			} else {
				responseChannel <- "Invalid upstream type for chat"
			}

		case "completion":
			if selectedUpstream.Type == "azure" {
				for content := range CreateAzureOpenAICompletion(cfg, logger, selectedUpstream.APIKey, selectedUpstream.URL, prompt, maxTokens) {

					responseChannel <- content
				}
			} else if selectedUpstream.Type == "openai" {
				// Assuming you have a corresponding function for OpenAI completion
				for content := range CreateOpenAICompletion(cfg, logger, selectedUpstream.APIKey, prompt, maxTokens) {
					responseChannel <- content
				}
			} else {
				responseChannel <- "Invalid upstream type for completion"
			}

		default:
			responseChannel <- "Unknown request type"
		}
	}()
	// Return the response channel and the selected upstream name
	return responseChannel, selectedUpstreamName
}

// CreateOpenAIChatCompletionStream creates a chat completion stream using OpenAI.
func CreateOpenAIChatCompletionStream(
	cfg *Config,
	logger *log.Logger,
	apiKey string,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) <-chan string {
	responseChannel := make(chan string)

	go func() {
		defer close(responseChannel)

		client := openai.NewClient(apiKey)

		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:            openai.GPT3Dot5Turbo,
			MaxTokens:        maxTokens,
			Messages:         messages,
			Stream:           true,
			Temperature:      0,
			TopP:             0,
			N:                0,
			Stop:             nil,
			PresencePenalty:  0,
			FrequencyPenalty: 0,
			LogitBias:        nil,
			User:             "",
			Functions:        nil,
			FunctionCall:     nil,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			logger.WithFields(log.Fields{"error": err}).Error("openai api error")

			return
		}

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				logger.WithFields(log.Fields{"error": err}).Error("stream error")

				return
			}

			responseChannel <- response.Choices[0].Delta.Content
		}
	}()

	return responseChannel
}

// CreateAzureChatCompletionStream creates a chat completion stream using Azure.
func CreateAzureChatCompletionStream(
	cfg *Config,
	logger *log.Logger,
	apiKey string,
	azureURL string,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) <-chan string {
	responseChannel := make(chan string)
	go func() {
		defer close(responseChannel)

		config := openai.DefaultAzureConfig(apiKey, azureURL)
		client := openai.NewClientWithConfig(config)

		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:            openai.GPT3Dot5Turbo,
			MaxTokens:        maxTokens,
			Messages:         messages,
			Stream:           true,
			Temperature:      0,
			TopP:             0,
			N:                0,
			Stop:             nil,
			PresencePenalty:  0,
			FrequencyPenalty: 0,
			LogitBias:        nil,
			User:             "",
			Functions:        nil,
			FunctionCall:     nil,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			logger.WithFields(log.Fields{"error": err}).Error("openai api error")
			return
		}

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				logger.WithFields(log.Fields{"error": err}).Error("stream error")

				return
			}

			responseChannel <- response.Choices[0].Delta.Content
		}
	}()

	return responseChannel
}

// CreateAzureOpenAICompletionStream creates an OpenAI completion stream using Azure.
func CreateAzureOpenAICompletion(
	cfg *Config,
	logger *log.Logger,
	apiKey string,
	azureURL string,
	prompt string,
	maxTokens int,
) <-chan string {
	responseChannel := make(chan string)

	logger.WithFields(log.Fields{"qqqqq": prompt, "maxtokens": maxTokens}).Debug("I'm a CreateAzureOpenAICompletion")

	go func() {
		defer close(responseChannel)

		config := openai.DefaultAzureConfig(apiKey, azureURL) // Assuming you have a similar method for this
		client := openai.NewClientWithConfig(config)

		ctx := context.Background()

		req := openai.CompletionRequest{
			Model:     openai.GPT3Ada, // Model used for completion
			MaxTokens: maxTokens,
			Prompt:    prompt,
		}

		resp, err := client.CreateCompletion(ctx, req)
		logger.WithFields(log.Fields{
			"error":    err,
			"response": resp, // log the whole response object
		}).Debug("openai api debug")

		if err != nil {
			logger.WithFields(log.Fields{
				"error":    err,
				"response": resp, // log the whole response object
			}).Error("openai api error")
			return
		}

		// Send the completed text to the response channel
		responseChannel <- resp.Choices[0].Text
	}()

	return responseChannel
}

// CreateOpenAICompletion creates a completion stream using OpenAI (non-Azure).
func CreateOpenAICompletion(
	cfg *Config,
	logger *log.Logger,
	apiKey string,
	prompt string,
	maxTokens int,
) <-chan string {
	responseChannel := make(chan string)

	go func() {
		defer close(responseChannel)

		client := openai.NewClient(apiKey)

		ctx := context.Background()

		req := openai.CompletionRequest{
			Model:     openai.GPT3Dot5Turbo, // Model used for completion
			MaxTokens: maxTokens,
			Prompt:    prompt,
		}

		resp, err := client.CreateCompletion(ctx, req)
		if err != nil {
			logger.WithFields(log.Fields{
				"error":    err,
				"response": resp, // log the whole response object
			}).Error("openai api error")
			return
		}

		// Send the completed text to the response channel
		responseChannel <- resp.Choices[0].Text
	}()

	return responseChannel
}
