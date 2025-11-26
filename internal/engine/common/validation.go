package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/InnovationEngine/internal/parsers"
	"github.com/yuin/goldmark/ast"
)

// ValidationSeverity identifies whether an inspection issue is fatal or informational.
type ValidationSeverity string

const (
	// ValidationSeverityError represents a blocking issue discovered during inspection.
	ValidationSeverityError ValidationSeverity = "error"
	// ValidationSeverityWarning surfaces non-blocking guidance.
	ValidationSeverityWarning ValidationSeverity = "warning"
)

// ValidationIssue captures a single inspection finding.
type ValidationIssue struct {
	Severity ValidationSeverity
	Message  string
}

var (
	exportStatementRegex = regexp.MustCompile(`(?m)^\s*export\s+([A-Za-z_][A-Za-z0-9_]*)`)
	envReferenceRegex    = regexp.MustCompile(`\$(\{)?([A-Za-z_][A-Za-z0-9_]*)`)
	assignmentRegex      = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)=`)
)

var allowedExternalEnvVars = map[string]struct{}{
	"HOME":     {},
	"PATH":     {},
	"PWD":      {},
	"OLDPWD":   {},
	"TMPDIR":   {},
	"TMP":      {},
	"TEMP":     {},
	"SHELL":    {},
	"USER":     {},
	"USERNAME": {},
	"HOSTNAME": {},
	"RANDOM":   {},
	"UID":      {},
	"EUID":     {},
	"GROUPS":   {},
}

// ValidateScenarioForInspect runs static checks that help authors find structural issues
// before executing a document.
func ValidateScenarioForInspect(s *Scenario) []ValidationIssue {
	if s == nil {
		return nil
	}

	var issues []ValidationIssue
	issues = append(issues, validateCodeBlockDescriptions(s)...)              // Author hygiene
	issues = append(issues, validateLanguageTags(s.MarkdownAst, s.Source)...) // Missing language tags
	issues = append(issues, validatePrerequisiteExpectedOutputs(s)...)        // Prerequisite verification blocks
	exports := collectEnvExports(s.Steps)
	issues = append(issues, validateEnvPrefixConsistency(exports)...)      // Prefix conventions
	issues = append(issues, validateEnvUsage(s, exports)...)               // Unused exports
	issues = append(issues, validateUndefinedEnvReferences(s, exports)...) // Missing exports
	issues = append(issues, validateExpectedSimilarityRanges(s)...)        // Similarity bounds
	return issues
}

func validateCodeBlockDescriptions(s *Scenario) []ValidationIssue {
	var issues []ValidationIssue
	for _, step := range s.Steps {
		for idx, block := range step.CodeBlocks {
			if isSystemGeneratedBlock(block) {
				continue
			}
			if strings.TrimSpace(block.Description) == "" {
				issues = append(issues, ValidationIssue{
					Severity: ValidationSeverityError,
					Message:  fmt.Sprintf("Step %q command #%d must include descriptive text before the code block.", step.Name, idx+1),
				})
			}
			if strings.TrimSpace(block.Language) == "" {
				issues = append(issues, ValidationIssue{
					Severity: ValidationSeverityError,
					Message:  fmt.Sprintf("Step %q command #%d must declare a language tag (e.g. ```bash).", step.Name, idx+1),
				})
			}
		}
	}
	return issues
}

func validateLanguageTags(node ast.Node, source []byte) []ValidationIssue {
	if node == nil {
		return nil
	}
	var issues []ValidationIssue
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if block, ok := n.(*ast.FencedCodeBlock); ok {
			language := strings.TrimSpace(string(block.Language(source)))
			if language == "" {
				snippet := truncateSnippet(extractFirstLine(block, source))
				issues = append(issues, ValidationIssue{
					Severity: ValidationSeverityError,
					Message:  fmt.Sprintf("Code block starting with %q is missing a language tag (```bash, ```azurecli, etc.).", snippet),
				})
			}
		}
		return ast.WalkContinue, nil
	})
	return issues
}

func validatePrerequisiteExpectedOutputs(s *Scenario) []ValidationIssue {
	var issues []ValidationIssue
	for _, step := range s.Steps {
		for idx, block := range step.CodeBlocks {
			if !block.InPrerequisiteSection || isSystemGeneratedBlock(block) {
				continue
			}
			if codeBlockContainsOnlyExports(block) {
				continue
			}
			hasLiteral := strings.TrimSpace(block.ExpectedOutput.Content) != ""
			if !hasLiteral && block.ExpectedOutput.ExpectedRegex == nil {
				issues = append(issues, ValidationIssue{
					Severity: ValidationSeverityError,
					Message:  fmt.Sprintf("Prerequisite command %q #%d must include an expected_results block to verify success.", step.Name, idx+1),
				})
			}
		}
	}
	return issues
}

func validateExpectedSimilarityRanges(s *Scenario) []ValidationIssue {
	var issues []ValidationIssue
	for _, step := range s.Steps {
		for idx, block := range step.CodeBlocks {
			sim := block.ExpectedOutput.ExpectedSimilarity
			if sim < 0 || sim > 1 {
				issues = append(issues, ValidationIssue{
					Severity: ValidationSeverityWarning,
					Message:  fmt.Sprintf("Step %q command #%d declares expected_similarity %.2f which is outside the 0-1 range.", step.Name, idx+1, sim),
				})
			}
		}
	}
	return issues
}

func validateEnvPrefixConsistency(exports []envExport) []ValidationIssue {
	if len(exports) == 0 {
		return nil
	}
	var issues []ValidationIssue
	for _, export := range exports {
		if export.Name == "HASH" {
			continue // HASH is a special helper variable and does not require a prefix.
		}
		_, ok := extractEnvPrefix(export.Name)
		if !ok {
			issues = append(issues, ValidationIssue{
				Severity: ValidationSeverityError,
				Message:  fmt.Sprintf("Environment variable %s (%s) must use an uppercase prefix followed by '_' (e.g. PREFIX_value).", export.Name, export.Location),
			})
		}
	}
	return issues
}

func validateEnvUsage(s *Scenario, exports []envExport) []ValidationIssue {
	if len(exports) == 0 {
		return nil
	}
	usage := make(map[string]bool, len(exports))
	patterns := make(map[string]*regexp.Regexp, len(exports))
	for _, export := range exports {
		patterns[export.Name] = compileEnvReferenceRegex(export.Name)
	}

	for _, step := range s.Steps {
		for _, block := range step.CodeBlocks {
			if isSystemGeneratedBlock(block) {
				continue
			}
			lines := strings.Split(block.Content, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				var exportedName string
				if strings.HasPrefix(trimmed, "export ") {
					if matches := exportStatementRegex.FindStringSubmatch(line); len(matches) > 1 {
						exportedName = matches[1]
					}
				}
				if isEchoLikeLine(trimmed) {
					continue
				}
				for _, export := range exports {
					if usage[export.Name] {
						continue
					}
					if exportedName != "" && exportedName == export.Name {
						continue
					}
					loc := patterns[export.Name].FindStringIndex(line)
					if loc == nil {
						continue
					}
					if isEchoBeforeMatch(line, loc[0]) {
						continue
					}
					usage[export.Name] = true
				}
			}
		}
	}

	var issues []ValidationIssue
	for _, export := range exports {
		if usage[export.Name] {
			continue
		}
		issues = append(issues, ValidationIssue{
			Severity: ValidationSeverityWarning,
			Message:  fmt.Sprintf("Environment variable %s (%s) is exported but never referenced outside echo/printf statements.", export.Name, export.Location),
		})
	}
	return issues
}

func validateUndefinedEnvReferences(s *Scenario, exports []envExport) []ValidationIssue {
	if s == nil {
		return nil
	}
	defined := make(map[string]struct{}, len(exports))
	for _, export := range exports {
		defined[export.Name] = struct{}{}
	}
	missing := make(map[string]string)
	for _, step := range s.Steps {
		for blockIdx, block := range step.CodeBlocks {
			if isSystemGeneratedBlock(block) {
				continue
			}
			lines := strings.Split(block.Content, "\n")
			for lineIdx, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				if name, ok := findAssignedVariable(trimmed); ok {
					defined[name] = struct{}{}
				}
				for _, ref := range findEnvReferences(line) {
					if isLowerCaseName(ref) {
						continue
					}
					if _, ok := defined[ref]; ok {
						continue
					}
					if _, ok := allowedExternalEnvVars[ref]; ok {
						continue
					}
					if _, recorded := missing[ref]; recorded {
						continue
					}
					missing[ref] = fmt.Sprintf("step %q block %d line %d", step.Name, blockIdx+1, lineIdx+1)
				}
			}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	issues := make([]ValidationIssue, 0, len(missing))
	for name, location := range missing {
		issues = append(issues, ValidationIssue{
			Severity: ValidationSeverityError,
			Message:  fmt.Sprintf("Environment variable %s (%s) is referenced but never exported in this document.", name, location),
		})
	}
	return issues
}

type envExport struct {
	Name     string
	Location string
}

func collectEnvExports(steps []Step) []envExport {
	seen := make(map[string]envExport)
	order := make([]envExport, 0)
	for _, step := range steps {
		for blockIdx, block := range step.CodeBlocks {
			if isSystemGeneratedBlock(block) {
				continue
			}
			lines := strings.Split(block.Content, "\n")
			for lineIdx, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				matches := exportStatementRegex.FindStringSubmatch(line)
				if len(matches) == 0 {
					continue
				}
				name := matches[1]
				if _, exists := seen[name]; exists {
					continue
				}
				location := fmt.Sprintf("step %q block %d line %d", step.Name, blockIdx+1, lineIdx+1)
				export := envExport{Name: name, Location: location}
				seen[name] = export
				order = append(order, export)
			}
		}
	}
	return order
}

func extractEnvPrefix(name string) (string, bool) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 || parts[0] == "" {
		return "", false
	}
	prefix := parts[0]
	if prefix != strings.ToUpper(prefix) {
		return prefix, false
	}
	return prefix, true
}

func compileEnvReferenceRegex(name string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(name)
	pattern := fmt.Sprintf(`\$(\{%s\}|%s)([^A-Za-z0-9_]|$)`, escaped, escaped)
	return regexp.MustCompile(pattern)
}

func findEnvReferences(line string) []string {
	if line == "" {
		return nil
	}
	matches := envReferenceRegex.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil
	}
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		refs = append(refs, match[2])
	}
	return refs
}

func findAssignedVariable(line string) (string, bool) {
	matches := assignmentRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

func isLowerCaseName(name string) bool {
	return name != "" && name == strings.ToLower(name)
}

func codeBlockContainsOnlyExports(block parsers.CodeBlock) bool {
	lines := strings.Split(block.Content, "\n")
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(trimmed, "export ") {
			return false
		}
		if !exportStatementRegex.MatchString(line) {
			return false
		}
		found = true
	}
	return found
}

func isEchoLikeLine(line string) bool {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return false
	}
	first := normalizeCommandToken(tokens)
	return first == "echo" || first == "printf"
}

func normalizeCommandToken(tokens []string) string {
	helpers := map[string]bool{"sudo": true, "env": true, "time": true}
	for len(tokens) > 0 {
		token := strings.ToLower(tokens[0])
		if helpers[token] {
			tokens = tokens[1:]
			continue
		}
		return token
	}
	return ""
}

func isSystemGeneratedBlock(block parsers.CodeBlock) bool {
	if strings.Contains(block.Content, "ie:auto-prereq") {
		return true
	}
	if strings.HasPrefix(block.Header, "Exporting variables defined via the CLI") {
		return true
	}
	return false
}

func extractFirstLine(block *ast.FencedCodeBlock, source []byte) string {
	if block == nil {
		return ""
	}
	lines := block.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}
	line := lines.At(0)
	return strings.TrimSpace(string(line.Value(source)))
}

func truncateSnippet(snippet string) string {
	snippet = strings.TrimSpace(snippet)
	if snippet == "" {
		return "<empty>"
	}
	if len(snippet) > 60 {
		return snippet[:60] + "..."
	}
	return snippet
}

func isEchoBeforeMatch(line string, matchStart int) bool {
	segment := commandSegmentBeforeMatch(line, matchStart)
	if segment == "" {
		return false
	}
	tokens := strings.Fields(segment)
	if len(tokens) == 0 {
		return false
	}
	cmd := normalizeCommandToken(tokens)
	return cmd == "echo" || cmd == "printf"
}

func commandSegmentBeforeMatch(line string, matchStart int) string {
	if matchStart <= 0 {
		return ""
	}
	prefix := line[:matchStart]
	cut := lastSeparatorIndex(prefix)
	if cut != -1 && cut < len(prefix) {
		prefix = prefix[cut:]
	}
	return strings.TrimSpace(prefix)
}

func lastSeparatorIndex(s string) int {
	separators := []string{"&&", "||", ";", "|"}
	last := -1
	for _, sep := range separators {
		if idx := strings.LastIndex(s, sep); idx != -1 {
			candidate := idx + len(sep)
			if candidate > last {
				last = candidate
			}
		}
	}
	return last
}
