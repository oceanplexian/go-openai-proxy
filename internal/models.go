package internal

import (
	openai "github.com/sashabaranov/go-openai"
)

type Listener struct {
	Interface string `yaml:"interface"`
	Port      string `yaml:"port"`
}

type Upstream struct {
	Type     string `yaml:"type"`
	URL      string `yaml:"url,omitempty"`
	Model    string `yaml:"model"`
	Priority int    `yaml:"priority"`
	APIKey   string `yaml:"apiKey"`
}

type Config struct {
	Upstreams map[string]Upstream `yaml:"upstreams"`
	Listeners []Listener          `yaml:"listeners"`
	CertFile  string              `yaml:"certFile"`
	KeyFile   string              `yaml:"keyFile"`
	UseTLS    bool                `yaml:"useTLS"`
	LogConfig LogConfig           `yaml:"logConfig"`
}

type LogConfig struct {
	LogOutput string `yaml:"logOutput"` // Can be "stdout" or "file"
	Filepath  string `yaml:"filepath"`
	Filesize  int64  `yaml:"filesize"`
	FileCount int    `yaml:"fileCount"`
	LogLevel  string `yaml:"logLevel"` // Log level like "info", "debug", "warn", etc.
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestData struct {
	RequestType string                         `json:"requestType"`
	Prompt      string                         `json:"prompt,omitempty"`
	Model       string                         `json:"model"`
	Temperature float64                        `json:"temperature"`
	MaxTokens   int                            `json:"maxTokens"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
}

type JSONResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Usage   map[string]interface{} `json:"usage"`
	Choices []Choice               `json:"choices"`
}

type Choice struct {
	Index        int               `json:"index"`
	FinishReason string            `json:"finish_reason"`
	Text         string            `json:"text"`
	Message      map[string]string `json:"message"`
	Delta        map[string]string `json:"delta"`
}

type ClosingResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []Choice               `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}
