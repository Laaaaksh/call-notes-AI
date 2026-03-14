package followup

import (
	"github.com/call-notes-ai-service/pkg/database"
)

var fuModule IModule

var NewModule = func(pool database.IPool) IModule {
	if fuModule == nil {
		repo := NewRepository(pool)
		core := NewCore(repo)
		handler := NewHTTPHandler(core)
		fuModule = &Module{Core: core, Handler: handler, Repo: repo}
	}
	return fuModule
}

// IModule defines the follow-up module interface
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
