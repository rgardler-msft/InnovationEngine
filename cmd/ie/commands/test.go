package commands

import (
	"fmt"

	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
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
		markdownFile := args[0]
		if markdownFile == "" {
			cmd.Help()
			return fmt.Errorf("no markdown file specified")
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		streamOutput, _ := cmd.Flags().GetBool("stream-output")
		subscription, _ := cmd.Flags().GetString("subscription")
		workingDirectory, _ := cmd.Flags().GetString("working-directory")
		generateReport, _ := cmd.Flags().GetString("report")

		environmentVariables, _ := cmd.Flags().GetStringArray("var")

		environmentSetting, err := getEnvironmentSetting(cmd)
		if err != nil {
			logging.GlobalLogger.Errorf("Error resolving environment: %s", err)
			fmt.Printf("Error resolving environment: %s\n", err)
			return err
		}

		cliEnvironmentVariables, err := lib.ParseEnvironmentVariableAssignments(environmentVariables)
		if err != nil {
			logging.GlobalLogger.Errorf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			cmd.Help()
			return err
		}

		innovationEngine, err := engine.NewEngine(engine.EngineConfiguration{
			Verbose:          verbose,
			DoNotDelete:      false,
			StreamOutput:     streamOutput,
			Subscription:     subscription,
			CorrelationId:    "",
			WorkingDirectory: workingDirectory,
			Environment:      environmentSetting,
			ReportFile:       generateReport,
		})
		if err != nil {
			logging.GlobalLogger.Errorf("Error creating engine %s", err)
			fmt.Printf("Error creating engine %s", err)
			return fmt.Errorf("error creating engine: %w", err)
		}

		scenario, err := common.CreateScenarioFromMarkdown(
			markdownFile,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			cliEnvironmentVariables,
		)
		if err != nil {
			logging.GlobalLogger.Errorf("Error creating scenario %s", err)
			fmt.Printf("Error creating engine %s", err)
			return fmt.Errorf("error creating scenario: %w", err)
		}

		err = innovationEngine.TestScenario(scenario)
		if err != nil {
			logging.GlobalLogger.Errorf("Error testing scenario: %s", err)
			fmt.Printf("Scenario did not finish successfully.")
			return fmt.Errorf("scenario did not finish successfully: %w", err)
		}

		return nil
	},
}
