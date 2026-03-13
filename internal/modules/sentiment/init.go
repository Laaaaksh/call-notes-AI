package sentiment

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var sentModule IModule

var NewModule = func(pool *pgxpool.Pool) IModule {
	if sentModule == nil {
		repo := NewRepository(pool)
		core := NewCore(repo)
		sentModule = &Module{Core: core, Repo: repo}
	}
	return sentModule
}

type IModule interface {
	GetCore() ICore
	GetRepository() IRepository
}

type Module struct {
	Core ICore
	Repo IRepository
}

func (m *Module) GetCore() ICore             { return m.Core }
func (m *Module) GetRepository() IRepository { return m.Repo }
