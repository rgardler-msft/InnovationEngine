package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/lib/fs"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/parsers"
	"github.com/Azure/InnovationEngine/internal/patterns"
	"github.com/yuin/goldmark/ast"
)

// injectPrerequisitesRecursively walks the prerequisites graph starting from the
// provided markdown AST and source, inlining prerequisite execution blocks into
// the supplied codeBlocks slice. It merges YAML metadata and scenario
// variables from each prerequisite document and uses seenPrereqs to avoid
// infinite recursion on cyclic graphs.
func injectPrerequisitesRecursively(
	codeBlocks []parsers.CodeBlock,
	markdown ast.Node,
	source []byte,
	path string,
	languagesToExecute []string,
	introText string,
	prerequisiteSectionText string,
	properties map[string]interface{},
	environmentVariables map[string]string,
	seenPrereqs map[string]bool,
	prerequisiteSectionUsed *bool,
) []parsers.CodeBlock {
	ctx := &prerequisiteInjectionContext{
		languagesToExecute:      languagesToExecute,
		introText:               introText,
		prerequisiteSectionText: prerequisiteSectionText,
		properties:              properties,
		environmentVariables:    environmentVariables,
		seenPrereqs:             seenPrereqs,
		prerequisiteSectionUsed: prerequisiteSectionUsed,
	}

	return ctx.inject(codeBlocks, markdown, source, path)
}

type prerequisiteInjectionContext struct {
	languagesToExecute      []string
	introText               string
	prerequisiteSectionText string
	properties              map[string]interface{}
	environmentVariables    map[string]string
	seenPrereqs             map[string]bool
	prerequisiteSectionUsed *bool
}

func (ctx *prerequisiteInjectionContext) inject(
	codeBlocks []parsers.CodeBlock,
	markdown ast.Node,
	source []byte,
	path string,
) []parsers.CodeBlock {
	prerequisiteUrls, err := parsers.ExtractPrerequisiteUrlsFromAst(markdown, source)
	if err != nil {
		logging.GlobalLogger.Warn(err)
		return codeBlocks
	}
	if len(prerequisiteUrls) == 0 {
		return codeBlocks
	}

	for _, rawURL := range prerequisiteUrls {
		codeBlocks = ctx.processPrerequisite(codeBlocks, rawURL, path)
	}

	return codeBlocks
}

func (ctx *prerequisiteInjectionContext) processPrerequisite(
	codeBlocks []parsers.CodeBlock,
	rawURL string,
	parentPath string,
) []parsers.CodeBlock {
	logging.GlobalLogger.Infof("Preparing to execute prerequisite: %s", rawURL)
	resolvedURL := ctx.resolveURL(rawURL, parentPath)

	if ctx.alreadyProcessed(resolvedURL) {
		return codeBlocks
	}

	if !ctx.prerequisiteIsAvailable(resolvedURL) {
		return codeBlocks
	}

	prerequisiteSource, prerequisiteMarkdown, prereqTitle, prereqDisplay, ok := ctx.loadPrerequisite(resolvedURL)
	if !ok {
		return codeBlocks
	}

	ctx.mergePrerequisiteMetadata(prerequisiteMarkdown, prerequisiteSource)

	codeBlocks = ctx.inject(
		codeBlocks,
		prerequisiteMarkdown,
		prerequisiteSource,
		resolvedURL,
	)

	prerequisiteCodeBlocks := parsers.ExtractCodeBlocksFromAst(prerequisiteMarkdown, prerequisiteSource, ctx.languagesToExecute, resolvedURL)
	verificationBlocks, bodyBlocks := partitionPrerequisiteBlocks(prerequisiteCodeBlocks)

	beforePrereqs, afterPrereqs := splitScenarioBlocks(codeBlocks)
	afterPrereqs = stripTextFromFirstDescription(afterPrereqs, ctx.introText)
	afterPrereqs = stripTextFromFirstDescription(afterPrereqs, ctx.prerequisiteSectionText)

	markerFile := ctx.markerFile(prereqTitle)
	newPrereqBlocks := ctx.buildPrerequisiteExecutionBlocks(markerFile, prereqDisplay, verificationBlocks, bodyBlocks)

	updated := append([]parsers.CodeBlock{}, beforePrereqs...)
	updated = append(updated, newPrereqBlocks...)
	updated = append(updated, afterPrereqs...)

	return updated
}

func (ctx *prerequisiteInjectionContext) resolveURL(rawURL, parentPath string) string {
	if isRemotePath(rawURL) {
		return rawURL
	}
	return filepath.Join(filepath.Dir(parentPath), rawURL)
}

func (ctx *prerequisiteInjectionContext) alreadyProcessed(url string) bool {
	if ctx.seenPrereqs[url] {
		logging.GlobalLogger.Infof("Skipping already-processed prerequisite: %s", url)
		return true
	}
	ctx.seenPrereqs[url] = true
	return false
}

func (ctx *prerequisiteInjectionContext) prerequisiteIsAvailable(url string) bool {
	if isRemotePath(url) || fs.FileExists(url) {
		return true
	}
	msg := fmt.Sprintf("Prerequisite '%s' not found (continuing without it)", url)
	RegisterMissingPrerequisite(msg)
	return false
}

func (ctx *prerequisiteInjectionContext) loadPrerequisite(url string) ([]byte, ast.Node, string, string, bool) {
	prerequisiteSource, err := resolveMarkdownSource(url)
	if err != nil {
		msg := fmt.Sprintf("Prerequisite '%s' could not be loaded: %v (continuing without it)", url, err)
		RegisterMissingPrerequisite(msg)
		return nil, nil, "", "", false
	}

	prerequisiteMarkdown := parsers.ParseMarkdownIntoAst(prerequisiteSource)
	prereqTitle, titleErr := parsers.ExtractScenarioTitleFromAst(prerequisiteMarkdown, prerequisiteSource)
	if titleErr != nil || prereqTitle == "" {
		prereqTitle = filepath.Base(url)
	}
	prereqDisplay := fmt.Sprintf("%s [%s]", prereqTitle, filepath.Base(url))
	logging.GlobalLogger.Infof("Executing Prerequisite: %s", prereqDisplay)

	return prerequisiteSource, prerequisiteMarkdown, prereqTitle, prereqDisplay, true
}

func (ctx *prerequisiteInjectionContext) mergePrerequisiteMetadata(markdown ast.Node, source []byte) {
	prerequisiteProperties := parsers.ExtractYamlMetadataFromAst(markdown)
	for key, value := range prerequisiteProperties {
		ctx.properties[key] = value
	}

	prerequisiteVariables := parsers.ExtractScenarioVariablesFromAst(markdown, source)
	for key, value := range prerequisiteVariables {
		ctx.environmentVariables[key] = value
	}
}

func partitionPrerequisiteBlocks(blocks []parsers.CodeBlock) (verification, body []parsers.CodeBlock) {
	for _, block := range blocks {
		if strings.EqualFold(block.Header, "Verification") {
			verification = append(verification, block)
		} else {
			body = append(body, block)
		}
	}
	return verification, body
}

func splitScenarioBlocks(blocks []parsers.CodeBlock) (beforePrereqs, afterPrereqs []parsers.CodeBlock) {
	for _, block := range blocks {
		if block.Header == "Prerequisites" {
			beforePrereqs = append(beforePrereqs, block)
		} else {
			afterPrereqs = append(afterPrereqs, block)
		}
	}
	return beforePrereqs, afterPrereqs
}

func (ctx *prerequisiteInjectionContext) buildPrerequisiteExecutionBlocks(
	markerFile string,
	display string,
	verificationBlocks []parsers.CodeBlock,
	bodyBlocks []parsers.CodeBlock,
) []parsers.CodeBlock {
	blocks := []parsers.CodeBlock{ctx.buildValidationBanner(markerFile, display)}
	blocks = append(blocks, ctx.buildVerificationBlocks(markerFile, display, verificationBlocks)...)
	blocks = append(blocks, ctx.buildDecisionBanner(markerFile, display))
	blocks = append(blocks, wrapPrerequisiteBodyBlocks(markerFile, display, bodyBlocks)...)
	return blocks
}

func (ctx *prerequisiteInjectionContext) buildValidationBanner(markerFile, display string) parsers.CodeBlock {
	banner := parsers.CodeBlock{
		Language: "bash",
		Header:   "Prerequisites",
		Content:  fmt.Sprintf("# ie:auto-prereq-banner marker=\"%s\" display=\"%s\"\necho \"Validating Prerequisite: %s\"\n", markerFile, display, display),
	}

	if !*ctx.prerequisiteSectionUsed && strings.TrimSpace(ctx.prerequisiteSectionText) != "" {
		banner.Description = strings.TrimSpace(ctx.prerequisiteSectionText)
		*ctx.prerequisiteSectionUsed = true
	}

	return banner
}

func (ctx *prerequisiteInjectionContext) buildVerificationBlocks(markerFile, display string, verificationBlocks []parsers.CodeBlock) []parsers.CodeBlock {
	annotated := make([]parsers.CodeBlock, 0, len(verificationBlocks))
	for i, block := range verificationBlocks {
		updated := block
		metadata := fmt.Sprintf("# ie:auto-prereq-verification marker=\"%s\" display=\"%s\" index=\"%d\" total=\"%d\"\n", markerFile, display, i+1, len(verificationBlocks))
		updated.Content = metadata + block.Content
		originalHeader := updated.Header
		updated.Header = "Prerequisites"
		if originalHeader != "" && !strings.EqualFold(originalHeader, "Prerequisites") {
			if strings.TrimSpace(updated.Description) != "" {
				updated.Description = fmt.Sprintf("%s\n\n%s", originalHeader, updated.Description)
			} else {
				updated.Description = originalHeader
			}
		}
		annotated = append(annotated, updated)
	}
	return annotated
}

func (ctx *prerequisiteInjectionContext) buildDecisionBanner(markerFile, display string) parsers.CodeBlock {
	return parsers.CodeBlock{
		Language: "bash",
		Header:   "Prerequisites",
		Content:  fmt.Sprintf("# ie:auto-prereq-banner marker=\"%s\" display=\"%s\"\nif [ -f \"%s\" ]; then echo \"Skipping Prerequisite: %s (verification passed)\"; else echo \"Executing Prerequisite: %s\"; fi\n", markerFile, display, markerFile, display, display),
	}
}

func wrapPrerequisiteBodyBlocks(markerFile, display string, bodyBlocks []parsers.CodeBlock) []parsers.CodeBlock {
	wrapped := make([]parsers.CodeBlock, 0, len(bodyBlocks))
	for i := range bodyBlocks {
		content := fmt.Sprintf("# ie:auto-prereq-body marker=\"%s\" display=\"%s\"\nif [ ! -f \"%s\" ]; then\n%s\nfi\n", markerFile, display, markerFile, bodyBlocks[i].Content)
		bodyBlocks[i].Content = content
		originalHeader := bodyBlocks[i].Header
		bodyBlocks[i].Header = "Prerequisites"
		if originalHeader != "" && !strings.EqualFold(originalHeader, "Prerequisites") {
			if strings.TrimSpace(bodyBlocks[i].Description) != "" {
				bodyBlocks[i].Description = fmt.Sprintf("%s\n\n%s", originalHeader, bodyBlocks[i].Description)
			} else {
				bodyBlocks[i].Description = originalHeader
			}
		}
		wrapped = append(wrapped, bodyBlocks[i])
	}
	return wrapped
}

func (ctx *prerequisiteInjectionContext) markerFile(prereqTitle string) string {
	slug := strings.ToLower(prereqTitle)
	slug = prerequisiteSlugRegex.ReplaceAllString(slug, "_")
	return fmt.Sprintf("/tmp/prereq_%s_skip", slug)
}

func isRemotePath(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

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

// missingPrereqMessages holds any missing/unloadable prerequisite warnings for the current process.
// This is intentionally package-level so we can emit a single summary at the end of scenario execution.
var missingPrereqMessages []string

var prerequisiteSlugRegex = regexp.MustCompile("[^a-z0-9]+")

// RegisterMissingPrerequisite records a warning message about a missing or unloadable prerequisite for
// later summarization. Duplicates are allowed here and will be de-duplicated when summarized.
func RegisterMissingPrerequisite(msg string) {
	missingPrereqMessages = append(missingPrereqMessages, msg)
}

// SummarizeMissingPrerequisites logs a consolidated, de-duplicated summary of any missing prerequisites.
// Intended to be called once at the end of scenario execution.
func SummarizeMissingPrerequisites() {
	if len(missingPrereqMessages) == 0 {
		return
	}

	// De-duplicate while preserving stable output order.
	seen := make(map[string]bool)
	unique := make([]string, 0, len(missingPrereqMessages))
	for _, m := range missingPrereqMessages {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	// Sort for deterministic output when there are many, then re-log each warning.
	sort.Strings(unique)
	for _, m := range unique {
		logging.GlobalLogger.Warn(m)
	}

	// Reset so subsequent scenarios in the same process start clean.
	missingPrereqMessages = nil
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
	// Use a recursive helper so that prerequisites of prerequisites are also processed.
	codeBlocks = injectPrerequisitesRecursively(codeBlocks, markdown, source, path, languagesToExecute, introText, prerequisiteSectionText, properties, environmentVariables, make(map[string]bool), &prerequisiteSectionUsed)

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
