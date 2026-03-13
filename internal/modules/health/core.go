package health

import (
	"context"
	"sync/atomic"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
)

type ICore interface {
	RunLivenessCheck(ctx context.Context) (string, int)
	RunReadinessCheck(ctx context.Context) (string, int)
	MarkUnhealthy()
	IsHealthy() bool
}

type Core struct {
	db      IDatabase
	healthy int32
}

var _ ICore = (*Core)(nil)

func NewCore(db IDatabase) ICore {
	return &Core{db: db, healthy: 1}
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (c *Core) RunLivenessCheck(ctx context.Context) (string, int) {
	if !c.IsHealthy() {
		return constants.StatusNotServing, constants.HTTPStatusServiceUnavailable
	}
	return constants.StatusServing, constants.HTTPStatusOK
}

func (c *Core) RunReadinessCheck(ctx context.Context) (string, int) {
	if !c.IsHealthy() {
		return constants.StatusNotServing, constants.HTTPStatusServiceUnavailable
	}
	if c.db != nil {
		if err := c.db.Ping(ctx); err != nil {
			logger.Ctx(ctx).Warnw(constants.LogMsgReadinessCheckFailed, constants.LogKeyError, err)
			return constants.StatusNotServing, constants.HTTPStatusServiceUnavailable
		}
	}
	return constants.StatusServing, constants.HTTPStatusOK
}

func (c *Core) MarkUnhealthy() {
	atomic.StoreInt32(&c.healthy, 0)
	logger.Info(constants.LogMsgServiceMarkedUnhealthy)
}

func (c *Core) IsHealthy() bool {
	return atomic.LoadInt32(&c.healthy) == 1
}
