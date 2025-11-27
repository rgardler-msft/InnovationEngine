package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	enginepkg "github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// helper to execute the root Cobra command with args and capture any
// execution error. This exercises the full wiring from rootCommand down to
// subcommands and their RunE handlers.
func runRootWithArgs(t *testing.T, args ...string) error {
	t.Helper()
	_, _, err := runRootWithArgsCapturing(t, args...)
	return err
}

func runRootWithArgsCapturing(t *testing.T, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
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
			"log-path",
			"",
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
	tempLogPath := filepath.Join(t.TempDir(), "ie.log")
	if err := rootCommand.PersistentFlags().Set("log-path", tempLogPath); err != nil {
		t.Fatalf("failed to set log path flag: %v", err)
	}
	t.Setenv(logPathEnvVar, tempLogPath)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCommand.SetArgs(args)
	rootCommand.SetOut(stdout)
	rootCommand.SetErr(stderr)
	return stdout, stderr, rootCommand.Execute()
}

// patchEngineNew swaps engine.NewEngine with a stub for the duration of a test.

func patchEngineNew(t *testing.T) *int {
	original := engineNewEngine
	callCount := 0
	engineNewEngine = func(cfg enginepkg.EngineConfiguration) (engineRunner, error) {
		callCount++
		return stubEngine{}, nil
	}
	t.Cleanup(func() {
		engineNewEngine = original
	})
	return &callCount
}

func captureEngineConfigurations(t *testing.T) *[]enginepkg.EngineConfiguration {
	original := buildEngineConfiguration
	captured := make([]enginepkg.EngineConfiguration, 0)
	buildEngineConfiguration = func(opts *executionOptions, overrides ...func(*enginepkg.EngineConfiguration)) enginepkg.EngineConfiguration {
		cfg := original(opts, overrides...)
		captured = append(captured, cfg)
		return cfg
	}
	t.Cleanup(func() {
		buildEngineConfiguration = original
	})
	return &captured
}

type stubEngine struct{}

func (stubEngine) ExecuteScenario(*common.Scenario) error { return nil }
func (stubEngine) TestScenario(*common.Scenario) error    { return nil }
func (stubEngine) InteractWithScenario(*common.Scenario) error {
	return nil
}

type failingExecuteEngine struct {
	err error
}

func (f failingExecuteEngine) ExecuteScenario(*common.Scenario) error { return f.err }
func (f failingExecuteEngine) TestScenario(*common.Scenario) error    { return nil }
func (f failingExecuteEngine) InteractWithScenario(*common.Scenario) error {
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
	content := fmt.Sprintf(
		"# %s\n\n## Step\n\nThis step echoes hello.\n\n```bash\necho hello\n```\n",
		name,
	)
	return writeScenarioWithContent(t, content)
}

func writeScenarioWithContent(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp scenario: %v", err)
	}
	return path
}

func overrideDefaultStateFiles(t *testing.T) (string, string) {
	t.Helper()
	originalEnv := lib.DefaultEnvironmentStateFile
	originalWD := lib.DefaultWorkingDirectoryStateFile
	newEnv := filepath.Join(t.TempDir(), "env-vars")
	newWD := filepath.Join(t.TempDir(), "working-dir")
	lib.DefaultEnvironmentStateFile = newEnv
	lib.DefaultWorkingDirectoryStateFile = newWD
	if flag := envConfigCommand.Flags().Lookup("state-file"); flag != nil {
		flag.DefValue = newEnv
		_ = flag.Value.Set(newEnv)
	}
	t.Cleanup(func() {
		lib.DefaultEnvironmentStateFile = originalEnv
		lib.DefaultWorkingDirectoryStateFile = originalWD
		if flag := envConfigCommand.Flags().Lookup("state-file"); flag != nil {
			flag.DefValue = originalEnv
			_ = flag.Value.Set(originalEnv)
		}
	})
	return newEnv, newWD
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

func TestEnvConfigCommand_PrintsExports(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "env-vars")
	data := "EV_ALPHA=\"one\"\nEV_BETA=\"two words\"\n"
	if err := os.WriteFile(stateFile, []byte(data), 0600); err != nil {
		t.Fatalf("failed to seed env file: %v", err)
	}

	stdout, _, err := runRootWithArgsCapturing(t, "env-config", "--state-file", stateFile)
	if err != nil {
		t.Fatalf("env-config should succeed, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "export EV_ALPHA=\"one\"") {
		t.Fatalf("expected EV_ALPHA export, got %q", output)
	}
	alphaIdx := strings.Index(output, "export EV_ALPHA")
	betaIdx := strings.Index(output, "export EV_BETA")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("expected both exports in output: %q", output)
	}
	if !(alphaIdx < betaIdx) {
		t.Fatalf("expected alphabetical ordering, got %q", output)
	}
}

func TestEnvConfigReadsDefaultSnapshotAfterExecute(t *testing.T) {
	envSnapshot, _ := overrideDefaultStateFiles(t)
	scenario := writeScenarioWithContent(t, "# Scenario\n\n## Step\n\n```bash\nexport AZ_TEST_VAR=demo\n```\n")
	if err := runRootWithArgs(t, "clear-env", "--force"); err != nil {
		t.Fatalf("clear-env failed: %v", err)
	}
	if err := runRootWithArgs(t, "execute", scenario); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if _, err := os.Stat(envSnapshot); err != nil {
		t.Fatalf("expected environment snapshot at %s: %v", envSnapshot, err)
	}
	stdout, _, err := runRootWithArgsCapturing(t, "env-config")
	if err != nil {
		t.Fatalf("env-config failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "export AZ_TEST_VAR=\"demo\"") {
		t.Fatalf("expected AZ_TEST_VAR export, got %q", stdout.String())
	}
}

func TestEnvConfigCommand_MissingFileErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, _, err := runRootWithArgsCapturing(t, "env-config", "--state-file", missing); err == nil {
		t.Fatalf("expected error when env state file is missing")
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
	configs := captureEngineConfigurations(t)
	callCount := patchEngineNew(t)
	if err := runRootWithArgs(t, "execute", markdown); err != nil {
		t.Fatalf("execute command should succeed, got %v", err)
	}
	if *callCount != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", *callCount)
	}
	if len(*configs) != 1 {
		t.Fatalf("expected engine configuration to be built once, got %d", len(*configs))
	}
	if (*configs)[0].Environment != environments.EnvironmentsLocal {
		t.Fatalf("expected environment to default to local, got %s", (*configs)[0].Environment)
	}
}

func TestExecuteCommandScenarioFailureDoesNotPrintUsage(t *testing.T) {
	markdown := writeTempScenario(t, "Execute Failure Scenario")
	original := engineNewEngine
	engineNewEngine = func(cfg enginepkg.EngineConfiguration) (engineRunner, error) {
		return failingExecuteEngine{err: errors.New("boom")}, nil
	}
	t.Cleanup(func() {
		engineNewEngine = original
	})

	_, stderr, err := runRootWithArgsCapturing(t, "execute", markdown)
	if err == nil {
		t.Fatalf("expected error when scenario execution fails")
	}
	if strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected stderr to omit usage, got %q", stderr.String())
	}
}

func TestTestCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Test Scenario")
	configs := captureEngineConfigurations(t)
	callCount := patchEngineNew(t)
	if err := runRootWithArgs(t, "test", markdown); err != nil {
		t.Fatalf("test command should succeed, got %v", err)
	}
	if *callCount != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", *callCount)
	}
	if len(*configs) != 1 {
		t.Fatalf("expected engine configuration to be built once, got %d", len(*configs))
	}
	if (*configs)[0].ReportFile != "" {
		t.Fatalf("expected default test command report file to be empty, got %s", (*configs)[0].ReportFile)
	}
}

func TestInteractiveCommand_Succeeds(t *testing.T) {
	markdown := writeTempScenario(t, "Interactive Scenario")
	configs := captureEngineConfigurations(t)
	callCount := patchEngineNew(t)
	if err := runRootWithArgs(t, "interactive", markdown); err != nil {
		t.Fatalf("interactive command should succeed, got %v", err)
	}
	if *callCount != 1 {
		t.Fatalf("expected engine.NewEngine to be called once, got %d", *callCount)
	}
	if len(*configs) != 1 {
		t.Fatalf("expected engine configuration to be built once, got %d", len(*configs))
	}
	if !(*configs)[0].StreamOutput {
		t.Fatalf("interactive command should force StreamOutput, got %v", (*configs)[0].StreamOutput)
	}
}
