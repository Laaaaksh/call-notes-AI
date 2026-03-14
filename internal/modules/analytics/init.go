package analytics

import (
	"github.com/call-notes-ai-service/pkg/database"
)

var analyticsModule IModule

var NewModule = func(pool database.IPool) IModule {
	if analyticsModule == nil {
		repo := NewRepository(pool)
		core := NewCore(repo)
		handler := NewHTTPHandler(core)
		analyticsModule = &Module{Core: core, Handler: handler, Repo: repo}
	}
	return analyticsModule
}

// IModule defines the analytics module interface
type IModule interface {
	GetCore() ICore
	GetHandler() *HTTPHandler
	GetRepository() IRepository
}

// Module implements IModule
type Module struct {
	Core    ICore
	Handler *HTTPHandler
	Repo    IRepository
}

func (m *Module) GetCore() ICore             { return m.Core }
func (m *Module) GetHandler() *HTTPHandler   { return m.Handler }
func (m *Module) GetRepository() IRepository { return m.Repo }
