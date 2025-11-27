package common

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/ui"
	"github.com/xrash/smetrics"
)

// Compares the actual output of a command to the expected output of a command.
func CompareCommandOutputs(
	actualOutput string,
	expectedOutput string,
	expectedSimilarity float64,
	expectedRegexPattern string,
	expectedOutputLanguage string,
) (float64, error) {
	actualNormalized := normalizeOutput(actualOutput)
	expectedNormalized := normalizeOutput(expectedOutput)

	if strings.TrimSpace(expectedRegexPattern) != "" {
		expandedPattern, compiledRegex, usedEnvValues, err := compileRegexWithEnv(expectedRegexPattern)
		if err != nil {
			return 0.0, err
		}

		if !compiledRegex.MatchString(actualNormalized) {
			patternDisplay := strings.TrimSpace(expectedRegexPattern)
			if patternDisplay == "" {
				patternDisplay = expandedPattern
			}
			if details := formatRegexEnvDetails(usedEnvValues); details != "" {
				patternDisplay = fmt.Sprintf("%s\n%s", patternDisplay, details)
			}
			return 0.0, fmt.Errorf(
				ui.ErrorMessageStyle.Render(
					"Expected output does not match actual output.\nExpected Pattern:\n%s\nActual:\n%s",
				),
				ui.VerboseStyle.Render(patternDisplay),
				ui.VerboseStyle.Render(summarizeOutput(actualNormalized, 20)),
			)
		}

		return 0.0, nil
	}

	if strings.ToLower(expectedOutputLanguage) == "json" {
		logging.GlobalLogger.Debugf(
			"Comparing JSON strings:\nExpected: %s\nActual%s",
			expectedNormalized,
			actualNormalized,
		)
		results, err := lib.CompareJsonStrings(actualNormalized, expectedNormalized, expectedSimilarity)
		if err != nil {
			return results.Score, err
		}

		logging.GlobalLogger.Debugf(
			"Expected Similarity: %f, Actual Similarity: %f",
			expectedSimilarity,
			results.Score,
		)

		if !results.AboveThreshold {
			return results.Score, fmt.Errorf(
				ui.ErrorMessageStyle.Render(
					"Expected output does not match actual output.\nExpected:\n%s\nActual:\n%s\nExpected Score:%s\nActual Score:%s",
				),
				ui.VerboseStyle.Render(summarizeOutput(expectedNormalized, 20)),
				ui.VerboseStyle.Render(summarizeOutput(actualNormalized, 20)),
				ui.VerboseStyle.Render(fmt.Sprintf("%f", expectedSimilarity)),
				ui.VerboseStyle.Render(fmt.Sprintf("%f", results.Score)),
			)
		}

		return results.Score, nil
	}

	// Default case, using similarity on non JSON block.
	score := smetrics.JaroWinkler(expectedNormalized, actualNormalized, 0.7, 4)

	if expectedSimilarity > score {
		return score, fmt.Errorf(
			ui.ErrorMessageStyle.Render(
				"Expected output does not match actual output.\nExpected:\n%s\nActual:\n%s\nExpected Score:%s\nActual Score:%s",
			),
			ui.VerboseStyle.Render(summarizeOutput(expectedNormalized, 20)),
			ui.VerboseStyle.Render(summarizeOutput(actualNormalized, 20)),
			ui.VerboseStyle.Render(fmt.Sprintf("%f", expectedSimilarity)),
			ui.VerboseStyle.Render(fmt.Sprintf("%f", score)),
		)
	}

	return score, nil
}

func normalizeOutput(value string) string {
	// Strip ANSI color codes first
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	value = ansiPattern.ReplaceAllString(value, "")

	// Then normalize line endings
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}

func compileRegexWithEnv(pattern string) (string, *regexp.Regexp, map[string]string, error) {
	expanded, used := expandRegexPattern(pattern)
	compiled, err := regexp.Compile(expanded)
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot compile regex %q: %w", expanded, err)
	}
	return expanded, compiled, used, nil
}

func expandRegexPattern(pattern string) (string, map[string]string) {
	replacements := loadEnvironmentForRegex()
	const literalPlaceholder = "__IE_LITERAL_DOLLAR__"
	pattern = strings.ReplaceAll(pattern, `\$`, literalPlaceholder)
	used := make(map[string]string)
	expanded := os.Expand(pattern, func(key string) string {
		if value, ok := replacements[key]; ok {
			used[key] = value
			return value
		}
		return ""
	})
	return strings.ReplaceAll(expanded, literalPlaceholder, "$"), used
}

func loadEnvironmentForRegex() map[string]string {
	replacements := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		replacements[parts[0]] = parts[1]
	}

	if envFromState, err := lib.LoadEnvironmentStateFile(lib.DefaultEnvironmentStateFile); err == nil {
		for k, v := range envFromState {
			replacements[k] = v
		}
	}

	return replacements
}

func summarizeOutput(value string, maxLines int) string {
	trimmed := strings.TrimRight(value, "\n")
	if trimmed == "" {
		return "<empty>"
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}

	summary := append([]string{}, lines[:maxLines]...)
	summary = append(summary, fmt.Sprintf("... (%d more lines)", len(lines)-maxLines))
	return strings.Join(summary, "\n")
}

func formatRegexEnvDetails(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, values[key]))
	}
	return fmt.Sprintf("(where %s)", strings.Join(parts, ", "))
}
