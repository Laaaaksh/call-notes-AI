package transcription

import (
	"context"
	"sync"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/transcription/entities"
)

type ICore interface {
	ProcessChunk(ctx context.Context, chunk *entities.TranscriptChunk) error
	GetTranscript(ctx context.Context, sessionID string) (*entities.FullTranscript, error)
}

type Core struct {
	transcripts map[string]*entities.FullTranscript
	mu          sync.RWMutex
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context) ICore {
	return &Core{
		transcripts: make(map[string]*entities.FullTranscript),
	}
}

func (c *Core) ProcessChunk(ctx context.Context, chunk *entities.TranscriptChunk) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	transcript, ok := c.transcripts[chunk.SessionID]
	if !ok {
		transcript = &entities.FullTranscript{SessionID: chunk.SessionID}
		c.transcripts[chunk.SessionID] = transcript
	}

	transcript.Chunks = append(transcript.Chunks, *chunk)

	logger.Ctx(ctx).Debugw(constants.LogMsgTranscriptChunkProcessed,
		constants.LogFieldSessionID, chunk.SessionID,
		constants.LogFieldSequence, chunk.Sequence,
		constants.LogFieldIsFinal, chunk.IsFinal,
		constants.LogFieldSpeaker, chunk.Speaker,
	)

	return nil
}

func (c *Core) GetTranscript(ctx context.Context, sessionID string) (*entities.FullTranscript, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	transcript, ok := c.transcripts[sessionID]
	if !ok {
		return &entities.FullTranscript{SessionID: sessionID}, nil
	}
	return transcript, nil
}
