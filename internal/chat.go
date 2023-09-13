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
	ctx context.Context,
	upstreams map[string]Upstream,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) (<-chan string, string) { // Modified to return an additional string value
	logger, _ := ctx.Value("logger").(*log.Logger)
	cfg, ok := ctx.Value("config").(*Config)

	if !ok {
		logger.Error("No config found in context")
		return nil, ""
	}

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
		channel = CreateAzureChatCompletionStream(ctx, selectedUpstream.APIKey, selectedUpstream.URL, messages, maxTokens)
	case "openai":
		channel = CreateOpenAIChatCompletionStream(ctx, selectedUpstream.APIKey, messages, maxTokens)
	default:
		logger.Error("Invalid upstream type")
		return nil, ""
	}

	return channel, selectedUpstreamName // Return the channel and the selected upstream name
}

// CreateOpenAIChatCompletionStream creates a chat completion stream using OpenAI.
func CreateOpenAIChatCompletionStream(
	ctx context.Context,
	apiKey string,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) <-chan string {
	logger, _ := ctx.Value("logger").(*log.Logger)

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
	ctx context.Context,
	apiKey string,
	azureURL string,
	messages []openai.ChatCompletionMessage,
	maxTokens int,
) <-chan string {
	responseChannel := make(chan string)

	logger, _ := ctx.Value("logger").(*log.Logger)

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
