package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
)

var (
	ErrMarshalLLMRequest = errors.New("failed to marshal LLM request")
	ErrCreateLLMRequest  = errors.New("failed to create LLM request")
	ErrLLMRequestFailed  = errors.New("LLM request failed")
	ErrReadLLMResponse   = errors.New("failed to read LLM response")
	ErrParseLLMResponse  = errors.New("failed to parse LLM response")
	ErrEmptyLLMResponse  = errors.New("empty LLM response")
)

type IClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Client struct {
	cfg        *config.LLMConfig
	httpClient *http.Client
}

func NewClient(cfg *config.LLMConfig) IClient {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.GetTimeout(),
		},
	}
}

type bedrockRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Temperature      float64   `json:"temperature"`
	Messages         []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bedrockResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	start := time.Now()

	reqBody := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        c.cfg.MaxTokens,
		Temperature:      c.cfg.Temperature,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMarshalLLMRequest, err)
	}

	endpoint := fmt.Sprintf(
		"https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke",
		c.cfg.Region, c.cfg.Model,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCreateLLMRequest, err)
	}
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMRequestFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrReadLLMResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d, body: %s", ErrLLMRequestFailed, resp.StatusCode, string(respBody))
	}

	var bedrockResp bedrockResponse
	if err := json.Unmarshal(respBody, &bedrockResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseLLMResponse, err)
	}

	if len(bedrockResp.Content) == 0 {
		return "", ErrEmptyLLMResponse
	}

	logger.Info(constants.LogMsgLLMComplete,
		constants.LogFieldModel, c.cfg.Model,
		constants.LogFieldLatencyMs, time.Since(start).Milliseconds(),
	)

	return bedrockResp.Content[0].Text, nil
}
