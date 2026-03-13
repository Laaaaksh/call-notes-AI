package reasoning

import (
	"context"

	"github.com/call-notes-ai-service/internal/services/llm"
)

var RModule IModule

var NewModule = func(ctx context.Context, llmClient llm.IClient) IModule {
	if RModule == nil {
		core := NewCore(ctx, llmClient)
		RModule = &Module{Core: core}
	}
	return RModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore { return m.Core }
