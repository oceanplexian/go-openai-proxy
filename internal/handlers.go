package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type RequestInterceptor func(ctx context.Context, requestData *RequestData) error

// You can define a list of interceptors here, if you like
var interceptors = []RequestInterceptor{
	GoogleSearchInterceptor,
	// Add more interceptors here
}

func SetCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers, Authorization")
}

func HandleOptionsRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func Response(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	value := ctx.Value("config")
	fmt.Printf("Debug: Retrieved config from context in handlers.go: %v, Type: %T\n", value, value)

	if value == nil {
		fmt.Println("Debug: Value is nil, key 'config' not found in context.")
		http.Error(w, "Config key not found in context", http.StatusInternalServerError)
		return
	}

	cfg, ok := value.(*Config)
	if !ok {
		fmt.Printf("Debug: Type assertion failed. Expected 'Config', got '%T'\n", value)
		http.Error(w, "Config type assertion failed", http.StatusInternalServerError)
		return
	}

	SetCommonHeaders(w)

	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		err := fmt.Errorf("Logger not found in context or type assertion failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "OPTIONS" {
		HandleOptionsRequest(w)
		return
	}

	requestData, err := ReadAndUnmarshalBody(ctx, r, w)
	if err != nil {
		logger.Error("Error reading or parsing request body", zap.Error(err))
		return
	}

	// Run through interceptors
	for _, interceptor := range interceptors {
		if err := interceptor(ctx, &requestData); err != nil {
			logger.Error("Error running interceptor", zap.Error(err))
			return
		}
	}

	// Create a Response Channel between Client and Upstream
	responseChannel := CreateChatCompletionStream(cfg.Upstreams, requestData.Messages, requestData.MaxTokens)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Iterate over the Upstream Streaming Output, and deliver it to the client
	var accumulatedContents []string
	for content := range responseChannel {
		// Log individual segments
		logger.Debug("JSON Response Segment", zap.String("content", content))
		accumulatedContents = append(accumulatedContents, content)

		resp := createJsonResponse(content, false)
		sendJsonResponse(w, resp, flusher)
	}

	completedResponse := strings.Join(accumulatedContents, "")
	finalContentMap := map[string]interface{}{
		"completedResponse": completedResponse,
		"requestMessages":   requestData.Messages,
	}
	jsonFinalContent, err := json.Marshal(finalContentMap)
	if err != nil {
		logger.Error("Failed to marshal final content to JSON", zap.Error(err))
		return
	}
	logger.Info("JSON Completed Response", zap.String("response", string(jsonFinalContent)))

	closingResp := createJsonResponse("", true)
	closingData, err := json.Marshal(closingResp)
	if err != nil {
		logger.Error("Failed to marshal final content to JSON", zap.Error(err))
		return
	}

	// Some magic text we need to send to close the HTTP stream
	fmt.Fprintf(w, "data: %s\r\n\r\n", closingData)
	flusher.Flush()

	fmt.Fprintf(w, "data: [DONE]\r\n\r\n")
	flusher.Flush() // Ensure all data is sent before closing

}
