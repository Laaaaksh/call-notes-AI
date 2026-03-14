package sentiment

import (
	"github.com/call-notes-ai-service/pkg/database"
)

var sentModule IModule

var NewModule = func(pool database.IPool) IModule {
	if sentModule == nil {
		repo := NewRepository(pool)
		core := NewCore(repo)
		sentModule = &Module{Core: core, Repo: repo}
	}
	return sentModule
}

// IModule defines the sentiment module interface
type IModule interface {
	GetCore() ICore
	GetRepository() IRepository
}

// Module implements IModule
type Module struct {
	Core ICore
	Repo IRepository
}

func (m *Module) GetCore() ICore             { return m.Core }
func (m *Module) GetRepository() IRepository { return m.Repo }
