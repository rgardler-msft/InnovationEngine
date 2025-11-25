package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	enginepkg "github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	rootCommand.PersistentFlags().
		StringArray(
			"feature",
			[]string{},
			"",
		)
	resetCommandTreeFlags(rootCommand)

	rootCommand.SetArgs(args)
	rootCommand.SetOut(&bytes.Buffer{})
	rootCommand.SetErr(&bytes.Buffer{})
	return rootCommand.Execute()
}

// patchEngineNew swaps engine.NewEngine with a stub for the duration of a test.
func patchEngineNew(t *testing.T) *[]enginepkg.EngineConfiguration {
	original := engineNewEngine
	stubCalls := make([]enginepkg.EngineConfiguration, 0)
	engineNewEngine = func(cfg enginepkg.EngineConfiguration) (engineRunner, error) {
		stubCalls = append(stubCalls, cfg)
		return stubEngine{}, nil
	}
	t.Cleanup(func() {
		engineNewEngine = original
	})
	return &stubCalls
}

type stubEngine struct{}

func (stubEngine) ExecuteScenario(*common.Scenario) error { return nil }
func (stubEngine) TestScenario(*common.Scenario) error    { return nil }
func (stubEngine) InteractWithScenario(*common.Scenario) error {
	return nil
}

func resetCommandTreeFlags(cmd *cobra.Command) {
	resetFlags := func(flagSet *pflag.FlagSet) {
		flagSet.VisitAll(func(f *pflag.Flag) {
			if sliceValue, ok := f.Value.(pflag.SliceValue); ok {
				_ = sliceValue.Replace([]string{})
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		})
	}
	resetFlags(cmd.Flags())
	resetFlags(cmd.PersistentFlags())
	for _, child := range cmd.Commands() {
		resetCommandTreeFlags(child)
	}
}

func writeTempScenario(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.md")
	content := "# " + name + "\n\n## Step\n\n```bash\necho hello\n```\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp scenario: %v", err)
	}
	return path
}

func TestCommandsRequireMarkdownArgument(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"execute", []string{"execute"}},
		{"test", []string{"test"}},
		{"interactive", []string{"interactive"}},
		{"inspect", []string{"inspect"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runRootWithArgs(t, tt.args...)
			if err == nil {
				t.Fatalf("expected error for %s with missing markdown file", tt.args[0])
			}
		})
	}
}

func TestCommandsRejectInvalidEnvVarFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"execute", []string{"execute", "doc.md", "--var", "INVALID"}},
		{"test", []string{"test", "doc.md", "--var", "INVALID"}},
		{"interactive", []string{"interactive", "doc.md", "--var", "INVALID"}},
		{"inspect", []string{"inspect", "doc.md", "--var", "INVALID"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runRootWithArgs(t, tt.args...)
			if err == nil {
				t.Fatalf("expected error for %s with invalid --var", tt.args[0])
			}
		})
	}
}

func TestClearEnvCommand_ForceNoFiles(t *testing.T) {
	// Should succeed (or at least not error) when forcing clear with no existing files.
	err := runRootWithArgs(t, "clear-env", "--force")
	if err != nil {
		t.Fatalf("expected no error for clear-env --force with no files, got %v", err)
	}
}

func TestRootCommandRejectsInvalidEnvironment(t *testing.T) {
	err := runRootWithArgs(t, "clear-env", "--force", "--environment", "invalid-env")
	if err == nil {
		t.Fatalf("expected error for invalid environment flag")
	}
}

func TestRootCommandAcceptsKnownEnvironment(t *testing.T) {
	err := runRootWithArgs(t, "clear-env", "--force", "--environment", string(environments.EnvironmentsAzure))
	if err != nil {
		t.Fatalf("expected azure environment to be accepted, got %v", err)
	}
}

func TestInspectCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Temp Scenario")
	if err := runRootWithArgs(t, "inspect", markdown); err != nil {
		t.Fatalf("inspect command should succeed for valid scenario, got %v", err)
	}
}

func TestExecuteCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Execute Scenario")
	calls := patchEngineNew(t)
	if err := runRootWithArgs(t, "execute", markdown); err != nil {
		t.Fatalf("execute command should succeed, got %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", len(*calls))
	}
	if (*calls)[0].Environment != environments.EnvironmentsLocal {
		t.Fatalf("expected environment to default to local, got %s", (*calls)[0].Environment)
	}
}

func TestTestCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Test Scenario")
	calls := patchEngineNew(t)
	if err := runRootWithArgs(t, "test", markdown); err != nil {
		t.Fatalf("test command should succeed, got %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", len(*calls))
	}
	if (*calls)[0].ReportFile != "" {
		t.Fatalf("expected default test command report file to be empty, got %s", (*calls)[0].ReportFile)
	}
}

func TestInteractiveCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Interactive Scenario")
	calls := patchEngineNew(t)
	if err := runRootWithArgs(t, "interactive", markdown); err != nil {
		t.Fatalf("interactive command should succeed, got %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", len(*calls))
	}
	if !(*calls)[0].StreamOutput {
		t.Fatalf("interactive command should force StreamOutput, got %v", (*calls)[0].StreamOutput)
	}
}
