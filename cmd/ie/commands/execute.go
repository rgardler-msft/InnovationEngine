package commands

import (
	"os"

	"github.com/Azure/InnovationEngine/internal/lib"

	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
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
		markdownFile := args[0]
		if markdownFile == "" {
			return commandError(cmd, nil, true, "no markdown file specified")
		}

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

		verbose, _ := cmd.Flags().GetBool("verbose")
		doNotDelete, _ := cmd.Flags().GetBool("do-not-delete")
		streamOutput, _ := cmd.Flags().GetBool("stream-output")

		subscription, _ := cmd.Flags().GetString("subscription")
		correlationId, _ := cmd.Flags().GetString("correlation-id")
		workingDirectory, _ := cmd.Flags().GetString("working-directory")

		environmentSetting, err := getEnvironmentSetting(cmd)
		if err != nil {
			return commandError(cmd, err, false, "error resolving environment")
		}

		environmentVariables, _ := cmd.Flags().GetStringArray("var")
		features, _ := cmd.Flags().GetStringArray("feature")

		// Known features
		renderValues := false

		cliEnvironmentVariables, err := lib.ParseEnvironmentVariableAssignments(environmentVariables)
		if err != nil {
			return commandError(cmd, err, true, "invalid --var assignment")
		}

		for _, feature := range features {
			switch feature {
			case "render-values":
				renderValues = true
			default:
				return commandError(cmd, nil, true, "invalid feature: %s", feature)
			}
		}

		// Parse the markdown file and create a scenario
		scenario, err := common.CreateScenarioFromMarkdown(
			markdownFile,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			cliEnvironmentVariables,
		)
		if err != nil {
			return commandError(cmd, err, false, "error creating scenario")
		}

		innovationEngine, err := engineNewEngine(engine.EngineConfiguration{
			Verbose:          verbose,
			DoNotDelete:      doNotDelete,
			StreamOutput:     streamOutput,
			Subscription:     subscription,
			CorrelationId:    correlationId,
			Environment:      environmentSetting,
			WorkingDirectory: workingDirectory,
			RenderValues:     renderValues,
		})
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
