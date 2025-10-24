package commands

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/Azure/InnovationEngine/internal/lib"
)

// Test that when a stale working directory state file exists from a previous run,
// invoking the CLI logic overwrites it with the original invocation directory.
func TestOverwriteStaleWorkingDirectoryState(t *testing.T) {
    // Create a temporary directory to act as invocation dir.
    tempDir, err := os.MkdirTemp("", "ie-invocation-test-*")
    if err != nil {
        t.Fatalf("failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Simulate starting in the temp dir.
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("failed to chdir to temp dir: %v", err)
    }

    // Manually set the global (init would normally do this).
    OriginalInvocationDirectory = tempDir

    // Write a stale state file pointing somewhere else.
    staleDir := filepath.Join(tempDir, "stale")
    if err := os.WriteFile(lib.DefaultWorkingDirectoryStateFile, []byte(staleDir+"\n"), 0644); err != nil {
        t.Fatalf("failed to write stale state file: %v", err)
    }

    // Execute the overwrite logic: mimic snippet from execute.go.
    if err := lib.SaveWorkingDirectoryStateFile(lib.DefaultWorkingDirectoryStateFile, OriginalInvocationDirectory); err != nil {
        t.Fatalf("failed to save working directory state: %v", err)
    }

    // Read back the file and assert it matches the invocation directory.
    contents, err := os.ReadFile(lib.DefaultWorkingDirectoryStateFile)
    if err != nil {
        t.Fatalf("failed to read working directory state: %v", err)
    }
    if string(contents) != OriginalInvocationDirectory+"\n" {
        t.Fatalf("expected working directory state '%s', got '%s" , OriginalInvocationDirectory+"\n", string(contents))
    }
}
