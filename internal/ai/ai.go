package ai

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type ChatMessage struct {
	Role    string
	Content string
}

type StreamEvent struct {
	Type    string // "content", "tool_call", "tool_result", "done", "error"
	Tool    string
	Content string
	PromptTokens     int
	CompletionTokens int
}

type StreamCallback func(StreamEvent)

type AskResponse struct {
	Answer          string
	Model           string
	PromptTokens     int
	CompletionTokens int
}

type AIService struct {
	client   *openai.Client
	model    string
	baseURL  string
}

func NewAIService(provider, baseURL, apiKey, model string) *AIService {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
		if provider == "ollama" && !strings.HasSuffix(baseURL, "/v1") {
			config.BaseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
		}
	}
	return &AIService{
		client:  openai.NewClientWithConfig(config),
		model:   model,
		baseURL: baseURL,
	}
}

func (s *AIService) AskStream(ctx context.Context, question string, history []ChatMessage, tools []openai.Tool, cb StreamCallback) error {
	return NewAgent(s.client, s.model).AskStream(ctx, question, history, tools, cb)
}

func (s *AIService) Ask(ctx context.Context, question string, history []ChatMessage, tools []openai.Tool) (*AskResponse, error) {
	return NewAgent(s.client, s.model).Ask(ctx, question, history, tools)
}
