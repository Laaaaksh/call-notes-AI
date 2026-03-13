package salesforce

import (
	"context"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/salesforce/entities"
	"github.com/call-notes-ai-service/internal/services/sfdc"
)

type ICore interface {
	UpsertCase(ctx context.Context, req *entities.SFUpsertRequest) (*entities.SFUpsertResponse, error)
}

type Core struct {
	sfdcClient sfdc.IClient
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context, sfdcClient sfdc.IClient) ICore {
	return &Core{sfdcClient: sfdcClient}
}

func (c *Core) UpsertCase(ctx context.Context, req *entities.SFUpsertRequest) (*entities.SFUpsertResponse, error) {
	recordID, err := c.sfdcClient.UpsertRecord(ctx, req.ExternalIDField, req.ExternalIDValue, req.Fields)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgSFUpsertFailed,
			constants.LogKeyError, err,
			constants.LogFieldSessionID, req.SessionID,
		)
		return nil, err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgSFCaseUpserted,
		constants.LogFieldSessionID, req.SessionID,
		constants.LogFieldRecordID, recordID,
	)

	return &entities.SFUpsertResponse{
		RecordID:  recordID,
		IsCreated: true,
	}, nil
}
