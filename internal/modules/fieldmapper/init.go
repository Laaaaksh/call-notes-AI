package fieldmapper

import "context"

var FMModule IModule

var NewModule = func(ctx context.Context) IModule {
	if FMModule == nil {
		core := NewCore(ctx)
		FMModule = &Module{Core: core}
	}
	return FMModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore { return m.Core }
