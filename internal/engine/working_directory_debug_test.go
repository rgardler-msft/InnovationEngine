package engine

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/Azure/InnovationEngine/internal/parsers"
)

// Test that when verbosity or debug logging is enabled we output the working directory
// before executing a command block.
func TestWorkingDirectoryDebugLine(t *testing.T) {
	// Initialize logger at debug level to exercise the condition.
	logging.Init(logging.Debug)

	// Build a minimal scenario step with a single bash command.
	step := common.Step{
		Name: "Test Step",
		CodeBlocks: []parsers.CodeBlock{
			{
				Language:       "bash",
				Content:        "echo hello world",
				Header:         "Test",
				ExpectedOutput: parsers.ExpectedOutputBlock{},
			},
		},
	}

	e, err := NewEngine(EngineConfiguration{Verbose: true, Environment: "local"})
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Capture stdout.
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	execErr := e.ExecuteAndRenderSteps([]common.Step{step}, map[string]string{})

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout
	_, _ = buf.ReadFrom(r)

	if execErr != nil {
		t.Fatalf("execution failed: %v", execErr)
	}

	if !strings.Contains(buf.String(), "Working directory:") {
		t.Fatalf("expected working directory debug line, got output: %s", buf.String())
	}
}
