package commands

import (
	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
)

type engineRunner interface {
	ExecuteScenario(*common.Scenario) error
	TestScenario(*common.Scenario) error
	InteractWithScenario(*common.Scenario) error
}

// engineNewEngine abstracts engine.NewEngine for easier test overrides.
var engineNewEngine = func(cfg engine.EngineConfiguration) (engineRunner, error) {
	return engine.NewEngine(cfg)
}
