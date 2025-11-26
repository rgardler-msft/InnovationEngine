package commands

import (
	"fmt"
	"strings"

	"github.com/Azure/InnovationEngine/internal/engine/common"
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

func partitionValidationIssues(issues []common.ValidationIssue) (warnings []string, errors []string) {
	for _, issue := range issues {
		switch issue.Severity {
		case common.ValidationSeverityWarning:
			warnings = append(warnings, issue.Message)
		case common.ValidationSeverityError:
			errors = append(errors, issue.Message)
		}
	}
	return warnings, errors
}

func formatValidationDetails(messages []string) string {
	if len(messages) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, message := range messages {
		builder.WriteString("- ")
		builder.WriteString(message)
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func formatWarningSummary(count int) string {
	return fmt.Sprintf("Warning: validation warnings detected (%d); see details below.", count)
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

var inspectCommand = &cobra.Command{
	Use:   "inspect [markdown file]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Lint an executable document without running code blocks.",
	Long:  `inspect performs structural linting against a document before you run it. The command validates language tags, prerequisite expected_results blocks (with exceptions for export-only code), environment variable prefixes, and usage (unused exports become warnings, undefined uppercase variables become errors). It never executes the fenced code blocks—use inspect as a safe preflight step before interactive, execute, or test modes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := bindExecutionOptions(cmd, args)
		if err != nil {
			return handleExecutionOptionError(cmd, err)
		}

		stopCapture := logging.StartWarningCapture()
		scenario, err := createScenarioFromOptions(opts, inspectRunnerTypes)
		capturedWarnings := stopCapture()
		if err != nil {
			writer := cmd.ErrOrStderr()
			for _, warning := range capturedWarnings {
				fmt.Fprintln(writer, ui.WarningStyle.Render(warning))
			}
			return commandError(cmd, err, false, "error creating scenario")
		}

		issues := common.ValidateScenarioForInspect(scenario)
		for _, warning := range capturedWarnings {
			issues = append(issues, common.ValidationIssue{Severity: common.ValidationSeverityWarning, Message: warning})
		}
		for _, msg := range common.DrainMissingPrerequisites() {
			issues = append(issues, common.ValidationIssue{Severity: common.ValidationSeverityError, Message: msg})
		}
		warnings, errors := partitionValidationIssues(issues)
		writer := cmd.ErrOrStderr()
		if len(errors) > 0 {
			if len(warnings) > 0 {
				fmt.Fprintln(writer, ui.WarningStyle.Render(formatWarningSummary(len(warnings))))
				for _, warning := range warnings {
					fmt.Fprintln(writer, ui.WarningStyle.Render(fmt.Sprintf("- %s", warning)))
				}
			}
			summary := fmt.Sprintf("document failed inspection checks (%d validation error%s)", len(errors), pluralSuffix(len(errors)))
			details := formatValidationDetails(errors)
			var errPayload string
			if details != "" {
				errPayload = ui.ErrorStyle.Render(details)
			}
			errResult := commandError(cmd, fmt.Errorf("%s", errPayload), false, summary)
			fmt.Fprintln(writer, "")
			fmt.Fprintln(writer, summary)
			return errResult
		}
		if len(warnings) > 0 {
			summary := formatWarningSummary(len(warnings))
			fmt.Fprintln(writer, ui.WarningStyle.Render(summary))
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

		if len(warnings) > 0 {
			fmt.Fprintln(writer, "")
			fmt.Fprintln(writer, ui.WarningStyle.Render("Validation warning details:"))
			for _, warning := range warnings {
				fmt.Fprintln(writer, ui.WarningStyle.Render(fmt.Sprintf("- %s", warning)))
			}
		}

		return nil
	},
}
