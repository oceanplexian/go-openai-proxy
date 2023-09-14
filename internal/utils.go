package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Define static errors.
var (
	ErrJSONUnmarshalFailed  = fmt.Errorf("json.Unmarshal failed")
	ErrJSONMarshalFailed    = fmt.Errorf("json.MarshalIndent failed")
	ErrLoggerNotFound       = fmt.Errorf("logger not found in context")
	ErrInvalidRequestFormat = fmt.Errorf("invalid request format")
)

// Reads Data from the Client.
func ReadAndUnmarshalBody(cfg *Config, logger *log.Logger, resp http.ResponseWriter, req *http.Request) (RequestData, error) {
	var requestData RequestData

	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "error reading request body",
		}
		err, marshalErr := json.MarshalIndent(errorResponse, "", "  ")

		if err != nil {
			logger.WithFields(log.Fields{"error": marshalErr}).Error("failed to marshal JSON")
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)

			return requestData, fmt.Errorf("json.MarshalIndent failed: %w", marshalErr)
		}

		http.Error(resp, string(err), http.StatusInternalServerError)

		return requestData, fmt.Errorf("io.ReadAll failed: %w", err)
	}
	defer req.Body.Close()

	// Check Content-Type and error if it's not a JSON payload
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "Invalid request format. Please send a JSON payload.",
		}
		errorData, err := json.MarshalIndent(errorResponse, "", "  ")

		if err != nil {
			logger.WithFields(log.Fields{"error": err}).Error("Failed to marshal JSON")
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)

			return requestData, fmt.Errorf("json.MarshalIndent failed: %w", err)
		}

		http.Error(resp, string(errorData), http.StatusBadRequest)

		return requestData, ErrInvalidRequestFormat
	}

	// Try to unmarshal the body into the requestData struct
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		errorResponse := map[string]interface{}{
			"status":  "error",
			"message": "Error parsing JSON payload",
		}
		errorData, marshalErr := json.MarshalIndent(errorResponse, "", "  ")

		if marshalErr != nil {
			logger.WithFields(log.Fields{"error": marshalErr}).Error("Failed to marshal JSON")
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)

			return requestData, fmt.Errorf("%w: %v", ErrJSONMarshalFailed, marshalErr)
		}

		http.Error(resp, string(errorData), http.StatusBadRequest)
		logger.WithFields(log.Fields{"error": err, "body": string(body)}).Error("Error parsing JSON payload")

		return requestData, fmt.Errorf("%w: %v", ErrJSONUnmarshalFailed, err)
	}

	return requestData, nil
}

func createJSONResponse(content string, isClosing bool) JSONResponse {
	commonFields := JSONResponse{
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

// Create a function to send JSONResponse.
func sendJSONResponse(writer http.ResponseWriter, resp JSONResponse, flusher http.Flusher) {
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(writer, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(writer, "data: %s\r\n\r\n", data)
	flusher.Flush()
}
