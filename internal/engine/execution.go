package engine

import (
	"fmt"
	"os"
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
)

const (
	// TODO - Make this configurable for terminals that support it.
	// spinnerFrames  = `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
	spinnerFrames  = `-\|/`
	spinnerRefresh = 100 * time.Millisecond
)

type stepTiming struct {
	name     string
	section  string
	duration time.Duration
	segments []childTiming
}

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
					Section:    step.Section,
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
	failedVerificationMarkers := make(map[string]bool)

	err := az.SetSubscription(e.Configuration.Subscription)
	if err != nil {
		logging.GlobalLogger.Errorf("Invalid Config: Failed to set subscription: %s", err)
		azureStatus.SetError(err)
		environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
		return err
	}

	stepsToExecute := filterDeletionCommands(steps, e.Configuration.DoNotDelete)
	stepTimings := make([]stepTiming, 0, len(stepsToExecute))
	defer func() {
		if len(stepTimings) == 0 {
			return
		}
		printExecutionSummary(stepTimings)
	}()

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

	environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))

	for stepNumber, step := range stepsToExecute {
		stepStart := time.Now()
		segmentOrder := make([]string, 0)
		segmentAccumulators := make(map[string]*segmentAccumulator)
		recorded := false
		recordStepDuration := func() {
			if recorded {
				return
			}
			segments := make([]childTiming, 0, len(segmentOrder))
			for _, key := range segmentOrder {
				acc := segmentAccumulators[key]
				segments = append(segments, acc.toChildTiming())
			}
			stepTimings = append(stepTimings, stepTiming{
				name:     step.Name,
				section:  step.Section,
				duration: time.Since(stepStart),
				segments: segments,
			})
			recorded = true
		}
		recordPrereqSegment := func(name string, elapsed time.Duration, sectionType, source, heading string) {
			if name == "" {
				return
			}
			key := buildSegmentKey(name, source)
			acc := segmentAccumulators[key]
			if acc == nil {
				acc = &segmentAccumulator{name: name, source: source}
				segmentAccumulators[key] = acc
				segmentOrder = append(segmentOrder, key)
			}
			acc.duration += elapsed
			if sectionType != "" {
				acc.addStep(sectionType, elapsed, heading)
			}
		}
		stepTitle := fmt.Sprintf("%d. %s\n", stepNumber+1, step.Name)
		fmt.Println(ui.StepTitleStyle.Render(stepTitle))
		azureStatus.CurrentStep = stepNumber + 1

		var prereqDocSeq int
		var prereqDocOrder map[string]*struct {
			index    int
			sections int
		}
		if step.Name == "Prerequisites" {
			prereqDocOrder = make(map[string]*struct {
				index    int
				sections int
			})
		}

		for _, block := range step.CodeBlocks {
			blockType, autoMeta, hasAutoMeta := common.ParseAutoPrereqMetadata(block.Content)
			isBannerBlock := hasAutoMeta && blockType == "banner"
			isVerificationBlock := hasAutoMeta && blockType == "verification"
			isBodyBlock := hasAutoMeta && blockType == "body"

			markerValue := ""
			if hasAutoMeta {
				markerValue = autoMeta["marker"]
			}

			var prereqDocState *struct {
				index    int
				sections int
			}
			if prereqDocOrder != nil && hasAutoMeta && markerValue != "" {
				prereqDocState = prereqDocOrder[markerValue]
				if prereqDocState == nil {
					prereqDocSeq++
					prereqDocState = &struct {
						index    int
						sections int
					}{index: prereqDocSeq}
					prereqDocOrder[markerValue] = prereqDocState
				}
			}

			if isVerificationBlock && markerValue != "" && failedVerificationMarkers[markerValue] {
				continue
			}

			prereqSegmentName := ""
			prereqSegmentType := ""
			prereqSegmentSource := ""
			prereqSegmentHeading := ""
			if strings.EqualFold(step.Name, "Prerequisites") && (isVerificationBlock || isBodyBlock) {
				display := autoMeta["display"]
				if strings.TrimSpace(display) == "" {
					display = block.Header
				}
				nameOnly, fileName := splitDisplayAndSource(display)
				sectionLabel := "verification"
				if isBodyBlock {
					sectionLabel = "execution"
				}
				prereqSegmentName = strings.TrimSpace(nameOnly)
				if prereqSegmentName == "" {
					prereqSegmentName = strings.TrimSpace(display)
				}
				prereqSegmentType = sectionLabel
				prereqSegmentSource = autoMeta["source"]
				if strings.TrimSpace(prereqSegmentSource) == "" {
					prereqSegmentSource = fileName
				}
				headingValue := autoMeta["section"]
				if strings.TrimSpace(headingValue) == "" {
					headingValue = block.Section
				}
				if strings.EqualFold(headingValue, "Prerequisites") {
					headingValue = ""
				}
				prereqSegmentHeading = strings.TrimSpace(headingValue)
			}

			blockStart := time.Now()
			segmentRecorded := false
			recordBlockDuration := func() {
				if segmentRecorded {
					return
				}
				segmentRecorded = true
				if prereqSegmentName != "" {
					recordPrereqSegment(prereqSegmentName, time.Since(blockStart), prereqSegmentType, prereqSegmentSource, prereqSegmentHeading)
				}
			}

			commandContent := block.Content
			if hasAutoMeta {
				commandContent = common.StripAutoPrereqComment(commandContent)
			}

			// If this is a body block and the marker file exists (verification passed), skip rendering & execution entirely.
			if isBodyBlock && markerValue != "" {
				if _, err := os.Stat(markerValue); err == nil {
					// Body is skipped; continue to next block without any output.
					recordBlockDuration()
					continue
				}
			}

			if prereqDocState != nil && (isVerificationBlock || isBodyBlock) {
				prereqDocState.sections++
				label := fmt.Sprintf("%d.%d.%d", stepNumber+1, prereqDocState.index, prereqDocState.sections)
				sectionLabel := "Verification"
				if isBodyBlock {
					sectionLabel = "Execution"
				}
				display := autoMeta["display"]
				if strings.TrimSpace(display) == "" {
					display = block.Header
				}
				fmt.Printf("    %s %s – %s\n\n", label, sectionLabel, display)
			}

			displayContent := commandContent
			renderContent := commandContent
			if isBodyBlock {
				trimmed := stripPrereqBodyWrapper(commandContent)
				displayContent = trimmed
				renderContent = trimmed
			}
			if isBannerBlock {
				displayContent = ""
				renderContent = ""
			}

			blockToExecute := block
			blockToExecute.Content = commandContent

			// Remove any existing marker before starting verification blocks to ensure fresh evaluation.
			if isVerificationBlock && markerValue != "" {
				_ = common.RemovePrereqMarker(markerValue)
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

			suppressOutput := isBannerBlock
			var finalCommandOutput string
			if e.Configuration.RenderValues {
				// Render the codeblock.
				renderedCommand, err := renderCommand(renderContent)
				if err != nil {
					logging.GlobalLogger.Errorf("Failed to render command: %s", err.Error())
					azureStatus.SetError(err)
					environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
					recordBlockDuration()
					recordStepDuration()
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
			if !isBannerBlock {
				// Attempt to read persisted working directory state first; fall back to current process working directory.
				workingDir, err := lib.LoadWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile)
				if err != nil || workingDir == "" {
					cwd, cwdErr := os.Getwd()
					if cwdErr == nil {
						workingDir = cwd
					}
				}
				if e.Configuration.Verbose {
					// Print to console (indented to align with command blocks) when verbose is enabled.
					fmt.Printf("    %s\n", ui.VerboseStyle.Render("Working directory: "+workingDir))
				}
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

				// When streaming output, we skip spinner rendering since output
				// will appear in real-time
				streamOutput := e.Configuration.StreamOutput && !suppressOutput

				if !streamOutput {
					terminal.MoveCursorPositionUp(lines)
					// Render the spinner and hide the cursor.
					fmt.Print(ui.SpinnerStyle.Render("  "+string(spinnerFrames[0])) + " ")
					terminal.HideCursor()
				} else {
					// For streaming, just print a newline to separate from command display
					fmt.Println()
				}

				go func(block parsers.CodeBlock) {
					output, err := shells.ExecuteBashCommand(
						block.Content,
						shells.BashCommandConfiguration{
							EnvironmentVariables: lib.CopyMap(env),
							InheritEnvironment:   true,
							InteractiveCommand:   false,
							WriteToHistory:       true,
							StreamOutput:         streamOutput,
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
						if !streamOutput {
							terminal.ShowCursor()
						}

						if commandErr == nil {

							actualOutput := commandOutput.StdOut
							expectedOutput := block.ExpectedOutput.Content
							expectedSimilarity := block.ExpectedOutput.ExpectedSimilarity
							expectedRegexPattern := block.ExpectedOutput.ExpectedRegexPattern
							expectedOutputLanguage := block.ExpectedOutput.Language

							_, outputComparisonError := common.CompareCommandOutputs(actualOutput, expectedOutput, expectedSimilarity, expectedRegexPattern, expectedOutputLanguage)

							if outputComparisonError != nil {
								if isVerificationBlock {
									if markerValue != "" {
										failedVerificationMarkers[markerValue] = true
									}
									if !streamOutput {
										fmt.Print("\r    \n")
										terminal.MoveCursorPositionDown(lines)
									}
									renderExpectedActual(
										block.ExpectedOutput.Content,
										commandOutput.StdOut,
										expectedSimilarity,
										expectedRegexPattern,
										true,
									)
									// Suppress noisy warning log for expected verification failure; body will execute.
									// Failure means body should execute; marker stays absent.
									break renderingLoop
								}

								logging.GlobalLogger.Errorf("Error comparing command outputs: %s", outputComparisonError.Error())
								if !streamOutput {
									fmt.Print("\r    \n")
									terminal.MoveCursorPositionDown(lines)
								}
								renderExpectedActual(
									block.ExpectedOutput.Content,
									commandOutput.StdOut,
									expectedSimilarity,
									expectedRegexPattern,
									false,
								)

								azureStatus.SetError(outputComparisonError)
								environments.AttachResourceURIsToAzureStatus(
									&azureStatus,
									resourceGroupName,
									string(e.Configuration.Environment),
								)
								environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))

								recordBlockDuration()
								recordStepDuration()
								return outputComparisonError
							}

							// Suppress final success tick per UI refinement request.
							if !streamOutput {
								fmt.Printf("\r    \n")
								terminal.MoveCursorPositionDown(lines)
							}

							if strings.TrimSpace(commandOutput.StdOut) != "" && !streamOutput && !suppressOutput {
								fmt.Printf("%s\n", ui.RemoveHorizontalAlign(ui.VerboseStyle.Render(commandOutput.StdOut)))
							}

							// For a successful verification, create marker immediately (static banner will reflect outcome).
							if isVerificationBlock && markerValue != "" {
								if err := common.WritePrereqMarker(markerValue, autoMeta["display"]); err != nil {
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
								environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
							}

						} else {
							if !streamOutput {
								fmt.Printf("\r  %s \n", ui.ErrorStyle.Render("✗"))
								terminal.MoveCursorPositionDown(lines)
							}
							fmt.Printf("  %s\n", ui.ErrorMessageStyle.Render(commandErr.Error()))

							logging.GlobalLogger.Errorf("Error executing command: %s", commandErr.Error())

							if isVerificationBlock {
								if markerValue != "" {
									failedVerificationMarkers[markerValue] = true
								}
								logging.GlobalLogger.Warnf("Verification command execution failed for %s", autoMeta["display"])
								break renderingLoop
							}

							azureStatus.SetError(commandErr)
							environments.AttachResourceURIsToAzureStatus(
								&azureStatus,
								resourceGroupName,
								string(e.Configuration.Environment),
							)
							environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))

							recordBlockDuration()
							recordStepDuration()
							return commandErr
						}

						break renderingLoop
					default:
						if !streamOutput {
							frame = (frame + 1) % len(spinnerFrames)
							fmt.Printf("\r  %s", ui.SpinnerStyle.Render(string(spinnerFrames[frame])))
							time.Sleep(spinnerRefresh)
						} else {
							// In streaming mode, just wait a bit before checking again
							time.Sleep(spinnerRefresh)
						}
					}
				}
			} else {
				lines := strings.Count(displayContent, "\n")

				// If we're on the last step and the command is an SSH command, we need
				// to report the status before executing the command. This is needed for
				// one click deployments and does not affect the normal execution flow.
				if stepNumber == len(stepsToExecute)-1 && patterns.SshCommand.MatchString(commandContent) {
					azureStatus.Status = "Succeeded"
					environments.AttachResourceURIsToAzureStatus(&azureStatus, resourceGroupName, string(e.Configuration.Environment))
					environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
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

					if strings.TrimSpace(output.StdOut) != "" && !suppressOutput {
						fmt.Printf("  %s\n", ui.VerboseStyle.Render(output.StdOut))
					}

					if stepNumber != len(stepsToExecute)-1 {
						environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
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
						environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))
						recordBlockDuration()
						recordStepDuration()
						return commandExecutionError
					}
				}
			}

			// No dynamic messaging post-verification (static banners handle status).
			recordBlockDuration()
		}
		recordStepDuration()
	}

	// Report the final status of the deployment (Only applies to one click deployments).
	azureStatus.Status = "Succeeded"
	environments.AttachResourceURIsToAzureStatus(
		&azureStatus,
		resourceGroupName,
		string(e.Configuration.Environment),
	)
	environments.ReportAzureStatus(azureStatus, string(e.Configuration.Environment))

	switch {
	case e.Configuration.Environment.IsAzureLike():
		logging.GlobalLogger.Info(
			"Cleaning environment variable file located at /tmp/ie-env-vars",
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
		if err := lib.CleanEnvironmentStateFile(lib.DefaultEnvironmentStateFile); err != nil {
			logging.GlobalLogger.Warnf("Error cleaning environment variables: %s", err.Error())
		}
		lib.DeleteWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile)
	}

	return nil
}

func stripPrereqBodyWrapper(content string) string {
	trimmedContent := strings.TrimSuffix(content, "\n")
	lines := strings.Split(trimmedContent, "\n")
	if len(lines) < 3 {
		return content
	}

	first := strings.TrimSpace(lines[0])
	last := strings.TrimSpace(lines[len(lines)-1])
	if !strings.HasPrefix(first, "if [ ! -f ") || !strings.HasSuffix(first, "]; then") || last != "fi" {
		return content
	}

	inner := strings.Join(lines[1:len(lines)-1], "\n")
	return inner
}

func renderExpectedActual(expected string, actual string, expectedSimilarity float64, expectedRegexPattern string, isVerification bool) {
	trimmedActual := strings.TrimRight(actual, "\n")
	trimmedExpected := strings.TrimRight(expected, "\n")

	if isVerification {
		fmt.Println("  " + ui.WarningStyle.Render("Prerequisite verification failed, prereq needs to be run:"))
	} else {
		fmt.Println("  " + ui.ErrorMessageStyle.Render("Expected output does not match:"))
	}

	showSimilarity := strings.TrimSpace(expectedRegexPattern) == ""
	regexPattern := ""
	if strings.TrimSpace(expectedRegexPattern) != "" {
		regexPattern = expectedRegexPattern
		if parsed, err := strconv.ParseFloat(regexPattern, 64); err == nil {
			showSimilarity = true
			if expectedSimilarity == 0 {
				expectedSimilarity = parsed
			}
		}
	}

	if showSimilarity {
		threshold := formatSimilarityValue(expectedSimilarity)
		fmt.Printf("    Expected similarity level of %s to:\n", threshold)
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

func printExecutionSummary(stepTimings []stepTiming) {
	total := time.Duration(0)
	for _, timing := range stepTimings {
		total += timing.duration
	}
	sections := buildSectionSummaries(stepTimings)
	fmt.Println()
	fmt.Println("execution_summary:")
	fmt.Printf("  total: \"%s\"\n", formatElapsed(total))
	if len(sections) == 0 {
		fmt.Println("  sections: []")
		fmt.Println()
		return
	}
	fmt.Println("  sections:")
	for _, section := range sections {
		fmt.Printf("    - title: %q\n", section.Name)
		fmt.Printf("      duration: \"%s\"\n", formatElapsed(section.Duration))
		if len(section.Children) == 0 {
			continue
		}
		fmt.Println("      steps:")
		for _, child := range section.Children {
			fmt.Printf("        - name: %q\n", child.Name)
			if child.Source != "" {
				fmt.Printf("          filename: %q\n", child.Source)
			}
			if len(child.Steps) == 0 {
				fmt.Printf("          duration: \"%s\"\n", formatElapsed(child.Duration))
				continue
			}
			fmt.Println("          steps:")
			for _, step := range child.Steps {
				fmt.Printf("            %s:\n", step.Name)
				fmt.Printf("              duration: \"%s\"\n", formatElapsed(step.Duration))
				if len(step.Headings) > 0 {
					fmt.Println("              headings:")
					for _, heading := range step.Headings {
						fmt.Printf("                - name: %q\n", heading.Name)
						fmt.Printf("                  duration: \"%s\"\n", formatElapsed(heading.Duration))
					}
				}
			}
		}
	}
	fmt.Println()
}

type sectionTiming struct {
	Name     string
	Duration time.Duration
	Children []childTiming
}

type childTiming struct {
	Name     string
	Duration time.Duration
	Source   string
	Steps    []stepBreakdown
}

type stepBreakdown struct {
	Name     string
	Duration time.Duration
	Headings []headingBreakdown
}

type headingBreakdown struct {
	Name     string
	Duration time.Duration
}

type segmentAccumulator struct {
	name          string
	source        string
	duration      time.Duration
	stepDurations map[string]*stepDuration
	stepOrder     []string
}

type stepDuration struct {
	duration     time.Duration
	headings     map[string]time.Duration
	headingOrder []string
}

func (s *segmentAccumulator) addStep(name string, duration time.Duration, heading string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return
	}
	if s.stepDurations == nil {
		s.stepDurations = make(map[string]*stepDuration)
	}
	entry := s.stepDurations[trimmed]
	if entry == nil {
		entry = &stepDuration{}
		s.stepDurations[trimmed] = entry
		s.stepOrder = append(s.stepOrder, trimmed)
	}
	entry.duration += duration
	entry.addHeading(heading, duration)
}

func (s *segmentAccumulator) toChildTiming() childTiming {
	child := childTiming{
		Name:     s.name,
		Duration: s.duration,
		Source:   s.source,
	}
	if len(s.stepOrder) > 0 {
		steps := make([]stepBreakdown, 0, len(s.stepOrder))
		for _, stepName := range s.stepOrder {
			entry := s.stepDurations[stepName]
			if entry == nil {
				continue
			}
			steps = append(steps, entry.toBreakdown(stepName))
		}
		child.Steps = steps
	}
	return child
}

func (sd *stepDuration) addHeading(name string, duration time.Duration) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return
	}
	if sd.headings == nil {
		sd.headings = make(map[string]time.Duration)
	}
	if _, exists := sd.headings[trimmed]; !exists {
		sd.headingOrder = append(sd.headingOrder, trimmed)
	}
	sd.headings[trimmed] += duration
}

func (sd *stepDuration) toBreakdown(name string) stepBreakdown {
	breakdown := stepBreakdown{Name: name, Duration: sd.duration}
	if len(sd.headingOrder) == 0 {
		return breakdown
	}
	headings := make([]headingBreakdown, 0, len(sd.headingOrder))
	for _, headingName := range sd.headingOrder {
		headings = append(headings, headingBreakdown{
			Name:     headingName,
			Duration: sd.headings[headingName],
		})
	}
	breakdown.Headings = headings
	return breakdown
}

func buildSegmentKey(name, source string) string {
	trimmedSource := strings.TrimSpace(source)
	if trimmedSource == "" {
		return name
	}
	return name + "|" + trimmedSource
}

func buildSectionSummaries(stepTimings []stepTiming) []sectionTiming {
	sections := make([]sectionTiming, 0)
	index := make(map[string]int)
	ensureSection := func(name string) int {
		if idx, ok := index[name]; ok {
			return idx
		}
		sections = append(sections, sectionTiming{Name: name})
		idx := len(sections) - 1
		index[name] = idx
		return idx
	}

	for _, timing := range stepTimings {
		name := strings.TrimSpace(timing.name)
		if name == "" {
			name = "Unnamed Section"
		}
		sectionLabel := strings.TrimSpace(timing.section)
		sectionLower := strings.ToLower(sectionLabel)
		switch {
		case sectionLower == "prerequisites":
			idx := ensureSection("Prerequisites")
			sections[idx].Duration += timing.duration
			for _, seg := range timing.segments {
				sections[idx].Children = append(sections[idx].Children, seg)
			}
		case sectionLower == "steps":
			idx := ensureSection("Steps")
			sections[idx].Duration += timing.duration
			sections[idx].Children = append(sections[idx].Children, childTiming{Name: name, Duration: timing.duration})
		case sectionLower == "validation":
			idx := ensureSection("Validation")
			sections[idx].Duration += timing.duration
		case sectionLower == "verification":
			idx := ensureSection("Verification")
			sections[idx].Duration += timing.duration
		default:
			// Fallback to heuristics for legacy documents lacking section metadata.
			lower := strings.ToLower(name)
			switch {
			case lower == "prerequisites":
				idx := ensureSection("Prerequisites")
				sections[idx].Duration += timing.duration
				for _, seg := range timing.segments {
					sections[idx].Children = append(sections[idx].Children, seg)
				}
			case strings.HasPrefix(lower, "verification"):
				idx := ensureSection("Verification")
				sections[idx].Duration += timing.duration
			case strings.HasPrefix(lower, "step "):
				idx := ensureSection("Steps")
				sections[idx].Duration += timing.duration
				sections[idx].Children = append(sections[idx].Children, childTiming{Name: name, Duration: timing.duration})
			default:
				sectionName := name
				if sectionLabel != "" {
					sectionName = sectionLabel
				}
				idx := ensureSection(sectionName)
				sections[idx].Duration += timing.duration
			}
		}
	}

	return sections
}

func formatElapsed(duration time.Duration) string {
	if duration < time.Millisecond {
		return "0s"
	}
	remaining := duration
	parts := make([]string, 0, 3)
	if remaining >= time.Hour {
		hours := remaining / time.Hour
		parts = append(parts, fmt.Sprintf("%dh", hours))
		remaining -= hours * time.Hour
	}
	if remaining >= time.Minute {
		minutes := remaining / time.Minute
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		remaining -= minutes * time.Minute
	}
	if remaining >= time.Second {
		seconds := remaining / time.Second
		parts = append(parts, fmt.Sprintf("%ds", seconds))
		remaining -= seconds * time.Second
	}
	if len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dms", remaining/time.Millisecond))
	}
	return strings.Join(parts, " ")
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

func splitDisplayAndSource(display string) (string, string) {
	trimmed := strings.TrimSpace(display)
	if trimmed == "" {
		return "", ""
	}
	open := strings.LastIndex(trimmed, "[")
	close := strings.LastIndex(trimmed, "]")
	if open == -1 || close == -1 || close <= open {
		return trimmed, ""
	}
	name := strings.TrimSpace(trimmed[:open])
	source := strings.TrimSpace(trimmed[open+1 : close])
	return name, source
}
