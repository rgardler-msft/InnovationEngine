package commands

import (
	"fmt"
	"os"

	"github.com/Azure/InnovationEngine/internal/lib"

	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/engine/environments"
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
		markdownFile := args[0]
		if markdownFile == "" {
			logging.GlobalLogger.Errorf("Error: No markdown file specified.")
			cmd.Help()
			return fmt.Errorf("no markdown file specified")
		}

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

		verbose, _ := cmd.Flags().GetBool("verbose")
		doNotDelete, _ := cmd.Flags().GetBool("do-not-delete")

		subscription, _ := cmd.Flags().GetString("subscription")
		correlationId, _ := cmd.Flags().GetString("correlation-id")
		environment, _ := cmd.Flags().GetString("environment")
		workingDirectory, _ := cmd.Flags().GetString("working-directory")

		environmentVariables, _ := cmd.Flags().GetStringArray("var")
		// features, _ := cmd.Flags().GetStringArray("feature")

		// Known features
		renderValues := false

		cliEnvironmentVariables, err := lib.ParseEnvironmentVariableAssignments(environmentVariables)
		if err != nil {
			logging.GlobalLogger.Errorf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			cmd.Help()
			return err
		}
		// Parse the markdown file and create a scenario
		scenario, err := common.CreateScenarioFromMarkdown(
			markdownFile,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			cliEnvironmentVariables,
		)
		if err != nil {
			logging.GlobalLogger.Errorf("Error creating scenario: %s", err)
			fmt.Printf("Error creating scenario: %s", err)
			return fmt.Errorf("error creating scenario: %w", err)
		}

		innovationEngine, err := engine.NewEngine(engine.EngineConfiguration{
			Verbose:          verbose,
			DoNotDelete:      doNotDelete,
			StreamOutput:     true, // Interactive mode always streams
			Subscription:     subscription,
			CorrelationId:    correlationId,
			Environment:      environments.Environment(environment),
			WorkingDirectory: workingDirectory,
			RenderValues:     renderValues,
		})
		if err != nil {
			logging.GlobalLogger.Errorf("Error creating engine: %s", err)
			fmt.Printf("Error creating engine: %s", err)
			return fmt.Errorf("error creating engine: %w", err)
		}

		// Execute the scenario
		err = innovationEngine.InteractWithScenario(scenario)
		if err != nil {
			logging.GlobalLogger.Errorf("Error executing scenario: %s", err)
			fmt.Printf("Error executing scenario: %s", err)
			return fmt.Errorf("error executing scenario: %w", err)
		}

		return nil
	},
}
