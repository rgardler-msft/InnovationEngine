package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Azure/InnovationEngine/internal/parsers"
)

func TestExecuteCodeBlockAsync_VerificationMismatchDoesNotFail(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	markerPath := filepath.Join(tmp, "marker")

	block := parsers.CodeBlock{
		Language: "bash",
		Content:  fmt.Sprintf("# ie:auto-prereq-verification marker=\"%s\" display=\"Test Prereq\"\necho \"Prerequisite needs to run\"\n", markerPath),
		ExpectedOutput: parsers.ExpectedOutputBlock{
			ExpectedRegex: regexp.MustCompile("already executed"),
		},
	}

	cmd := ExecuteCodeBlockAsync(block, map[string]string{})
	msg := cmd()

	if _, ok := msg.(SuccessfulCommandMessage); !ok {
		t.Fatalf("expected verification mismatch to be treated as success, got %T", msg)
	}

	if _, err := os.Stat(markerPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected marker %s to be absent after failed verification", markerPath)
	}
}

func TestExecuteCodeBlockAsync_VerificationSuccessCreatesMarker(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	markerPath := filepath.Join(tmp, "marker")
	display := "Skip Me"

	block := parsers.CodeBlock{
		Language: "bash",
		Content:  fmt.Sprintf("# ie:auto-prereq-verification marker=\"%s\" display=\"%s\"\necho \"validated\"\n", markerPath, display),
		ExpectedOutput: parsers.ExpectedOutputBlock{
			ExpectedRegex: regexp.MustCompile("validated"),
		},
	}

	cmd := ExecuteCodeBlockAsync(block, map[string]string{})
	msg := cmd()
	if _, ok := msg.(SuccessfulCommandMessage); !ok {
		t.Fatalf("expected verification success to be treated as success, got %T", msg)
	}

	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("expected marker file to exist: %v", err)
	}

	if string(data) != display {
		t.Fatalf("expected marker to contain %q, got %q", display, string(data))
	}
}

func TestExecuteCodeBlockAsync_VerificationCommandErrorDoesNotFail(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	markerPath := filepath.Join(tmp, "marker")

	block := parsers.CodeBlock{
		Language: "bash",
		Content:  fmt.Sprintf("# ie:auto-prereq-verification marker=\"%s\" display=\"Broken\"\nfalse\n", markerPath),
	}

	cmd := ExecuteCodeBlockAsync(block, map[string]string{})
	msg := cmd()

	if _, ok := msg.(SuccessfulCommandMessage); !ok {
		t.Fatalf("expected verification command failures to be treated as success, got %T", msg)
	}

	if _, err := os.Stat(markerPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected marker %s to be absent when verification failed", markerPath)
	}
}
