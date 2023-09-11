package internal

import (
	"context"
	"errors"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"math"
)

func CreateChatCompletionStream(upstreams map[string]Upstream, messages []openai.ChatCompletionMessage, maxTokens int) <-chan string {
	// Find the lowest-priority upstream
	lowestPriority := math.MaxInt
	var selectedUpstream Upstream

	for _, upstream := range upstreams {
		if upstream.Priority < lowestPriority {
			lowestPriority = upstream.Priority
			selectedUpstream = upstream
		}
	}

	// Use the function corresponding to the lowest-priority upstream
	if selectedUpstream.Type == "azure" {
		return CreateAzureChatCompletionStream(selectedUpstream.URL, messages, maxTokens)
	} else if selectedUpstream.Type == "openai" {
		return CreateOpenAIChatCompletionStream(messages, maxTokens)
	} else {
		panic("Invalid upstream type")
	}
}

func CreateOpenAIChatCompletionStream(messages []openai.ChatCompletionMessage, maxTokens int) <-chan string {
	responseChannel := make(chan string)

	go func() {
		defer close(responseChannel)

		// Initialize the OpenAI client
		client := openai.NewClient("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")

		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: maxTokens,
			Messages:  messages,
			Stream:    true, // You can remove this line if you don't want streaming
		}

		stream, err := client.CreateChatCompletionStream(ctx, req) // Or use CreateChatCompletion for non-streaming
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return
			}

			responseChannel <- response.Choices[0].Delta.Content
		}
	}()

	return responseChannel

}

func CreateAzureChatCompletionStream(azureURL string, messages []openai.ChatCompletionMessage, maxTokens int) <-chan string {
	responseChannel := make(chan string)

	go func() {
		defer close(responseChannel)

		// Use azureURL from the function argument
		config := openai.DefaultAzureConfig("dummy", azureURL)
		c := openai.NewClientWithConfig(config)

		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: maxTokens,
			Messages:  messages,
			Stream:    true,
		}

		stream, err := c.CreateChatCompletionStream(ctx, req)
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return
			}

			responseChannel <- response.Choices[0].Delta.Content
		}
	}()

	return responseChannel
}
