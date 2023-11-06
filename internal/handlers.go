package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type RequestInterceptor func(cfg *Config, logger *log.Logger, requestData *RequestData) error

// You can define a list of interceptors here, if you like.
var interceptors = []RequestInterceptor{
	GoogleSearchInterceptor,
	// Add more interceptors here
}

// Prevent Content-Security-Policy Errors when used with webapps served from a different domain.
func SetCommonHeaders(w http.ResponseWriter, contentType string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	w.Header().Set("Access-Control-Allow-Headers", `
	Origin, 
	Accept, 
	X-Requested-With, 
	Content-Type, 
	Access-Control-Request-Method, 
	Access-Control-Request-Headers, 
	Authorization`)
}

func HandleOptionsRequest(w http.ResponseWriter) {
	SetCommonHeaders(w, "application/json") // Use the refactored function
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))

	if err != nil {
		return
	}
}

// New function to handle different request types
func handleRequestType(
	cfg *Config,
	logger *log.Logger,
	writer http.ResponseWriter,
	request *http.Request,
	requestData RequestData,
) {
	switch {
	case strings.HasSuffix(request.URL.Path, "/chat/completions"):
		requestData.RequestType = "chat"
		handleChatCompletion(cfg, logger, writer, request, requestData)
	case strings.HasSuffix(request.URL.Path, "/completions"):
		requestData.RequestType = "completion"
		handleTextCompletion(cfg, logger, writer, request, requestData)
	default:
		http.Error(writer, "Unknown endpoint", http.StatusNotFound)
	}
}

func Response(
	cfg *Config,
	logger *log.Logger,
	w http.ResponseWriter,
	r *http.Request,
) {
	SetCommonHeaders(w, "text/event-stream") // Stream is the default assumed

	if r.Method == "OPTIONS" {
		HandleOptionsRequest(w)
		return
	}

	requestData, err := ReadAndUnmarshalBody(cfg, logger, w, r)
	if err != nil {
		handleError(w, logger, err, "Error reading or parsing request body")
		return
	}

	// Determine and handle the request type
	handleRequestType(cfg, logger, w, r, requestData)

}

// HandleChatCompletion handles the logic specific to chat completions.
func handleChatCompletion(cfg *Config, logger *log.Logger, w http.ResponseWriter, r *http.Request, requestData RequestData) {
	responseChannel, upstreamName := CreateOpenAIRequest(cfg, logger, requestData.RequestType, requestData.Messages, requestData.Prompt, requestData.MaxTokens)
	sendResponseFromChannel(w, responseChannel, upstreamName, logger, "chat", requestData)
}

// HandleTextCompletion handles the logic specific to text completions.
func handleTextCompletion(cfg *Config, logger *log.Logger, w http.ResponseWriter, r *http.Request, requestData RequestData) {
	responseChannel, upstreamName := CreateOpenAIRequest(cfg, logger, requestData.RequestType, requestData.Messages, requestData.Prompt, requestData.MaxTokens)
	sendResponseFromChannel(w, responseChannel, upstreamName, logger, "completion", requestData)
}

// sendResponseFromChannel handles sending the response to the client from the response channel.
func sendResponseFromChannel(w http.ResponseWriter, responseChannel <-chan string, upstreamName string, logger *log.Logger, requestType string, requestData RequestData) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		handleError(w, logger, errors.New("streaming not supported"), "Streaming not supported")
		return
	}

	var accumulatedContents []string
	for content := range responseChannel {
		accumulatedContents = append(accumulatedContents, content)
		responseType := getResponseType(requestType)
		resp := createJSONResponse(content, responseType, false)
		sendJSONResponse(w, resp, flusher, requestType)
	}

	// After the channel is closed, send the final response.
	sendFinalResponse(w, accumulatedContents, upstreamName, logger, requestType, flusher, requestData)
}

// sendFinalResponse sends the final response after all the streaming content has been sent.
func sendFinalResponse(w http.ResponseWriter, accumulatedContents []string, upstreamName string, logger *log.Logger, requestType string, flusher http.Flusher, requestData RequestData) {
	completedResponse := strings.Join(accumulatedContents, "")
	finalContentMap := map[string]interface{}{
		"completedResponse": completedResponse,
		"requestMessages":   requestData.Messages,
	}
	jsonFinalContent, err := json.Marshal(finalContentMap)
	if err != nil {
		handleError(w, logger, err, "Failed to marshal final content to JSON")
		return
	}

	logger.WithFields(log.Fields{
		"response":     string(jsonFinalContent),
		"upstreamName": upstreamName,
	}).Info("JSON Completed Response")

	if requestType == "chat" {
		closingResp := createJSONResponse("", "chat.completion", true)
		closingData, err := json.Marshal(closingResp)
		if err != nil {
			handleError(w, logger, err, "Failed to marshal final content to JSON")
			return
		}
		fmt.Fprintf(w, "data: %s\r\n\r\n", closingData)
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\r\n\r\n")
	}
	flusher.Flush() // Ensure all data is sent before closing
}

// getResponseType determines the response type based on the request type.
func getResponseType(requestType string) string {
	if requestType == "chat" {
		return "chat.completion"
	}
	return "text_completion"
}

// New function to handle errors consistently
func handleError(w http.ResponseWriter, logger *log.Logger, err error, contextMessage string) {
	logger.WithFields(log.Fields{"error": err}).Error(contextMessage)
	http.Error(w, contextMessage, http.StatusInternalServerError)
}
