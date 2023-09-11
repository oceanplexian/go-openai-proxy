package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// Reads Data from the Client
func ReadAndUnmarshalBody(ctx context.Context, r *http.Request, w http.ResponseWriter) (RequestData, error) {
	var requestData RequestData

	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		http.Error(w, "Logger not found in context", http.StatusInternalServerError)
		return requestData, fmt.Errorf("Logger not found in context")
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "Error reading request body",
		}
		errorData, _ := json.MarshalIndent(errorResponse, "", "  ")
		http.Error(w, string(errorData), http.StatusInternalServerError)
		logger.Error("Error reading request body", zap.String("raw_body", string(body)), zap.Error(err))
		return requestData, err
	}
	defer r.Body.Close()

	// Check Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "Invalid request format. Please send a JSON payload.",
		}
		errorData, _ := json.MarshalIndent(errorResponse, "", "  ")
		http.Error(w, string(errorData), http.StatusBadRequest)
		logger.Error("Invalid request format", zap.String("raw_body", string(body)))
		return requestData, fmt.Errorf("Invalid request format")
	}

	// Try to unmarshal the body into the requestData struct
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "Error parsing JSON payload",
		}
		errorData, _ := json.MarshalIndent(errorResponse, "", "  ")
		http.Error(w, string(errorData), http.StatusBadRequest)
		logger.Error("Error parsing JSON payload", zap.String("raw_body", string(body)), zap.Error(err))
		return requestData, err
	}

	return requestData, nil
}

func createJsonResponse(content string, isClosing bool) JsonResponse {
	commonFields := JsonResponse{
		ID:      "chatcmpl-1692118020279965440",
		Object:  "chat.completions.chunk",
		Created: 1692118020,
		Model:   "Nous-Hermes-Llama2-GPTQ",
		Usage: map[string]interface{}{
			"prompt_tokens":     58,
			"completion_tokens": 100,
			"total_tokens":      1000,
		},
	}

	if isClosing {
		commonFields.Choices = []Choice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: map[string]string{
					"role":    "assistant",
					"content": "", // Empty Content sent to the Client
				},
				Delta: map[string]string{
					"role":    "assistant",
					"content": "", // Empty Content sent to the Client
				},
			},
		}
	} else {
		commonFields.Choices = []Choice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: map[string]string{
					"role":    "assistant",
					"content": content,
				},
				Delta: map[string]string{
					"role":    "assistant",
					"content": content,
				},
			},
		}
	}

	return commonFields
}

// Create a function to send JsonResponse
func sendJsonResponse(w http.ResponseWriter, resp JsonResponse, flusher http.Flusher) {
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "data: %s\r\n\r\n", data)
	flusher.Flush()
}
