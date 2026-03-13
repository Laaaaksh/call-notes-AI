package session

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var SessModule IModule

var NewModule = func(ctx context.Context, pool *pgxpool.Pool, redisClient *redis.Client) IModule {
	if SessModule == nil {
		repo := NewRepository(pool)
		core := NewCore(ctx, repo, redisClient)
		handler := NewHTTPHandler(core)
		SessModule = &Module{Core: core, Handler: handler, Repo: repo}
	}
	return SessModule
}

type IModule interface {
	GetCore() ICore
	GetHandler() *HTTPHandler
	GetRepository() IRepository
}

type Module struct {
	Core    ICore
	Handler *HTTPHandler
	Repo    IRepository
}

func (m *Module) GetCore() ICore              { return m.Core }
func (m *Module) GetHandler() *HTTPHandler    { return m.Handler }
func (m *Module) GetRepository() IRepository  { return m.Repo }
