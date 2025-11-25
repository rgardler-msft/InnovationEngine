package commands

import (
	"os"

	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

// Register the command with our command runner.
func init() {
	rootCommand.AddCommand(interactiveCommand)

	addCommonExecutionFlags(interactiveCommand)
	addCorrelationFlag(interactiveCommand)
}

var interactiveCommand = &cobra.Command{
	Use:   "interactive [markdown file]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Execute a document in interactive mode.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Ensure we are in the original invocation directory before parsing
		// the first document.
		if OriginalInvocationDirectory != "" {
			if err := os.Chdir(OriginalInvocationDirectory); err != nil {
				logging.GlobalLogger.Warnf("Failed to change to invocation directory '%s': %s", OriginalInvocationDirectory, err)
			} else {
				logging.GlobalLogger.Debugf("Changed to original invocation directory: %s", OriginalInvocationDirectory)
				if err := lib.SaveWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile, OriginalInvocationDirectory); err != nil {
					logging.GlobalLogger.Warnf("Failed to persist invocation working directory: %s", err)
				}
			}
		}

		opts, err := bindExecutionOptions(cmd, args)
		if err != nil {
			return handleExecutionOptionError(cmd, err)
		}
		// Parse the markdown file and create a scenario
		scenario, err := common.CreateScenarioFromMarkdown(
			opts.MarkdownPath,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			opts.EnvironmentVariables,
		)
		if err != nil {
			return commandError(cmd, err, false, "error creating scenario")
		}

		innovationEngine, err := engineNewEngine(engine.EngineConfiguration{
			Verbose:          opts.Verbose,
			DoNotDelete:      opts.DoNotDelete,
			StreamOutput:     true, // Interactive mode always streams
			Subscription:     opts.Subscription,
			CorrelationId:    opts.CorrelationID,
			Environment:      opts.Environment,
			WorkingDirectory: opts.WorkingDirectory,
			RenderValues:     opts.RenderValues,
		})
		if err != nil {
			return commandError(cmd, err, false, "error creating engine")
		}

		// Execute the scenario
		err = innovationEngine.InteractWithScenario(scenario)
		if err != nil {
			return commandError(cmd, err, false, "error executing scenario")
		}

		return nil
	},
}
