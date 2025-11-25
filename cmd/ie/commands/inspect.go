package commands

import (
	"fmt"

	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/ui"
	"github.com/spf13/cobra"
)

// Register the command with our command runner.
func init() {
	rootCommand.AddCommand(inspectCommand)

	addCommonExecutionFlags(inspectCommand)
	addCorrelationFlag(inspectCommand)
}

var inspectCommand = &cobra.Command{
	Use:   "inspect [markdown file]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Execute a document in inspect mode.",
	RunE: func(cmd *cobra.Command, args []string) error {
		markdownFile := args[0]
		if markdownFile == "" {
			logging.GlobalLogger.Errorf("Error: No markdown file specified.")
			cmd.Help()
			return fmt.Errorf("no markdown file specified")
		}

		environmentVariables, _ := cmd.Flags().GetStringArray("var")
		// features, _ := cmd.Flags().GetStringArray("feature")

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
			[]string{"bash", "azurecli", "azurecli-inspect", "terraform"},
			cliEnvironmentVariables,
		)
		if err != nil {
			logging.GlobalLogger.Errorf("Error creating scenario: %s", err)
			fmt.Printf("Error creating scenario: %s", err)
			return fmt.Errorf("error creating scenario: %w", err)
		}

		fmt.Println(ui.ScenarioTitleStyle.Render(scenario.Name))
		for stepNumber, step := range scenario.Steps {
			stepTitle := fmt.Sprintf("  %d. %s\n", stepNumber+1, step.Name)
			fmt.Println(ui.StepTitleStyle.Render(stepTitle))
			for codeBlockNumber, codeBlock := range step.CodeBlocks {
				fmt.Println(
					ui.InteractiveModeCodeBlockDescriptionStyle.Render(
						fmt.Sprintf(
							"    %d.%d %s",
							stepNumber+1,
							codeBlockNumber+1,
							codeBlock.Description,
						),
					),
				)
				fmt.Print(
					ui.IndentMultiLineCommand(
						fmt.Sprintf(
							"      %s",
							ui.InteractiveModeCodeBlockStyle.Render(
								codeBlock.Content,
							),
						),
						6),
				)
				fmt.Println()
			}
		}

		return nil
	},
}
