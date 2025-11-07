package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/InnovationEngine/internal/az"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/parsers"
	"github.com/Azure/InnovationEngine/internal/patterns"
	"github.com/Azure/InnovationEngine/internal/shells"
	"github.com/Azure/InnovationEngine/internal/terminal"
	"github.com/Azure/InnovationEngine/internal/ui"
	"github.com/sirupsen/logrus"
)

const (
	// TODO - Make this configurable for terminals that support it.
	// spinnerFrames  = `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
	spinnerFrames  = `-\|/`
	spinnerRefresh = 100 * time.Millisecond
)

var (
	autoPrereqCommentPattern  = regexp.MustCompile(`^#\s*ie:auto-prereq-([a-z-]+)\s+(.*)$`)
	autoPrereqMetadataPattern = regexp.MustCompile(`([a-zA-Z0-9_-]+)="([^"]*)"`)
)

// If a scenario has an `az group delete` command and the `--do-not-delete`
// flag is set, we remove it from the steps.
func filterDeletionCommands(steps []common.Step, preserveResources bool) []common.Step {
	filteredSteps := []common.Step{}
	if preserveResources {
		for _, step := range steps {
			newBlocks := []parsers.CodeBlock{}
			for _, block := range step.CodeBlocks {
				if patterns.AzGroupDelete.MatchString(block.Content) {
					continue
				} else {
					newBlocks = append(newBlocks, block)
				}
			}
			if len(newBlocks) > -1 {
				filteredSteps = append(filteredSteps, common.Step{
					Name:       step.Name,
					CodeBlocks: newBlocks,
				})
			}
		}
	} else {
		filteredSteps = steps
	}
	return filteredSteps
}

func renderCommand(blockContent string) (shells.CommandOutput, error) {
	escapedCommand := blockContent
	if !patterns.MultilineQuotedStringCommand.MatchString(blockContent) {
		escapedCommand = strings.ReplaceAll(blockContent, "\\\n", "\\\\\n")
	}
	renderedCommand, err := shells.ExecuteBashCommand(
		"echo -e \""+escapedCommand+"\"",
		shells.BashCommandConfiguration{
			EnvironmentVariables: map[string]string{},
			InteractiveCommand:   false,
			WriteToHistory:       false,
			InheritEnvironment:   true,
		},
	)
	return renderedCommand, err
}

// Executes the steps from a scenario and renders the output to the terminal.
func (e *Engine) ExecuteAndRenderSteps(steps []common.Step, env map[string]string) error {
	var resourceGroupName string = ""
	azureStatus := environments.NewAzureDeploymentStatus()

	err := az.SetSubscription(e.Configuration.Subscription)
	if err != nil {
		logging.GlobalLogger.Errorf("Invalid Config: Failed to set subscription: %s", err)
		azureStatus.SetError(err)
		environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
		return err
	}

	stepsToExecute := filterDeletionCommands(steps, e.Configuration.DoNotDelete)

	// Dynamic verification state removed (static banner approach).

	for stepNumber, step := range stepsToExecute {

		azureCodeBlocks := []environments.AzureCodeBlock{}
		for _, block := range step.CodeBlocks {
			azureCodeBlocks = append(azureCodeBlocks, environments.AzureCodeBlock{
				Command:     block.Content,
				Description: block.Description,
			})
		}

		azureStatus.AddStep(fmt.Sprintf("%d. %s", stepNumber+1, step.Name), azureCodeBlocks)
	}

	environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)

	for stepNumber, step := range stepsToExecute {
		stepTitle := fmt.Sprintf("%d. %s\n", stepNumber+1, step.Name)
		fmt.Println(ui.StepTitleStyle.Render(stepTitle))
		azureStatus.CurrentStep = stepNumber + 1

		for _, block := range step.CodeBlocks {
			blockType, autoMeta, hasAutoMeta := parseAutoPrereqMetadata(block.Content)
			isBannerBlock := hasAutoMeta && blockType == "banner"
			isVerificationBlock := hasAutoMeta && blockType == "verification"
			isBodyBlock := hasAutoMeta && blockType == "body"

			markerValue := ""
			if hasAutoMeta {
				markerValue = autoMeta["marker"]
			}

			commandContent := block.Content
			if hasAutoMeta {
				commandContent = stripAutoPrereqComment(commandContent)
			}

			// If this is a body block and the marker file exists (verification passed), skip rendering & execution entirely.
			if isBodyBlock && markerValue != "" {
				if _, err := os.Stat(markerValue); err == nil {
					// Body is skipped; continue to next block without any output.
					continue
				}
			}

			displayContent := commandContent
			if isBannerBlock {
				displayContent = ""
			}

			blockToExecute := block
			blockToExecute.Content = commandContent

			// Remove any existing marker before starting verification blocks to ensure fresh evaluation.
			if isVerificationBlock && markerValue != "" {
				_ = os.Remove(markerValue)
			}

			// Render any descriptive markdown paragraphs that appeared immediately
			// before this code block in the source document. These are parsed into
			// CodeBlock.Description by the markdown parser. We always show them in
			// execute mode to preserve narrative context (not just in verbose).
			// Suppress body block description entirely if marker indicates skip.
			markerPresent := false
			if isBodyBlock && markerValue != "" {
				if _, err := os.Stat(markerValue); err == nil {
					markerPresent = true
				}
			}
			if strings.TrimSpace(block.Description) != "" && !(isBodyBlock && markerPresent) {
				descLines := strings.Split(block.Description, "\n")
				for _, line := range descLines {
					// Indent to align with command blocks for visual grouping.
					fmt.Printf("    %s\n", ui.VerboseStyle.Render(line))
				}
				// Blank line separating description from the command that follows.
				fmt.Println()
			}

			var finalCommandOutput string
			if e.Configuration.RenderValues {
				// Render the codeblock.
				renderedCommand, err := renderCommand(commandContent)
				if err != nil {
					logging.GlobalLogger.Errorf("Failed to render command: %s", err.Error())
					azureStatus.SetError(err)
					environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
					return err
				}
				finalCommandOutput = ui.IndentMultiLineCommand(renderedCommand.StdOut, 4)
			} else {
				finalCommandOutput = ui.IndentMultiLineCommand(displayContent, 4)
			}

			if isBannerBlock {
				finalCommandOutput = ""
			}

			// Debug/verbose working directory output before each command block.
			if (e.Configuration.Verbose || logging.GlobalLogger.GetLevel() <= logrus.DebugLevel) && !isBannerBlock {
				// Attempt to read persisted working directory state first; fall back to current process working directory.
				workingDir, err := lib.LoadWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile)
				if err != nil || workingDir == "" {
					cwd, cwdErr := os.Getwd()
					if cwdErr == nil {
						workingDir = cwd
					}
				}
				// Print to console (indented to align with command blocks) and log for deeper tracing.
				fmt.Printf("    %s\n", ui.VerboseStyle.Render("Working directory: "+workingDir))
				logging.GlobalLogger.Debugf("Working directory before command: %s", workingDir)
			}

			if finalCommandOutput != "" {
				fmt.Print("    " + finalCommandOutput)
			}

			// execute the command as a goroutine to allow for the spinner to be
			// rendered while the command is executing.
			done := make(chan error)
			var commandOutput shells.CommandOutput

			// If the command is an SSH command, we need to forward the input and
			// output
			interactiveCommand := false
			if patterns.SshCommand.MatchString(commandContent) {
				interactiveCommand = true
			}

			logging.GlobalLogger.WithField("isInteractive", interactiveCommand).
				Infof("Executing command: %s", commandContent)

			var commandErr error
			var frame int = 0

			// If forwarding input/output, don't render the spinner.
			if !interactiveCommand {
				// Grab the number of lines it contains & set the cursor to the
				// beginning of the block.

				lines := strings.Count(finalCommandOutput, "\n")
				terminal.MoveCursorPositionUp(lines)

				// Render the spinner and hide the cursor.
				fmt.Print(ui.SpinnerStyle.Render("  "+string(spinnerFrames[0])) + " ")
				terminal.HideCursor()

				go func(block parsers.CodeBlock) {
					output, err := shells.ExecuteBashCommand(
						block.Content,
						shells.BashCommandConfiguration{
							EnvironmentVariables: lib.CopyMap(env),
							InheritEnvironment:   true,
							InteractiveCommand:   false,
							WriteToHistory:       true,
						},
					)
					logging.GlobalLogger.Infof("Command output to stdout:\n %s", output.StdOut)
					logging.GlobalLogger.Infof("Command output to stderr:\n %s", output.StdErr)
					commandOutput = output
					done <- err
				}(blockToExecute)
			renderingLoop:
				// While the command is executing, render the spinner.
				for {
					select {
					case commandErr = <-done:
						// Show the cursor, check the result of the command, and display the
						// final status.
						terminal.ShowCursor()

						if commandErr == nil {

							actualOutput := commandOutput.StdOut
							expectedOutput := block.ExpectedOutput.Content
							expectedSimilarity := block.ExpectedOutput.ExpectedSimilarity
							expectedRegex := block.ExpectedOutput.ExpectedRegex
							expectedOutputLanguage := block.ExpectedOutput.Language

							_, outputComparisonError := common.CompareCommandOutputs(actualOutput, expectedOutput, expectedSimilarity, expectedRegex, expectedOutputLanguage)

							if outputComparisonError != nil {
								if isVerificationBlock {
									fmt.Print("\r    \n")
									terminal.MoveCursorPositionDown(lines)
									renderExpectedActual(
										block.ExpectedOutput.Content,
										commandOutput.StdOut,
										expectedSimilarity,
										expectedRegex,
									)
									// Suppress noisy warning log for expected verification failure; body will execute.
									// Failure means body should execute; marker stays absent.
									break renderingLoop
								}

								logging.GlobalLogger.Errorf("Error comparing command outputs: %s", outputComparisonError.Error())
								fmt.Print("\r    \n")
								terminal.MoveCursorPositionDown(lines)
								renderExpectedActual(
									block.ExpectedOutput.Content,
									commandOutput.StdOut,
									expectedSimilarity,
									expectedRegex,
								)

								azureStatus.SetError(outputComparisonError)
								environments.AttachResourceURIsToAzureStatus(
									&azureStatus,
									resourceGroupName,
									e.Configuration.Environment,
								)
								environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)

								return outputComparisonError
							}

							// Suppress final success tick per UI refinement request.
							fmt.Printf("\r    \n")
							terminal.MoveCursorPositionDown(lines)

							if strings.TrimSpace(commandOutput.StdOut) != "" {
								fmt.Printf("%s\n", ui.RemoveHorizontalAlign(ui.VerboseStyle.Render(commandOutput.StdOut)))
							}

							// For a successful verification, create marker immediately (static banner will reflect outcome).
							if isVerificationBlock && markerValue != "" {
								if err := writePrereqMarker(markerValue, autoMeta["display"]); err != nil {
									logging.GlobalLogger.Warnf("Failed to write marker %s: %v", markerValue, err)
								}
							}

							// Extract the resource group name from the command output if
							// it's not already set.
							if resourceGroupName == "" && patterns.AzCommand.MatchString(commandContent) {
								logging.GlobalLogger.Info("Attempting to extract resource group name from command output")
								tmpResourceGroup := az.FindResourceGroupName(commandOutput.StdOut)
								if tmpResourceGroup != "" {
									logging.GlobalLogger.WithField("resourceGroup", tmpResourceGroup).Info("Found resource group")
									resourceGroupName = tmpResourceGroup
									azureStatus.AddResourceURI(az.BuildResourceGroupId(e.Configuration.Subscription, resourceGroupName))
								}
							}

							if stepNumber != len(stepsToExecute)-1 {
								environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
							}

						} else {
							fmt.Printf("\r  %s \n", ui.ErrorStyle.Render("✗"))
							terminal.MoveCursorPositionDown(lines)
							fmt.Printf("  %s\n", ui.ErrorMessageStyle.Render(commandErr.Error()))

							logging.GlobalLogger.Errorf("Error executing command: %s", commandErr.Error())

							if isVerificationBlock {
								logging.GlobalLogger.Warnf("Verification command execution failed for %s", autoMeta["display"])
								break renderingLoop
							}

							azureStatus.SetError(commandErr)
							environments.AttachResourceURIsToAzureStatus(
								&azureStatus,
								resourceGroupName,
								e.Configuration.Environment,
							)
							environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)

							return commandErr
						}

						break renderingLoop
					default:
						frame = (frame + 1) % len(spinnerFrames)
						fmt.Printf("\r  %s", ui.SpinnerStyle.Render(string(spinnerFrames[frame])))
						time.Sleep(spinnerRefresh)
					}
				}
			} else {
				lines := strings.Count(displayContent, "\n")

				// If we're on the last step and the command is an SSH command, we need
				// to report the status before executing the command. This is needed for
				// one click deployments and does not affect the normal execution flow.
				if stepNumber == len(stepsToExecute)-1 && patterns.SshCommand.MatchString(commandContent) {
					azureStatus.Status = "Succeeded"
					environments.AttachResourceURIsToAzureStatus(&azureStatus, resourceGroupName, e.Configuration.Environment)
					environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
				}

				output, commandExecutionError := shells.ExecuteBashCommand(
					blockToExecute.Content,
					shells.BashCommandConfiguration{
						EnvironmentVariables: lib.CopyMap(env),
						InheritEnvironment:   true,
						InteractiveCommand:   true,
						WriteToHistory:       false,
					},
				)

				terminal.ShowCursor()

				if commandExecutionError == nil {
					// Suppress final success tick per UI refinement request.
					fmt.Printf("\r    \n")
					terminal.MoveCursorPositionDown(lines)

					if strings.TrimSpace(output.StdOut) != "" {
						fmt.Printf("  %s\n", ui.VerboseStyle.Render(output.StdOut))
					}

					if stepNumber != len(stepsToExecute)-1 {
						environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
					}
				} else {
					fmt.Printf("\r  %s \n", ui.ErrorStyle.Render("✗"))
					terminal.MoveCursorPositionDown(lines)
					fmt.Printf("  %s\n", ui.ErrorMessageStyle.Render(commandExecutionError.Error()))

					if isVerificationBlock {
						logging.GlobalLogger.Warnf("Verification command execution failed for %s", autoMeta["display"])
						// Failure means marker not written; body may still execute later.
					} else {
						azureStatus.SetError(commandExecutionError)
						environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)
						return commandExecutionError
					}
				}
			}

			// No dynamic messaging post-verification (static banners handle status).
		}
	}

	// Report the final status of the deployment (Only applies to one click deployments).
	azureStatus.Status = "Succeeded"
	environments.AttachResourceURIsToAzureStatus(
		&azureStatus,
		resourceGroupName,
		e.Configuration.Environment,
	)
	environments.ReportAzureStatus(azureStatus, e.Configuration.Environment)

	switch e.Configuration.Environment {
	case environments.EnvironmentsAzure, environments.EnvironmentsOCD:
		logging.GlobalLogger.Info(
			"Cleaning environment variable file located at /tmp/env-vars",
		)
		err := lib.CleanEnvironmentStateFile(lib.DefaultEnvironmentStateFile)
		if err != nil {
			logging.GlobalLogger.Errorf("Error cleaning environment variables: %s", err.Error())
			return err
		}

		logging.GlobalLogger.Info(
			"Cleaning working directory file located at /tmp/working-dir",
		)
		err = lib.DeleteWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile)
		if err != nil {
			logging.GlobalLogger.Errorf("Error cleaning working directory: %s", err.Error())
			return err
		}

	default:
		lib.DeleteEnvironmentStateFile(lib.DefaultEnvironmentStateFile)
		lib.DeleteWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile)
	}

	return nil
}

func parseAutoPrereqMetadata(content string) (string, map[string]string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "", nil, false
	}

	firstLine := strings.TrimSpace(lines[0])
	matches := autoPrereqCommentPattern.FindStringSubmatch(firstLine)
	if len(matches) != 3 {
		return "", nil, false
	}

	metadata := make(map[string]string)
	for _, match := range autoPrereqMetadataPattern.FindAllStringSubmatch(matches[2], -1) {
		if len(match) != 3 {
			continue
		}
		metadata[match[1]] = match[2]
	}

	return matches[1], metadata, true
}

func stripAutoPrereqComment(content string) string {
	if _, _, hasMetadata := parseAutoPrereqMetadata(content); !hasMetadata {
		return content
	}

	parts := strings.SplitN(content, "\n", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

func renderExpectedActual(expected string, actual string, expectedSimilarity float64, expectedRegex *regexp.Regexp) {
	trimmedActual := strings.TrimRight(actual, "\n")
	trimmedExpected := strings.TrimRight(expected, "\n")
	fmt.Println("  " + ui.ErrorMessageStyle.Render("Expected output does not match:"))

	showSimilarity := expectedRegex == nil
	regexPattern := ""
	if expectedRegex != nil {
		regexPattern = expectedRegex.String()
		if parsed, err := strconv.ParseFloat(regexPattern, 64); err == nil {
			showSimilarity = true
			if expectedSimilarity == 0 {
				expectedSimilarity = parsed
			}
		}
	}

	if showSimilarity {
		threshold := formatSimilarityValue(expectedSimilarity)
		fmt.Printf("    Expected similarity level of %s against:\n", threshold)
		renderIndentedBlock(trimmedExpected, "      ")
	} else {
		fmt.Println("    Expected RE match:")
		renderIndentedBlock(regexPattern, "      ")
	}

	fmt.Println("    Actual:")
	renderIndentedBlock(trimmedActual, "      ")
}

func renderIndentedBlock(content string, indent string) {
	if strings.TrimSpace(content) == "" {
		fmt.Printf("%s<empty>\n", indent)
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Println(indent)
			continue
		}
		fmt.Printf("%s%s\n", indent, ui.VerboseStyle.Render(line))
	}
}

func formatSimilarityValue(value float64) string {
	if value == 0 {
		return "0"
	}
	formatted := strconv.FormatFloat(value, 'g', -1, 64)
	if !strings.Contains(formatted, ".") {
		return fmt.Sprintf("%.1f", value)
	}
	return formatted
}

func writePrereqMarker(markerPath, display string) error {
	if strings.TrimSpace(markerPath) == "" {
		return nil
	}

	dir := filepath.Dir(markerPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(markerPath, []byte(display), 0o600)
}
