package transcription

import "context"

var TxModule IModule

var NewModule = func(ctx context.Context) IModule {
	if TxModule == nil {
		core := NewCore(ctx)
		TxModule = &Module{Core: core}
	}
	return TxModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore { return m.Core }
