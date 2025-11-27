package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentVariableValidationAndFiltering(t *testing.T) {
	// Test key validation
	t.Run("Key Validation", func(t *testing.T) {
		validCases := []struct {
			key      string
			expected bool
		}{
			{"ValidKey", true},
			{"VALID_VARIABLE", true},
			{"_AnotherValidKey", true},
			{"123Key", false},                   // Starts with a digit
			{"key-with-hyphen", false},          // Contains a hyphen
			{"key.with.dot", false},             // Contains a period
			{"Fabric_NET-0-[Delegated]", false}, // From cloud shell environment.
		}

		for _, tc := range validCases {
			t.Run(tc.key, func(t *testing.T) {
				result := environmentVariableName.MatchString(tc.key)
				if result != tc.expected {
					t.Errorf(
						"Expected isValidKey(%s) to be %v, got %v",
						tc.key,
						tc.expected,
						result,
					)
				}
			})
		}
	})

	// Test key filtering
	t.Run("Key Filtering", func(t *testing.T) {
		envMap := map[string]string{
			"ValidKey":                 "value1",
			"_AnotherValidKey":         "value2",
			"123Key":                   "value3",
			"key-with-hyphen":          "value4",
			"key.with.dot":             "value5",
			"Fabric_NET-0-[Delegated]": "false", // From cloud shell environment.
		}

		validEnvMap := filterInvalidKeys(envMap)

		expectedValidEnvMap := map[string]string{
			"ValidKey":         "value1",
			"_AnotherValidKey": "value2",
		}

		if len(validEnvMap) != len(expectedValidEnvMap) {
			t.Errorf(
				"Expected validEnvMap to have %d keys, got %d",
				len(expectedValidEnvMap),
				len(validEnvMap),
			)
		}

		for key, value := range validEnvMap {
			if expectedValue, ok := expectedValidEnvMap[key]; !ok || value != expectedValue {
				t.Errorf("Expected validEnvMap[%s] to be %s, got %s", key, expectedValue, value)
			}
		}
	})
}

func TestFilterAgainstBaseline(t *testing.T) {
	current := map[string]string{
		"EV_ALPHA": "1",
		"PATH":     "/custom/bin",
		"HOME":     "/home/user",
	}
	baseline := map[string]string{
		"PATH": "/usr/bin",
		"HOME": "/home/user",
	}

	filtered := filterAgainstBaseline(current, baseline)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(filtered))
	}
	if filtered["EV_ALPHA"] != "1" {
		t.Fatalf("expected EV_ALPHA to survive filtering")
	}
	if filtered["PATH"] != "/custom/bin" {
		t.Fatalf("expected PATH difference to be retained")
	}
	if _, exists := filtered["HOME"]; exists {
		t.Fatalf("did not expect unchanged HOME to survive filtering")
	}
}

func TestFilterEnvironmentStateFile(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "env-vars")
	baselinePath := filepath.Join(tempDir, "env-vars.baseline")

	if err := os.WriteFile(statePath, []byte("PATH=/usr/bin\nEV_BETA=demo\n"), 0600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}
	if err := os.WriteFile(baselinePath, []byte("PATH=/usr/bin\n"), 0600); err != nil {
		t.Fatalf("failed to write baseline file: %v", err)
	}

	if err := FilterEnvironmentStateFile(statePath, baselinePath); err != nil {
		t.Fatalf("filtering failed: %v", err)
	}

	filtered, err := LoadEnvironmentStateFile(statePath)
	if err != nil {
		t.Fatalf("failed to reload filtered state: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected only one entry after filtering, got %d", len(filtered))
	}
	if filtered["EV_BETA"] != "demo" {
		t.Fatalf("expected EV_BETA to remain, got %v", filtered)
	}
	if _, exists := filtered["PATH"]; exists {
		t.Fatalf("PATH should have been removed by filtering")
	}
}
