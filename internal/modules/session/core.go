package session

//go:generate mockgen -source=core.go -destination=mock/mock_core.go -package=mock

import (
	"context"
	"errors"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/session/entities"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound      = errors.New(entities.ErrMsgSessionNotFound)
	ErrSessionAlreadyActive = errors.New(entities.ErrMsgSessionAlreadyActive)
	ErrInvalidSessionState  = errors.New(entities.ErrMsgInvalidSessionState)
	ErrCallIDRequired       = errors.New(entities.ErrMsgCallIDRequired)
	ErrAgentIDRequired      = errors.New(entities.ErrMsgAgentIDRequired)
)

type ICore interface {
	StartSession(ctx context.Context, req *entities.StartSessionRequest) (*entities.SessionResponse, error)
	EndSession(ctx context.Context, sessionID uuid.UUID) error
	GetSessionState(ctx context.Context, sessionID uuid.UUID) (*entities.SessionState, error)
	UpdateField(ctx context.Context, params *entities.UpdateFieldParams) error
	ApplyAgentOverride(ctx context.Context, sessionID uuid.UUID, fieldName, agentValue string) error
	SubmitSession(ctx context.Context, sessionID uuid.UUID, req *entities.SubmitRequest) (*entities.SubmitResponse, error)
}

type Core struct {
	repo        IRepository
	redisClient *redis.Client
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context, repo IRepository, redisClient *redis.Client) ICore {
	return &Core{repo: repo, redisClient: redisClient}
}

func (c *Core) StartSession(ctx context.Context, req *entities.StartSessionRequest) (*entities.SessionResponse, error) {
	if req.TalkdeskCallID == "" {
		return nil, ErrCallIDRequired
	}
	if req.AgentID == "" {
		return nil, ErrAgentIDRequired
	}

	var phone *string
	if req.PatientPhone != "" {
		phone = &req.PatientPhone
	}

	session := &entities.CallSession{
		ID:             uuid.New(),
		TalkdeskCallID: req.TalkdeskCallID,
		AgentID:        req.AgentID,
		PatientPhone:   phone,
		Status:         constants.SessionStatusCreated,
		StartedAt:      time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := c.repo.CreateSession(ctx, session); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgSessionCreateFailed,
			constants.LogKeyError, err,
			constants.LogFieldCallID, req.TalkdeskCallID,
		)
		return nil, err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgSessionCreated,
		constants.LogFieldSessionID, session.ID.String(),
		constants.LogFieldAgentID, session.AgentID,
		constants.LogFieldCallID, session.TalkdeskCallID,
	)

	return &entities.SessionResponse{
		SessionID: session.ID.String(),
		Status:    session.Status,
	}, nil
}

func (c *Core) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	return c.repo.UpdateSessionStatus(ctx, sessionID, constants.SessionStatusReviewing, &now)
}

func (c *Core) GetSessionState(ctx context.Context, sessionID uuid.UUID) (*entities.SessionState, error) {
	session, err := c.repo.GetSession(ctx, sessionID)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgSessionGetFailed,
			constants.LogKeyError, err,
			constants.LogFieldSessionID, sessionID.String(),
		)
		return nil, ErrSessionNotFound
	}

	fields, err := c.repo.GetLatestFields(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	fieldMap := make(map[string]*entities.FieldState, len(fields))
	for _, f := range fields {
		fieldMap[f.FieldName] = &entities.FieldState{
			Value:      f.FieldValue,
			Confidence: f.Confidence,
			Source:     f.Source,
			Version:    f.Version,
		}
	}

	return &entities.SessionState{
		SessionID: session.ID,
		Status:    session.Status,
		Fields:    fieldMap,
	}, nil
}

func (c *Core) UpdateField(ctx context.Context, params *entities.UpdateFieldParams) error {
	field := &entities.ExtractedField{
		ID:         uuid.New(),
		SessionID:  params.SessionID,
		FieldName:  params.FieldName,
		FieldValue: params.Value,
		Confidence: params.Confidence,
		Source:     params.Source,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := c.repo.UpsertField(ctx, field); err != nil {
		return err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgFieldUpdated,
		constants.LogFieldSessionID, params.SessionID.String(),
		constants.LogFieldFieldName, params.FieldName,
		constants.LogFieldConfidence, params.Confidence,
		constants.LogFieldSource, params.Source,
	)
	return nil
}

func (c *Core) ApplyAgentOverride(ctx context.Context, sessionID uuid.UUID, fieldName, agentValue string) error {
	previousValue := ""
	current, err := c.repo.GetFieldByName(ctx, sessionID, fieldName)
	if err == nil {
		previousValue = current.FieldValue
	}

	override := &entities.AgentOverride{
		ID:         uuid.New(),
		SessionID:  sessionID,
		FieldName:  fieldName,
		AIValue:    previousValue,
		AgentValue: agentValue,
		CreatedAt:  time.Now().UTC(),
	}

	if err := c.repo.CreateOverride(ctx, override); err != nil {
		return err
	}

	return c.UpdateField(ctx, &entities.UpdateFieldParams{
		SessionID:  sessionID,
		FieldName:  fieldName,
		Value:      agentValue,
		Confidence: 1.0,
		Source:     constants.SourceAgentOverride,
	})
}

func (c *Core) SubmitSession(ctx context.Context, sessionID uuid.UUID, req *entities.SubmitRequest) (*entities.SubmitResponse, error) {
	for _, o := range req.Overrides {
		if err := c.ApplyAgentOverride(ctx, sessionID, o.FieldName, o.Value); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	if err := c.repo.UpdateSessionStatus(ctx, sessionID, constants.SessionStatusSubmitted, &now); err != nil {
		return nil, err
	}

	// TODO: trigger Salesforce upsert via Kafka event
	logger.Ctx(ctx).Infow(constants.LogMsgSalesforceSubmitted,
		constants.LogFieldSessionID, sessionID.String(),
	)

	return &entities.SubmitResponse{
		SessionID: sessionID.String(),
		Status:    constants.SessionStatusSubmitted,
	}, nil
}
