package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// Each model provider defines their own individual request and response formats.
// For the format, ranges, and default values for the different models, refer to:
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html

const TYPE_TEXT = "text"

const CACHE_CONTROL_EPHEMERAL = "ephemeral"

type CacheControlBlock struct {
	Type string `json:"type"`
}

type ContentBlock struct {
	Type         string             `json:"type"`
	Text         *string            `json:"text,omitempty"`
	CacheControl *CacheControlBlock `json:"cache_control,omitempty"`
}

const ROLE_USER = "user"
const ROLE_SYSTEM = "system"
const ROLE_ASSISTANT = "assistant"

type ClaudeMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
	// Omitting optional request parameters
}

func NewClaudeUserMessage(text string, cachepoint bool) *ClaudeMessage {
	contentBlock := ContentBlock{
		Type: TYPE_TEXT,
		Text: &text,
	}
	if cachepoint {
		contentBlock.CacheControl = &CacheControlBlock{
			Type: CACHE_CONTROL_EPHEMERAL,
		}
	}
	return &ClaudeMessage{
		Role:    ROLE_USER,
		Content: []ContentBlock{contentBlock},
	}
}

func NewClaudeAssistantMessage(text string) *ClaudeMessage {
	contentBlock := ContentBlock{
		Type: TYPE_TEXT,
		Text: &text,
	}
	return &ClaudeMessage{
		Role:    ROLE_ASSISTANT,
		Content: []ContentBlock{contentBlock},
	}
}

type ClaudeRequest struct {
	AnthropicVersion  string           `json:"anthropic_version,omitempty"`
	Messages          []*ClaudeMessage `json:"messages"`
	MaxTokensToSample *int             `json:"max_tokens,omitempty"`
	Temperature       *float32         `json:"temperature,omitempty"`
	// Omitting optional request parameters
}

type ClaudeMessageResponse struct {
	Model        string         `json:"model"`
	Id           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        ClaudeUsage    `json:"usage"`
}

type ClaudeUsage struct {
	InputTokens              int                 `json:"input_tokens"`
	CacheCreationInputTokens int                 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int                 `json:"cache_read_input_tokens"`
	CacheCreation            ClaudeCacheCreation `json:"cache_creation"`
	OutputTokens             int                 `json:"output_tokens"`
}

type ClaudeCacheCreation struct {
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
}

func SendClaudeRequest(ctx context.Context, brClient *bedrockruntime.Client, modelId *string, request *ClaudeRequest) (*ClaudeMessageResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := brClient.InvokeModel(ctx,
		&bedrockruntime.InvokeModelInput{
			ModelId: modelId, // Model ID should be set in the client configuration
			Body:    body,    // Request body should be set here
		},
	)

	if err != nil {
		return nil, err
	}

	var messageResponse ClaudeMessageResponse
	err = json.Unmarshal(response.Body, &messageResponse)
	if err != nil {
		return nil, err
	}

	return &messageResponse, nil
}
