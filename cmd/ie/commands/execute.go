package commands

import (
	"os"

	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

// / Register the command with our command runner.
func init() {
	rootCommand.AddCommand(executeCommand)

	addCommonExecutionFlags(executeCommand)
	addCorrelationFlag(executeCommand)
}

var executeCommand = &cobra.Command{
	Use:   "execute [markdown file]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Execute the commands in an executable document.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Ensure we are in the original invocation directory before parsing
		// the first document regardless of any working directory flags that
		// will be applied later during execution.
		if OriginalInvocationDirectory != "" {
			if err := os.Chdir(OriginalInvocationDirectory); err != nil {
				logging.GlobalLogger.Warnf("Failed to change to invocation directory '%s': %s", OriginalInvocationDirectory, err)
			} else {
				logging.GlobalLogger.Debugf("Changed to original invocation directory: %s", OriginalInvocationDirectory)
				// Overwrite any stale working directory state so the first
				// command executes relative to the invocation directory, not
				// a previous run.
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

		cfg := buildEngineConfiguration(opts)

		innovationEngine, err := engineNewEngine(cfg)
		if err != nil {
			return commandError(cmd, err, false, "error creating engine")
		}

		// Execute the scenario
		err = innovationEngine.ExecuteScenario(scenario)
		if err != nil {
			return commandError(cmd, err, false, "error executing scenario")
		}

		return nil
	},
}
