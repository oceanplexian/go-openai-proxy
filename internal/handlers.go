package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type RequestInterceptor func(ctx context.Context, requestData *RequestData) error

// You can define a list of interceptors here, if you like.
var interceptors = []RequestInterceptor{
	GoogleSearchInterceptor,
	// Add more interceptors here
}

// Prevent Content-Security-Policy Errors when used with webapps served from a different domain.
func SetCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))

	if err != nil {
		return
	}
}

func Response(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cfg, _ := ctx.Value("config").(*Config)
	logger, _ := ctx.Value("logger").(*log.Logger) // Type assertion for Logrus

	SetCommonHeaders(w)

	if r.Method == "OPTIONS" {
		HandleOptionsRequest(w)
		return
	}

	requestData, err := ReadAndUnmarshalBody(ctx, r, w)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Error reading or parsing request body")
		return
	}

	// Run through interceptors
	for _, interceptor := range interceptors {
		if err := interceptor(ctx, &requestData); err != nil {
			logger.WithFields(log.Fields{"error": err}).Error("Error running interceptor")
			return
		}
	}
	// Create a Response Channel between Client and Upstream
	// Here, upstreamName would be a string that you capture once at the start,
	// assuming all segments in the conversation are from the same upstream

	responseChannel, upstreamName := CreateChatCompletionStream(ctx, cfg.Upstreams, requestData.Messages, requestData.MaxTokens)
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	flusher.Flush()

	// Iterate over the Upstream Streaming Output, and deliver it to the client
	var accumulatedContents []string

	for content := range responseChannel {
		// Log individual segments along with upstream name
		logger.WithFields(log.Fields{
			"content":      content,
			"upstreamName": upstreamName, // Add upstream name here
		}).Debug("JSON Response Segment")

		accumulatedContents = append(accumulatedContents, content)

		resp := createJSONResponse(content, false)
		sendJSONResponse(w, resp, flusher)
	}

	completedResponse := strings.Join(accumulatedContents, "")
	finalContentMap := map[string]interface{}{
		"completedResponse": completedResponse,
		"requestMessages":   requestData.Messages,
	}
	jsonFinalContent, err := json.Marshal(finalContentMap)

	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Failed to marshal final content to JSON")
		return
	}

	logger.WithFields(log.Fields{
		"response":     string(jsonFinalContent),
		"upstreamName": upstreamName, // Add upstream name here
	}).Info("JSON Completed Response")

	closingResp := createJSONResponse("", true)
	closingData, err := json.Marshal(closingResp)

	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Failed to marshal final content to JSON")
		return
	}

	// Some magic text we need to send to close the HTTP stream
	fmt.Fprintf(w, "data: %s\r\n\r\n", closingData)
	flusher.Flush()

	fmt.Fprintf(w, "data: [DONE]\r\n\r\n")
	flusher.Flush() // Ensure all data is sent before closing
}
