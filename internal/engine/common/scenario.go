package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/lib/fs"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/parsers"
	"github.com/Azure/InnovationEngine/internal/patterns"
	"github.com/yuin/goldmark/ast"
)

// Individual steps within a scenario.
type Step struct {
	Name       string
	CodeBlocks []parsers.CodeBlock
}

// Scenarios are the top-level object that represents a scenario to be executed.
type Scenario struct {
	Name        string
	IntroText   string
	MarkdownAst ast.Node
	Steps       []Step
	Properties  map[string]interface{}
	Environment map[string]string
	Source      []byte
}

// Get the markdown source for the scenario as a string.
func (s *Scenario) GetSourceAsString() string {
	return string(s.Source)
}

// Groups the codeblocks into steps based on the header of the codeblock.
// This organizes the codeblocks into steps that can be executed linearly.
func groupCodeBlocksIntoSteps(blocks []parsers.CodeBlock) []Step {
	var groupedSteps []Step
	headerIndex := make(map[string]int)

	for _, block := range blocks {
		if index, ok := headerIndex[block.Header]; ok {
			groupedSteps[index].CodeBlocks = append(groupedSteps[index].CodeBlocks, block)
		} else {
			headerIndex[block.Header] = len(groupedSteps)
			groupedSteps = append(groupedSteps, Step{
				Name:       block.Header,
				CodeBlocks: []parsers.CodeBlock{block},
			})
		}
	}

	return groupedSteps
}

func extractIntroTextBeforeSection(source []byte, sectionTitle string) string {
	if len(sectionTitle) == 0 {
		return ""
	}

	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	marker := "\n## " + sectionTitle
	idx := strings.Index(text, marker)
	if idx == -1 {
		return ""
	}

	intro := strings.TrimSpace(text[:idx])
	lines := strings.Split(intro, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "#") {
		lines = lines[1:]
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
	}
	intro = strings.TrimSpace(strings.Join(lines, "\n"))
	return intro
}

func stripTextFromFirstDescription(blocks []parsers.CodeBlock, text string) []parsers.CodeBlock {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return blocks
	}

	for i := range blocks {
		desc := blocks[i].Description
		if strings.TrimSpace(desc) == "" {
			continue
		}

		newDesc := strings.Replace(desc, text, "", 1)
		if newDesc == desc {
			newDesc = strings.Replace(desc, trimmed, "", 1)
		}
		if newDesc != desc {
			blocks[i].Description = strings.TrimSpace(newDesc)
			break
		}
	}

	return blocks
}

// Download the scenario markdown over http
func downloadScenarioMarkdown(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Given either a local or remote path to a markdown file, resolve the path to
// the markdown file and return the contents of the file.
func resolveMarkdownSource(path string) ([]byte, error) {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return downloadScenarioMarkdown(path)
	}

	if !fs.FileExists(path) {
		return nil, fmt.Errorf("markdown file '%s' does not exist", path)
	}

	return os.ReadFile(path)
}

// Creates a scenario object from a given markdown file. languagesToExecute is
// used to filter out code blocks that should not be parsed out of the markdown
// file.
func CreateScenarioFromMarkdown(
	path string,
	languagesToExecute []string,
	environmentVariableOverrides map[string]string,
) (*Scenario, error) {
	source, err := resolveMarkdownSource(path)
	if err != nil {
		return nil, err
	}

	// Load environment variables
	markdownINI := strings.TrimSuffix(path, filepath.Ext(path)) + ".ini"
	environmentVariables := make(map[string]string)

	// Check if the INI file exists & load it.
	if !fs.FileExists(markdownINI) {
		logging.GlobalLogger.Infof("INI file '%s' does not exist, skipping...", markdownINI)
	} else {
		logging.GlobalLogger.Infof("INI file '%s' exists, loading...", markdownINI)
		environmentVariables, err = parsers.ParseINIFile(markdownINI)
		if err != nil {
			return nil, err
		}

		for key, value := range environmentVariables {
			logging.GlobalLogger.Debugf("Setting %s=%s\n", key, value)
		}
	}

	// Convert the markdown into an AST and extract the scenario variables.
	markdown := parsers.ParseMarkdownIntoAst(source)
	properties := parsers.ExtractYamlMetadataFromAst(markdown)
	scenarioVariables := parsers.ExtractScenarioVariablesFromAst(markdown, source)
	for key, value := range scenarioVariables {
		environmentVariables[key] = value
	}

	// Extract the code blocks from the markdown file.
	codeBlocks := parsers.ExtractCodeBlocksFromAst(markdown, source, languagesToExecute, path)
	logging.GlobalLogger.WithField("CodeBlocks", codeBlocks).
		Debugf("Found %d code blocks", len(codeBlocks))

	prerequisiteSectionText := parsers.ExtractSectionTextFromMarkdown(source, "Prerequisites")
	prerequisiteSectionUsed := false
	introText := extractIntroTextBeforeSection(source, "Prerequisites")

	// Extract the URLs of any prerequisite documents linked from the markdown file.
	// TODO: This is a bit of a hack. Should be refactored to remove duplication. Use recursion.
	prerequisiteUrls, err := parsers.ExtractPrerequisiteUrlsFromAst(markdown, source)
	if err == nil && len(prerequisiteUrls) > 0 {
		for _, url := range prerequisiteUrls {
			// Load the prerequisite markdown and wrap its execution with start/end messages.
			logging.GlobalLogger.Infof("Preparing to execute prerequisite: %s", url)
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = filepath.Join(filepath.Dir(path), url)
			}
			// Explicit pre-check for local file existence to avoid bubbling up an error.
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !fs.FileExists(url) {
				msg := fmt.Sprintf("Prerequisite '%s' not found (continuing without it)", url)
				logging.GlobalLogger.Warn(msg)
				continue
			}
			prerequisiteSource, err := resolveMarkdownSource(url)
			if err != nil {
				// When a prerequisite document is not found or cannot be loaded output a warning and continue.
				msg := fmt.Sprintf("Prerequisite '%s' could not be loaded: %v (continuing without it)", url, err)
				logging.GlobalLogger.Warn(msg)
				continue
			}

			prerequisiteMarkdown := parsers.ParseMarkdownIntoAst(prerequisiteSource)
			// Attempt to extract a title for the prerequisite document; fallback to filename/URL.
			prereqTitle, titleErr := parsers.ExtractScenarioTitleFromAst(prerequisiteMarkdown, prerequisiteSource)
			if titleErr != nil || prereqTitle == "" {
				prereqTitle = filepath.Base(url)
			}
			prereqDisplay := fmt.Sprintf("%s [%s]", prereqTitle, filepath.Base(url))
			logging.GlobalLogger.Infof("Executing Prerequisite: %s", prereqDisplay)
			prerequisiteProperties := parsers.ExtractYamlMetadataFromAst(prerequisiteMarkdown)
			for key, value := range prerequisiteProperties {
				properties[key] = value
			}

			prerequisiteVariables := parsers.ExtractScenarioVariablesFromAst(prerequisiteMarkdown, prerequisiteSource)
			for key, value := range prerequisiteVariables {
				environmentVariables[key] = value
			}

			prerequisiteCodeBlocks := parsers.ExtractCodeBlocksFromAst(prerequisiteMarkdown, prerequisiteSource, languagesToExecute, url)

			// Partition prerequisite code blocks into verification and non-verification blocks.
			var verificationBlocks, nonVerificationBlocks []parsers.CodeBlock
			for _, b := range prerequisiteCodeBlocks {
				if strings.EqualFold(b.Header, "Verification") {
					verificationBlocks = append(verificationBlocks, b)
				} else {
					nonVerificationBlocks = append(nonVerificationBlocks, b)
				}
			}

			// Generate a slug for this prerequisite title to create unique marker file paths.
			slug := strings.ToLower(prereqTitle)
			// Replace any non-alphanumeric characters with underscores to create a safe slug.
			slug = regexp.MustCompile("[^a-z0-9]+").ReplaceAllString(slug, "_")
			markerFile := fmt.Sprintf("/tmp/prereq_%s_skip", slug)

			var beforePrerequisites, afterPrerequisites []parsers.CodeBlock
			for _, block := range codeBlocks {
				if block.Header == "Prerequisites" {
					beforePrerequisites = append(beforePrerequisites, block)
				} else {
					afterPrerequisites = append(afterPrerequisites, block)
				}
			}

			// Remove intro/prerequisite narrative from subsequent code blocks since we'll surface it on the banner.
			afterPrerequisites = stripTextFromFirstDescription(afterPrerequisites, introText)
			afterPrerequisites = stripTextFromFirstDescription(afterPrerequisites, prerequisiteSectionText)

			var rebuiltPrereqBlocks []parsers.CodeBlock

			// 1. Validation banner first so users see we're validating before running verification code.
			validationBanner := parsers.CodeBlock{
				Language: "bash",
				Header:   "Prerequisites",
				Content:  fmt.Sprintf("# ie:auto-prereq-banner marker=\"%s\" display=\"%s\"\necho \"Validating Prerequisite: %s\"\n", markerFile, prereqDisplay, prereqDisplay),
			}
			descriptionParts := []string{}
			if !prerequisiteSectionUsed && strings.TrimSpace(prerequisiteSectionText) != "" {
				descriptionParts = append(descriptionParts, strings.TrimSpace(prerequisiteSectionText))
				prerequisiteSectionUsed = true
			}
			if len(descriptionParts) > 0 {
				validationBanner.Description = strings.Join(descriptionParts, "\n\n")
			}
			rebuiltPrereqBlocks = append(rebuiltPrereqBlocks, validationBanner)

			// 2. Verification blocks so their output appears after the validation banner.
			// Preserve original subheading by injecting it into Description while forcing Header to 'Prerequisites'.
			for i, vb := range verificationBlocks {
				annotated := vb
				metadata := fmt.Sprintf("# ie:auto-prereq-verification marker=\"%s\" display=\"%s\" index=\"%d\" total=\"%d\"\n", markerFile, prereqDisplay, i+1, len(verificationBlocks))
				annotated.Content = metadata + vb.Content
				originalHeader := annotated.Header
				annotated.Header = "Prerequisites"
				if originalHeader != "" && !strings.EqualFold(originalHeader, "Prerequisites") {
					if strings.TrimSpace(annotated.Description) != "" {
						annotated.Description = fmt.Sprintf("%s\n\n%s", originalHeader, annotated.Description)
					} else {
						annotated.Description = originalHeader
					}
				}
				rebuiltPrereqBlocks = append(rebuiltPrereqBlocks, annotated)
			}

			// 3. Static decision banner (skip or execute) based on marker file written by any successful verification.
			decisionBanner := parsers.CodeBlock{
				Language: "bash",
				Header:   "Prerequisites",
				Content:  fmt.Sprintf("# ie:auto-prereq-banner marker=\"%s\" display=\"%s\"\nif [ -f \"%s\" ]; then echo \"Skipping Prerequisite: %s (verification passed)\"; else echo \"Executing Prerequisite: %s\"; fi\n", markerFile, prereqDisplay, markerFile, prereqDisplay, prereqDisplay),
			}
			rebuiltPrereqBlocks = append(rebuiltPrereqBlocks, decisionBanner)

			// 4. Non-verification prerequisite body wrapped so it only runs when marker absent.
			for i := range nonVerificationBlocks {
				wrapped := fmt.Sprintf("# ie:auto-prereq-body marker=\"%s\" display=\"%s\"\nif [ ! -f \"%s\" ]; then\n%s\nfi\n", markerFile, prereqDisplay, markerFile, nonVerificationBlocks[i].Content)
				nonVerificationBlocks[i].Content = wrapped
				originalHeader := nonVerificationBlocks[i].Header
				nonVerificationBlocks[i].Header = "Prerequisites"
				if originalHeader != "" && !strings.EqualFold(originalHeader, "Prerequisites") {
					if strings.TrimSpace(nonVerificationBlocks[i].Description) != "" {
						nonVerificationBlocks[i].Description = fmt.Sprintf("%s\n\n%s", originalHeader, nonVerificationBlocks[i].Description)
					} else {
						nonVerificationBlocks[i].Description = originalHeader
					}
				}
			}
			rebuiltPrereqBlocks = append(rebuiltPrereqBlocks, nonVerificationBlocks...)

			// 5. (Removed completion banner per request to suppress execution completed line.)

			// Recombine all codeblocks in the new order.
			codeBlocks = append([]parsers.CodeBlock{}, beforePrerequisites...)
			codeBlocks = append(codeBlocks, rebuiltPrereqBlocks...)
			codeBlocks = append(codeBlocks, afterPrerequisites...)
		}
	} else {
		logging.GlobalLogger.Warn(err)
	}

	varsToExport := lib.CopyMap(environmentVariableOverrides)
	for key, value := range environmentVariableOverrides {
		environmentVariables[key] = value
		logging.GlobalLogger.Debugf("Attempting to override %s with %s", key, value)
		exportRegex := patterns.ExportVariableRegex(key)

		for index, codeBlock := range codeBlocks {
			matches := exportRegex.FindAllStringSubmatch(codeBlock.Content, -1)

			if len(matches) != 0 {
				logging.GlobalLogger.Debugf(
					"Found %d matches for %s, deleting from varsToExport",
					len(matches),
					key,
				)
				delete(varsToExport, key)
			} else {
				logging.GlobalLogger.Debugf("Found no matches for %s inside of %s", key, codeBlock.Content)
			}

			for _, match := range matches {
				oldLine := match[0]
				oldValue := match[1]

				// Replace the old export with the new export statement
				newLine := strings.Replace(oldLine, oldValue, value+" ", 1)
				logging.GlobalLogger.Debugf("Replacing '%s' with '%s'", oldLine, newLine)

				// Update the code block with the new export statement
				codeBlocks[index].Content = strings.Replace(codeBlock.Content, oldLine, newLine, 1)
			}

		}
	}

	// If there are some variables left after going through each of the codeblocks,
	// do not update the scenario
	// steps.
	if len(varsToExport) != 0 {
		logging.GlobalLogger.Debugf(
			"Found %d variables to add to the scenario as a step.",
			len(varsToExport),
		)
		exportCodeBlock := parsers.CodeBlock{
			Language:       "bash",
			Content:        "",
			Header:         "Exporting variables defined via the CLI and not in the markdown file.",
			ExpectedOutput: parsers.ExpectedOutputBlock{},
		}
		for key, value := range varsToExport {
			exportCodeBlock.Content += fmt.Sprintf("export %s=\"%s\"\n", key, value)
		}

		codeBlocks = append([]parsers.CodeBlock{exportCodeBlock}, codeBlocks...)
	}

	// Group the code blocks into steps.
	steps := groupCodeBlocksIntoSteps(codeBlocks)

	// If no title is found, we simply use the name of the markdown file as
	// the title of the scenario.
	title, err := parsers.ExtractScenarioTitleFromAst(markdown, source)
	if err != nil {
		logging.GlobalLogger.Warnf(
			"Failed to extract scenario title: '%s'. Using the name of the markdown as the scenario title",
			err,
		)
		title = filepath.Base(path)
	}

	logging.GlobalLogger.Infof("Successfully built out the scenario: %s", title)

	return &Scenario{
		Name:        title,
		IntroText:   strings.TrimSpace(introText),
		Environment: environmentVariables,
		Steps:       steps,
		Properties:  properties,
		MarkdownAst: markdown,
		Source:      source,
	}, nil
}

// Convert a scenario into a shell script
func (s *Scenario) ToShellScript() string {
	var script strings.Builder

	for key, value := range s.Environment {
		script.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, value))
	}

	for _, step := range s.Steps {
		script.WriteString(fmt.Sprintf("# %s\n", step.Name))
		for _, block := range step.CodeBlocks {
			script.WriteString(fmt.Sprintf("%s\n", block.Content))
		}
	}

	return script.String()
}
