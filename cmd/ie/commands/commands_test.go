package commands

import (
	"bytes"
	"testing"

	"github.com/Azure/InnovationEngine/internal/logging"
)

// helper to execute the root Cobra command with args and capture any
// execution error. This exercises the full wiring from rootCommand down to
// subcommands and their RunE handlers.
func runRootWithArgs(t *testing.T, args ...string) error {
	t.Helper()
	// Ensure root-level flags are registered before each Execute call.
	// In production this is done by ExecuteCLI, but tests call Execute
	// directly on rootCommand.
	rootCommand.ResetFlags()
	rootCommand.PersistentFlags().
		String(
			"log-level",
			string(logging.Debug),
			"",
		)
	rootCommand.PersistentFlags().
		String(
			"environment",
			"local",
			"",
		)

	rootCommand.SetArgs(args)
	rootCommand.SetOut(&bytes.Buffer{})
	rootCommand.SetErr(&bytes.Buffer{})
	return rootCommand.Execute()
}

func TestExecuteCommand_MissingMarkdownFile(t *testing.T) {
	// When required args are missing, Cobra Arg validation should fail and
	// return an error before reaching the RunE handler.
	err := runRootWithArgs(t, "execute")
	if err == nil {
		t.Fatalf("expected error when markdown file is missing, got nil")
	}
}

func TestExecuteCommand_InvalidEnvVarFormat(t *testing.T) {
	err := runRootWithArgs(t, "execute", "doc.md", "--var", "INVALID")
	if err == nil {
		t.Fatalf("expected error for invalid --var format, got nil")
	}
}

func TestTestCommand_MissingMarkdownFile(t *testing.T) {
	// When required args are missing, Cobra Arg validation should fail and
	// return an error before reaching the RunE handler.
	err := runRootWithArgs(t, "test")
	if err == nil {
		t.Fatalf("expected error when markdown file is missing, got nil")
	}
}

func TestTestCommand_InvalidEnvVarFormat(t *testing.T) {
	err := runRootWithArgs(t, "test", "doc.md", "--var", "INVALID")
	if err == nil {
		t.Fatalf("expected error for invalid --var format, got nil")
	}
}

func TestInteractiveCommand_MissingMarkdownFile(t *testing.T) {
	// When required args are missing, Cobra Arg validation should fail and
	// return an error before reaching the RunE handler.
	err := runRootWithArgs(t, "interactive")
	if err == nil {
		t.Fatalf("expected error when markdown file is missing, got nil")
	}
}

func TestInteractiveCommand_InvalidEnvVarFormat(t *testing.T) {
	err := runRootWithArgs(t, "interactive", "doc.md", "--var", "INVALID")
	if err == nil {
		t.Fatalf("expected error for invalid --var format, got nil")
	}
}

func TestInspectCommand_MissingMarkdownFile(t *testing.T) {
	err := runRootWithArgs(t, "inspect")
	if err == nil {
		t.Fatalf("expected error when markdown file is missing, got nil")
	}
}

func TestInspectCommand_InvalidEnvVarFormat(t *testing.T) {
	err := runRootWithArgs(t, "inspect", "doc.md", "--var", "INVALID")
	if err == nil {
		t.Fatalf("expected error for invalid --var format, got nil")
	}
}

func TestClearEnvCommand_ForceNoFiles(t *testing.T) {
	// Should succeed (or at least not error) when forcing clear with no existing files.
	err := runRootWithArgs(t, "clear-env", "--force")
	if err != nil {
		t.Fatalf("expected no error for clear-env --force with no files, got %v", err)
	}
}
