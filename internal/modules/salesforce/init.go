package salesforce

import (
	"context"

	"github.com/call-notes-ai-service/internal/services/sfdc"
)

var SFModule IModule

var NewModule = func(ctx context.Context, sfdcClient sfdc.IClient) IModule {
	if SFModule == nil {
		core := NewCore(ctx, sfdcClient)
		SFModule = &Module{Core: core}
	}
	return SFModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore { return m.Core }
