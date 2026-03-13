package deepgram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/gorilla/websocket"
)

var (
	ErrInvalidDeepgramURL  = errors.New("invalid deepgram URL")
	ErrWSConnectFailed     = errors.New("deepgram WebSocket connect failed")
	ErrNoSessionConnection = errors.New("no connection for session")
	ErrMaxReconnects       = errors.New("max reconnect attempts exceeded")
)

const (
	maxReconnectAttempts = 3
	baseReconnectDelay  = 100 * time.Millisecond
	maxReconnectDelay   = 5 * time.Second
	audioBufferSize     = 4096
	pongWaitDuration    = 10 * time.Second
	pingInterval        = 5 * time.Second
)

type TranscriptResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Speaker    int     `json:"speaker"`
	IsFinal    bool    `json:"is_final"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
}

type deepgramResponse struct {
	Channel struct {
		Alternatives []struct {
			Transcript string  `json:"transcript"`
			Confidence float64 `json:"confidence"`
			Words      []struct {
				Word       string  `json:"word"`
				Start      float64 `json:"start"`
				End        float64 `json:"end"`
				Confidence float64 `json:"confidence"`
				Speaker    int     `json:"speaker"`
			} `json:"words"`
		} `json:"alternatives"`
	} `json:"channel"`
	IsFinal bool    `json:"is_final"`
	Start   float64 `json:"start"`
	End     float64 `json:"duration"`
}

type sessionConn struct {
	conn        *websocket.Conn
	audioBuffer chan []byte
	cancel      context.CancelFunc
}

type IClient interface {
	Connect(ctx context.Context, sessionID string) error
	SendAudio(sessionID string, audio []byte) error
	Close(sessionID string) error
	OnTranscript(handler func(sessionID string, result *TranscriptResult))
}

type Client struct {
	cfg          *config.DeepgramConfig
	sessions     map[string]*sessionConn
	mu           sync.RWMutex
	onTranscript func(sessionID string, result *TranscriptResult)
}

func NewClient(cfg *config.DeepgramConfig) IClient {
	return &Client{
		cfg:      cfg,
		sessions: make(map[string]*sessionConn),
	}
}

func (c *Client) OnTranscript(handler func(sessionID string, result *TranscriptResult)) {
	c.onTranscript = handler
}

func (c *Client) Connect(ctx context.Context, sessionID string) error {
	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}

	connCtx, cancel := context.WithCancel(ctx)
	sc := &sessionConn{
		conn:        conn,
		audioBuffer: make(chan []byte, audioBufferSize),
		cancel:      cancel,
	}

	c.mu.Lock()
	c.sessions[sessionID] = sc
	c.mu.Unlock()

	logger.Info(constants.LogMsgDeepgramConnected, constants.LogFieldSessionID, sessionID)

	go c.readLoop(connCtx, sessionID, sc)
	go c.writeLoop(connCtx, sessionID, sc)
	go c.pingLoop(connCtx, sessionID, sc)

	return nil
}

func (c *Client) dial(ctx context.Context) (*websocket.Conn, error) {
	u, err := url.Parse(c.cfg.WSURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDeepgramURL, err)
	}

	q := u.Query()
	q.Set("model", c.cfg.Model)
	q.Set("language", c.cfg.Language)
	q.Set("smart_format", fmt.Sprintf("%t", c.cfg.SmartFormat))
	q.Set("diarize", fmt.Sprintf("%t", c.cfg.Diarize))
	q.Set("interim_results", fmt.Sprintf("%t", c.cfg.InterimResults))
	q.Set("redact", "pci")
	q.Set("redact", "ssn")
	u.RawQuery = q.Encode()

	header := make(map[string][]string)
	header["Authorization"] = []string{"Token " + c.cfg.APIKey}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWSConnectFailed, err)
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWaitDuration))
	})

	return conn, nil
}

func (c *Client) reconnect(ctx context.Context, sessionID string, sc *sessionConn) bool {
	for attempt := 1; attempt <= maxReconnectAttempts; attempt++ {
		delay := time.Duration(math.Min(
			float64(baseReconnectDelay)*math.Pow(2, float64(attempt-1)),
			float64(maxReconnectDelay),
		))

		logger.Warn("Deepgram reconnecting",
			constants.LogFieldSessionID, sessionID,
			constants.LogFieldAttempt, attempt,
			constants.LogFieldBackoff, delay.String(),
		)

		select {
		case <-ctx.Done():
			return false
		case <-time.After(delay):
		}

		conn, err := c.dial(ctx)
		if err != nil {
			logger.Warn("Deepgram reconnect failed",
				constants.LogFieldSessionID, sessionID,
				constants.LogFieldAttempt, attempt,
				constants.LogKeyError, err,
			)
			continue
		}

		c.mu.Lock()
		sc.conn = conn
		c.mu.Unlock()

		logger.Info("Deepgram reconnected",
			constants.LogFieldSessionID, sessionID,
			constants.LogFieldAttempt, attempt,
		)
		return true
	}
	return false
}

func (c *Client) SendAudio(sessionID string, audio []byte) error {
	c.mu.RLock()
	sc, ok := c.sessions[sessionID]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrNoSessionConnection, sessionID)
	}

	select {
	case sc.audioBuffer <- audio:
	default:
		logger.Warn("Audio buffer full, dropping chunk",
			constants.LogFieldSessionID, sessionID,
		)
	}
	return nil
}

func (c *Client) Close(sessionID string) error {
	c.mu.Lock()
	sc, ok := c.sessions[sessionID]
	if ok {
		delete(c.sessions, sessionID)
	}
	c.mu.Unlock()

	if !ok {
		return nil
	}

	sc.cancel()
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	_ = sc.conn.WriteMessage(websocket.CloseMessage, closeMsg)
	logger.Info(constants.LogMsgDeepgramDisconnected, constants.LogFieldSessionID, sessionID)
	return sc.conn.Close()
}

func (c *Client) writeLoop(ctx context.Context, sessionID string, sc *sessionConn) {
	for {
		select {
		case <-ctx.Done():
			return
		case audio, ok := <-sc.audioBuffer:
			if !ok {
				return
			}
			c.mu.RLock()
			conn := sc.conn
			c.mu.RUnlock()

			if err := conn.WriteMessage(websocket.BinaryMessage, audio); err != nil {
				logger.Warn("Deepgram write error",
					constants.LogFieldSessionID, sessionID,
					constants.LogKeyError, err,
				)
			}
		}
	}
}

func (c *Client) readLoop(ctx context.Context, sessionID string, sc *sessionConn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := sc.conn
		c.mu.RUnlock()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Warn(constants.LogMsgDeepgramReadErr,
				constants.LogFieldSessionID, sessionID,
				constants.LogKeyError, err,
			)

			if ctx.Err() != nil {
				return
			}

			if !c.reconnect(ctx, sessionID, sc) {
				logger.Error("Deepgram reconnect exhausted, session degraded",
					constants.LogFieldSessionID, sessionID,
				)
				return
			}
			continue
		}

		var resp deepgramResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue
		}

		if len(resp.Channel.Alternatives) == 0 {
			continue
		}

		alt := resp.Channel.Alternatives[0]
		if alt.Transcript == "" {
			continue
		}

		speaker := 0
		if len(alt.Words) > 0 {
			speaker = alt.Words[0].Speaker
		}

		if c.onTranscript != nil {
			c.onTranscript(sessionID, &TranscriptResult{
				Text:       alt.Transcript,
				Confidence: alt.Confidence,
				Speaker:    speaker,
				IsFinal:    resp.IsFinal,
				Start:      resp.Start,
				End:        resp.End,
			})
		}
	}
}

func (c *Client) pingLoop(ctx context.Context, sessionID string, sc *sessionConn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := sc.conn
			c.mu.RUnlock()

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Warn("Deepgram ping failed",
					constants.LogFieldSessionID, sessionID,
					constants.LogKeyError, err,
				)
			}
		}
	}
}

func (c *Client) ActiveConnections() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sessions)
}
