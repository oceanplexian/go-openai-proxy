package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"

	googlesearch "github.com/rocketlaunchr/google-search"
	log "github.com/sirupsen/logrus"
)

// Define static errors.
var (
	ErrRequestDataNil      = errors.New("requestData is nil")
	ErrRequestDataMessages = errors.New("requestData.Messages is nil")
)

// Interceptor function.
func GoogleSearchInterceptor(ctx context.Context, requestData *RequestData) error {
	if requestData == nil {
		return fmt.Errorf("%w", ErrRequestDataNil)
	}

	if requestData.Messages == nil {
		return fmt.Errorf("%w", ErrRequestDataMessages)
	}

	logger, _ := ctx.Value("logger").(*log.Logger) // Type assertion for Logrus

	for message := len(requestData.Messages) - 1; message >= 0; message-- {
		// Experimental logic for searching Google
		if strings.Contains(requestData.Messages[message].Content, "search google for") {
			index := strings.Index(requestData.Messages[message].Content, "search google for")
			afterSearchPhrase := requestData.Messages[message].Content[index+len("search google for "):]

			startQuoteIndex := strings.Index(afterSearchPhrase, "\"")
			endQuoteIndex := strings.LastIndex(afterSearchPhrase, "\"")

			if startQuoteIndex != -1 && endQuoteIndex != -1 && startQuoteIndex < endQuoteIndex {
				searchQuery := afterSearchPhrase[startQuoteIndex+1 : endQuoteIndex]

				searchResult, err := PerformGoogleSearch(ctx, searchQuery)
				if err == nil {
					logger.WithFields(log.Fields{"result": searchResult}).Info("Google Result Found")

					requestData.Messages[message].Content += "\n\nThe google search results are: ```" +
						searchResult + "``` use them to answer the user's question."
				}

				break
			}
		}
	}

	return nil
}

// Function to perform a Google search using the rocketlaunchr/google-search package.
func PerformGoogleSearch(ctx context.Context, query string) (string, error) {
	results, err := googlesearch.Search(ctx, query) // pass the context instead of nil
	if err != nil {
		return "", fmt.Errorf("failed to perform Google search: %w", err) // Wrap the external error
	}

	// Convert the results to a string
	var stringsBuilder strings.Builder
	for i, result := range results {
		stringsBuilder.WriteString(fmt.Sprintf("%d. %s - %s - %s\n", i+1, result.Title, result.URL, result.Description))

		if i >= 4 { // Limit to top 5 results
			break
		}
	}

	return stringsBuilder.String(), nil
}
