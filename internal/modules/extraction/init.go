package extraction

import (
	"context"

	"github.com/call-notes-ai-service/internal/services/llm"
)

var ExModule IModule

var NewModule = func(ctx context.Context, llmClient llm.IClient) IModule {
	if ExModule == nil {
		ruleEngine := NewRuleEngine()
		reasoner := NewLLMReasoner(llmClient)
		piiRedactor := NewPIIRedactor()
		core := NewCore(ctx, ruleEngine, reasoner, piiRedactor)
		ExModule = &Module{Core: core}
	}
	return ExModule
}

type IModule interface {
	GetCore() ICore
}

type Module struct {
	Core ICore
}

func (m *Module) GetCore() ICore { return m.Core }
