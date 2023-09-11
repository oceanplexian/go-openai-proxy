package internal

import (
	"context"
	"fmt"
	"go.uber.org/zap"

	"github.com/rocketlaunchr/google-search"
	"strings"
)

// Interceptor function
func GoogleSearchInterceptor(ctx context.Context, requestData *RequestData) error {
	if requestData == nil {
		return fmt.Errorf("requestData is nil")
	}
	if requestData.Messages == nil {
		return fmt.Errorf("requestData.Messages is nil")
	}

	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		return fmt.Errorf("Logger not found in context")
	}
	for i := len(requestData.Messages) - 1; i >= 0; i-- {
		// Experimental logic for for searching Google
		if strings.Contains(requestData.Messages[i].Content, "search google for") {
			index := strings.Index(requestData.Messages[i].Content, "search google for")
			afterSearchPhrase := requestData.Messages[i].Content[index+len("search google for "):]

			startQuoteIndex := strings.Index(afterSearchPhrase, "\"")
			endQuoteIndex := strings.LastIndex(afterSearchPhrase, "\"")

			if startQuoteIndex != -1 && endQuoteIndex != -1 && startQuoteIndex < endQuoteIndex {
				searchQuery := afterSearchPhrase[startQuoteIndex+1 : endQuoteIndex]

				searchResult, err := PerformGoogleSearch(searchQuery)
				if err == nil {
					logger.Info("Google Result", zap.String("result", string(searchResult)))
					requestData.Messages[i].Content += "\n\nThe google search results are: ```" + searchResult + "``` use them to answer the user's question."
				}
				break
			}
		}
		// Finish experimental logic
	}
	return nil
}

// Function to perform a Google search using the rocketlaunchr/google-search package
func PerformGoogleSearch(query string) (string, error) {
	// Perform the Google search and return the result
	results, err := googlesearch.Search(nil, query)
	if err != nil {
		return "", err
	}

	// Convert the results to a string
	var sb strings.Builder
	for i, result := range results {
		sb.WriteString(fmt.Sprintf("%d. %s - %s - %s\n", i+1, result.Title, result.URL, result.Description))
		if i >= 4 { // Limit to top 5 results
			break
		}
	}

	return sb.String(), nil
}
