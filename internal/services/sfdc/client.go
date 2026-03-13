package sfdc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
)

var (
	ErrCreateAuthRequest = errors.New("failed to create auth request")
	ErrSFAuthFailed      = errors.New("Salesforce auth failed")
	ErrParseAuthResp     = errors.New("failed to parse auth response")
	ErrMarshalSFRecord   = errors.New("failed to marshal SF record")
	ErrCreateSFRequest   = errors.New("failed to create SF request")
	ErrSFUpsertFailed    = errors.New("SF upsert failed after retries")
)

type IClient interface {
	Authenticate(ctx context.Context) error
	UpsertRecord(ctx context.Context, externalIDField, externalIDValue string, fields map[string]string) (string, error)
}

type Client struct {
	cfg         *config.SalesforceConfig
	httpClient  *http.Client
	accessToken string
	instanceURL string
	mu          sync.RWMutex
}

func NewClient(cfg *config.SalesforceConfig) IClient {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.GetTimeout()},
	}
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
}

func (c *Client) Authenticate(ctx context.Context) error {
	data := fmt.Sprintf(
		"grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s",
		c.cfg.ClientID, c.cfg.ClientSecret, c.cfg.Username, c.cfg.Password,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.InstanceURL+"/services/oauth2/token",
		bytes.NewBufferString(data),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreateAuthRequest, err)
	}
	req.Header.Set(constants.HeaderContentType, "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSFAuthFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d, body: %s", ErrSFAuthFailed, resp.StatusCode, string(body))
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("%w: %v", ErrParseAuthResp, err)
	}

	c.mu.Lock()
	c.accessToken = authResp.AccessToken
	c.instanceURL = authResp.InstanceURL
	c.mu.Unlock()

	logger.Info(constants.LogMsgSFAuthenticated, constants.LogFieldInstanceURL, authResp.InstanceURL)
	return nil
}

func (c *Client) UpsertRecord(ctx context.Context, externalIDField, externalIDValue string, fields map[string]string) (string, error) {
	start := time.Now()

	c.mu.RLock()
	token := c.accessToken
	baseURL := c.instanceURL
	c.mu.RUnlock()

	if token == "" {
		if err := c.Authenticate(ctx); err != nil {
			return "", err
		}
		c.mu.RLock()
		token = c.accessToken
		baseURL = c.instanceURL
		c.mu.RUnlock()
	}

	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s/%s",
		baseURL, c.cfg.APIVersion, c.cfg.ObjectName, externalIDField, externalIDValue,
	)

	body, err := json.Marshal(fields)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMarshalSFRecord, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCreateSFRequest, err)
	}
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp *http.Response
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if attempt < c.cfg.MaxRetries {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSFUpsertFailed, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.Authenticate(ctx); err != nil {
			return "", err
		}
		return c.UpsertRecord(ctx, externalIDField, externalIDValue, fields)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%w: status %d, body: %s", ErrSFUpsertFailed, resp.StatusCode, string(respBody))
	}

	logger.Info(constants.LogMsgSFUpsertComplete,
		constants.LogFieldObject, c.cfg.ObjectName,
		constants.LogFieldExternalID, externalIDValue,
		constants.LogFieldLatencyMs, time.Since(start).Milliseconds(),
	)

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.ID != "" {
		return result.ID, nil
	}
	return externalIDValue, nil
}
