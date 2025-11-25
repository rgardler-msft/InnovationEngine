package commands

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/InnovationEngine/internal/engine"
	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/spf13/cobra"
)

func newExecutionTestCommand(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	addCommonExecutionFlags(cmd)
	addCorrelationFlag(cmd)
	cmd.PersistentFlags().String("environment", string(environments.EnvironmentsLocal), "")
	cmd.PersistentFlags().StringArray("feature", []string{}, "")
	cmd.PersistentFlags().String("report", "", "")
	cmd.Flags().AddFlagSet(cmd.PersistentFlags())
	return cmd
}

func mustSetFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err == nil {
		return
	}
	if err := cmd.PersistentFlags().Set(name, value); err != nil {
		t.Fatalf("failed to set flag %s: %v", name, err)
	}
}

func TestBindExecutionOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		configure   func(*testing.T, *cobra.Command)
		expectErr   bool
		errUser     bool
		errContains string
		assert      func(*testing.T, *executionOptions)
	}{
		{
			name: "defaults and render feature",
			args: []string{"scenario.md"},
			configure: func(t *testing.T, cmd *cobra.Command) {
				mustSetFlag(t, cmd, "verbose", "true")
				mustSetFlag(t, cmd, "stream-output", "false")
				mustSetFlag(t, cmd, "subscription", "sub")
				mustSetFlag(t, cmd, "working-directory", "/tmp")
				mustSetFlag(t, cmd, "correlation-id", "corr")
				mustSetFlag(t, cmd, "report", "report.json")
				mustSetFlag(t, cmd, "feature", "render-values")
				mustSetFlag(t, cmd, "var", "KEY=value")
			},
			assert: func(t *testing.T, opts *executionOptions) {
				if !opts.Verbose || opts.StreamOutput {
					t.Fatalf("expected verbose true and stream-output false, got %v/%v", opts.Verbose, opts.StreamOutput)
				}
				if opts.Subscription != "sub" || opts.WorkingDirectory != "/tmp" {
					t.Fatalf("subscription/workdir mismatch: %+v", opts)
				}
				if opts.CorrelationID != "corr" {
					t.Fatalf("expected correlation corr, got %s", opts.CorrelationID)
				}
				if !opts.RenderValues {
					t.Fatalf("expected render values to be true")
				}
				if opts.EnvironmentVariables["KEY"] != "value" {
					t.Fatalf("expected env var KEY=value, got %+v", opts.EnvironmentVariables)
				}
				if opts.ReportFile != "report.json" {
					t.Fatalf("expected report file to be set")
				}
			},
		},
		{
			name: "invalid environment variable format",
			args: []string{"scenario.md"},
			configure: func(t *testing.T, cmd *cobra.Command) {
				mustSetFlag(t, cmd, "var", "INVALID")
			},
			expectErr:   true,
			errUser:     true,
			errContains: "invalid --var assignment",
		},
		{
			name: "invalid feature",
			args: []string{"scenario.md"},
			configure: func(t *testing.T, cmd *cobra.Command) {
				mustSetFlag(t, cmd, "feature", "unknown")
			},
			expectErr:   true,
			errUser:     true,
			errContains: "invalid feature",
		},
		{
			name:        "missing markdown",
			args:        []string{},
			expectErr:   true,
			errUser:     true,
			errContains: "no markdown file specified",
		},
		{
			name: "invalid environment",
			args: []string{"scenario.md"},
			configure: func(t *testing.T, cmd *cobra.Command) {
				mustSetFlag(t, cmd, "environment", "invalid")
			},
			expectErr:   true,
			errUser:     false,
			errContains: "error resolving environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newExecutionTestCommand(t)
			if tt.configure != nil {
				tt.configure(t, cmd)
			}
			opts, err := bindExecutionOptions(cmd, tt.args)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				bindingErr, ok := err.(*optionBindingError)
				if !ok {
					t.Fatalf("expected optionBindingError, got %T", err)
				}
				if bindingErr.user != tt.errUser {
					t.Fatalf("expected user=%v, got %v", tt.errUser, bindingErr.user)
				}
				if tt.errContains != "" && !strings.Contains(bindingErr.message, tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, bindingErr.message)
				}
				return
			}
			if err != nil {
				t.Fatalf("bindExecutionOptions returned error: %v", err)
			}
			if opts.MarkdownPath != tt.args[0] {
				t.Fatalf("expected markdown path %s, got %s", tt.args[0], opts.MarkdownPath)
			}
			if tt.assert != nil {
				tt.assert(t, opts)
			}
		})
	}
}

func TestBuildEngineConfiguration(t *testing.T) {
	base := &executionOptions{
		MarkdownPath:         "scenario.md",
		Verbose:              true,
		DoNotDelete:          true,
		StreamOutput:         false,
		Subscription:         "sub",
		CorrelationID:        "corr",
		WorkingDirectory:     "/tmp",
		Environment:          environments.EnvironmentsLocal,
		RenderValues:         true,
		EnvironmentVariables: map[string]string{"k": "v"},
		ReportFile:           "report.json",
	}

	cfg := buildEngineConfiguration(base)
	if cfg.CorrelationId != base.CorrelationID || cfg.Subscription != base.Subscription {
		t.Fatalf("config fields did not match execution options: %+v", cfg)
	}
	if !cfg.RenderValues || cfg.DoNotDelete != base.DoNotDelete {
		t.Fatalf("expected render values true and do not delete true, got %+v", cfg)
	}

	override := buildEngineConfiguration(
		base,
		func(cfg *engine.EngineConfiguration) {
			cfg.DoNotDelete = false
			cfg.StreamOutput = true
		},
	)
	if override.DoNotDelete || !override.StreamOutput {
		t.Fatalf("overrides were not applied: %+v", override)
	}
	if override.Environment != environments.EnvironmentsLocal {
		t.Fatalf("environment should remain unchanged")
	}

	if !reflect.DeepEqual(cfg, buildEngineConfiguration(base)) {
		t.Fatalf("expected builder to be deterministic for same input")
	}
}
