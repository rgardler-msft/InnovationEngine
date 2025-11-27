package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/InnovationEngine/internal/lib"
)

func TestCompareCommandOutputsRegexExpandsOsEnv(t *testing.T) {
	t.Setenv("IE_REGEX_VALUE", "World")

	if _, err := CompareCommandOutputs(
		"Hello World",
		"",
		0,
		"^Hello $IE_REGEX_VALUE$",
		"",
	); err != nil {
		t.Fatalf("expected regex comparison to succeed, got error: %v", err)
	}
}

func TestCompareCommandOutputsRegexUsesStateFileEnv(t *testing.T) {
	unsetEnv(t, "IE_REGEX_VALUE")

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "env-state")

	original := lib.DefaultEnvironmentStateFile
	lib.DefaultEnvironmentStateFile = stateFile
	t.Cleanup(func() {
		lib.DefaultEnvironmentStateFile = original
	})

	if err := os.WriteFile(stateFile, []byte("IE_REGEX_VALUE=\"World\"\n"), 0600); err != nil {
		t.Fatalf("failed to seed env state file: %v", err)
	}

	if _, err := CompareCommandOutputs(
		"Hello World",
		"",
		0,
		"^Hello $IE_REGEX_VALUE$",
		"",
	); err != nil {
		t.Fatalf("expected regex comparison to use state file env, got error: %v", err)
	}
}

func TestCompareCommandOutputsRegexErrorIncludesEnvValues(t *testing.T) {
	t.Setenv("IE_REGEX_VALUE", "World")

	if _, err := CompareCommandOutputs(
		"Hello there",
		"",
		0,
		"^Hello $IE_REGEX_VALUE$",
		"",
	); err == nil {
		t.Fatalf("expected regex comparison to fail")
	} else if !strings.Contains(err.Error(), "(where IE_REGEX_VALUE=World)") {
		t.Fatalf("expected error to include env value, got: %v", err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, ok := os.LookupEnv(key)
	if ok {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset env %s: %v", key, err)
		}
		t.Cleanup(func() {
			os.Setenv(key, value)
		})
		return
	}
	t.Cleanup(func() {
		os.Unsetenv(key)
	})
}
