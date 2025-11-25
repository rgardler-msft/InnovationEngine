package commands

import (
	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/spf13/cobra"
)

// / Register the command with our command runner.
func init() {
	rootCommand.AddCommand(testCommand)

	addCommonExecutionFlags(testCommand)
	testCommand.PersistentFlags().
		String("report", "", "The path to generate a report of the scenario execution. The contents of the report are in JSON and will only be generated when this flag is set.")
}

var testCommand = &cobra.Command{
	Use:   "test",
	Args:  cobra.MinimumNArgs(1),
	Short: "Test document commands against their expected outputs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := bindExecutionOptions(cmd, args)
		if err != nil {
			return handleExecutionOptionError(cmd, err)
		}

		innovationEngine, err := engineNewEngine(engine.EngineConfiguration{
			Verbose:          opts.Verbose,
			DoNotDelete:      false,
			StreamOutput:     opts.StreamOutput,
			Subscription:     opts.Subscription,
			CorrelationId:    "",
			WorkingDirectory: opts.WorkingDirectory,
			Environment:      opts.Environment,
			ReportFile:       opts.ReportFile,
		})
		if err != nil {
			return commandError(cmd, err, false, "error creating engine")
		}

		scenario, err := common.CreateScenarioFromMarkdown(
			opts.MarkdownPath,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			opts.EnvironmentVariables,
		)
		if err != nil {
			return commandError(cmd, err, false, "error creating scenario")
		}

		err = innovationEngine.TestScenario(scenario)
		if err != nil {
			return commandError(cmd, err, false, "scenario did not finish successfully")
		}

		return nil
	},
}
