package deepgram

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/gorilla/websocket"
)

var (
	ErrInvalidDeepgramURL   = errors.New("invalid deepgram URL")
	ErrWSConnectFailed      = errors.New("deepgram WebSocket connect failed")
	ErrNoSessionConnection  = errors.New("no connection for session")
)

type TranscriptResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Speaker    int     `json:"speaker"`
	IsFinal    bool    `json:"is_final"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
}

type IClient interface {
	Connect(ctx context.Context, sessionID string) error
	SendAudio(sessionID string, audio []byte) error
	Close(sessionID string) error
	OnTranscript(handler func(sessionID string, result *TranscriptResult))
}

type Client struct {
	cfg         *config.DeepgramConfig
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
	onTranscript func(sessionID string, result *TranscriptResult)
}

func NewClient(cfg *config.DeepgramConfig) IClient {
	return &Client{
		cfg:         cfg,
		connections: make(map[string]*websocket.Conn),
	}
}

func (c *Client) OnTranscript(handler func(sessionID string, result *TranscriptResult)) {
	c.onTranscript = handler
}

func (c *Client) Connect(ctx context.Context, sessionID string) error {
	u, err := url.Parse(c.cfg.WSURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDeepgramURL, err)
	}

	q := u.Query()
	q.Set("model", c.cfg.Model)
	q.Set("language", c.cfg.Language)
	q.Set("smart_format", fmt.Sprintf("%t", c.cfg.SmartFormat))
	q.Set("diarize", fmt.Sprintf("%t", c.cfg.Diarize))
	q.Set("interim_results", fmt.Sprintf("%t", c.cfg.InterimResults))
	u.RawQuery = q.Encode()

	header := make(map[string][]string)
	header["Authorization"] = []string{"Token " + c.cfg.APIKey}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWSConnectFailed, err)
	}

	c.mu.Lock()
	c.connections[sessionID] = conn
	c.mu.Unlock()

	logger.Info(constants.LogMsgDeepgramConnected, constants.LogFieldSessionID, sessionID)

	go c.readLoop(sessionID, conn)

	return nil
}

func (c *Client) SendAudio(sessionID string, audio []byte) error {
	c.mu.RLock()
	conn, ok := c.connections[sessionID]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrNoSessionConnection, sessionID)
	}

	return conn.WriteMessage(websocket.BinaryMessage, audio)
}

func (c *Client) Close(sessionID string) error {
	c.mu.Lock()
	conn, ok := c.connections[sessionID]
	if ok {
		delete(c.connections, sessionID)
	}
	c.mu.Unlock()

	if !ok {
		return nil
	}

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	_ = conn.WriteMessage(websocket.CloseMessage, closeMsg)
	logger.Info(constants.LogMsgDeepgramDisconnected, constants.LogFieldSessionID, sessionID)
	return conn.Close()
}

func (c *Client) readLoop(sessionID string, conn *websocket.Conn) {
	defer func() {
		c.mu.Lock()
		delete(c.connections, sessionID)
		c.mu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			logger.Warn(constants.LogMsgDeepgramReadErr, constants.LogFieldSessionID, sessionID, constants.LogKeyError, err)
			return
		}

		// TODO: Parse Deepgram JSON response into TranscriptResult
		// and call c.onTranscript(sessionID, &result)
	}
}

func (c *Client) ActiveConnections() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.connections)
}
