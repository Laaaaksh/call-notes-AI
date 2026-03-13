package reasoning

import (
	"context"
	"strings"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/reasoning/entities"
	"github.com/call-notes-ai-service/internal/services/llm"
)

type ICore interface {
	ResolveConflicts(ctx context.Context, req *entities.ReasoningRequest) (*entities.ReasoningResponse, error)
	GenerateSummary(ctx context.Context, sessionID, transcript string) (string, error)
}

type Core struct {
	llmClient llm.IClient
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context, llmClient llm.IClient) ICore {
	return &Core{llmClient: llmClient}
}

func (c *Core) ResolveConflicts(ctx context.Context, req *entities.ReasoningRequest) (*entities.ReasoningResponse, error) {
	prompt := buildConflictResolutionPrompt(req)

	response, err := c.llmClient.Complete(ctx, prompt)
	if err != nil {
		logger.Ctx(ctx).Warnw(constants.LogMsgLLMConflictFailed,
			constants.LogKeyError, err,
			constants.LogFieldSessionID, req.SessionID,
		)
		return &entities.ReasoningResponse{ResolvedFields: req.ExistingFields}, nil
	}

	// TODO: Parse structured LLM response
	_ = response
	return &entities.ReasoningResponse{
		ResolvedFields: req.ExistingFields,
	}, nil
}

func (c *Core) GenerateSummary(ctx context.Context, sessionID, transcript string) (string, error) {
	prompt := buildSummaryPrompt(transcript)

	summary, err := c.llmClient.Complete(ctx, prompt)
	if err != nil {
		logger.Ctx(ctx).Warnw(constants.LogMsgLLMSummaryFailed,
			constants.LogKeyError, err,
			constants.LogFieldSessionID, sessionID,
		)
		return "", err
	}

	return summary, nil
}

func buildConflictResolutionPrompt(req *entities.ReasoningRequest) string {
	var sb strings.Builder
	sb.WriteString("You are a medical data quality assistant.\n")
	sb.WriteString("Given the transcript and extracted fields, resolve any conflicts.\n\n")
	sb.WriteString("TRANSCRIPT:\n" + req.TranscriptText + "\n\n")
	sb.WriteString("CURRENT FIELDS:\n")
	for k, v := range req.ExistingFields {
		sb.WriteString("- " + k + ": " + v + "\n")
	}
	sb.WriteString("\nCONFLICT FIELDS: " + strings.Join(req.ConflictFields, ", ") + "\n")
	sb.WriteString("\nReturn ONLY corrected fields as JSON. No explanation.\n")
	return sb.String()
}

func buildSummaryPrompt(transcript string) string {
	return "Summarize this medical call transcript in 2-3 sentences. Include key symptoms, conditions, and follow-up items:\n\n" + transcript
}
