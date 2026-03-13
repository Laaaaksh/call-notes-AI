package health

import (
	"context"
)

// IDatabase is the interface for database health checks
type IDatabase interface {
	Ping(ctx context.Context) error
}

var HtModule IModule

var NewModule = func(ctx context.Context, db IDatabase) IModule {
	if HtModule == nil {
		core := NewCore(db)
		HtModule = &Module{Core: core}
	}
	return HtModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore {
	return m.Core
}
