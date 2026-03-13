package prediction

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var predModule IModule

var NewModule = func(pool *pgxpool.Pool) IModule {
	if predModule == nil {
		repo := NewRepository(pool)
		core := NewCore(repo)
		handler := NewHTTPHandler(core)
		predModule = &Module{Core: core, Handler: handler, Repo: repo}
	}
	return predModule
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

func (m *Module) GetCore() ICore             { return m.Core }
func (m *Module) GetHandler() *HTTPHandler   { return m.Handler }
func (m *Module) GetRepository() IRepository { return m.Repo }
