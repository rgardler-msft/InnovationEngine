package commands

import (
	"fmt"

	"github.com/Azure/InnovationEngine/internal/engine/common"
)

var executionRunnerTypes = []string{"bash", "azurecli", "azurecli-interactive", "terraform"}
var inspectRunnerTypes = []string{"bash", "azurecli", "azurecli-inspect", "terraform"}

var commonCreateScenarioFromMarkdown = common.CreateScenarioFromMarkdown

func createScenarioFromOptions(opts *executionOptions, runners []string) (*common.Scenario, error) {
	if opts == nil {
		return nil, fmt.Errorf("execution options are required")
	}
	if len(runners) == 0 {
		runners = executionRunnerTypes
	}
	return commonCreateScenarioFromMarkdown(
		opts.MarkdownPath,
		runners,
		opts.EnvironmentVariables,
	)
}
