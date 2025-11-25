package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIIntegrationExecuteCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	const skipEnv = "IE_SKIP_DOC_SCENARIOS"
	if os.Getenv(skipEnv) == "1" {
		t.Skipf("skipping doc scenario integration because %s=1", skipEnv)
	}

	repoRoot := findRepoRoot(t)
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "ie")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ie")
	buildCmd.Dir = repoRoot
	buildCmd.Env = append(os.Environ(), "GO111MODULE=on", "GOARCH="+runtime.GOARCH, "GOOS="+runtime.GOOS)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build ie binary: %v", err)
	}

	scenarioPath := filepath.Join(repoRoot, "scenarios", "testing", "test.md")
	if _, err := os.Stat(scenarioPath); err != nil {
		t.Fatalf("failed to locate scenario at %s: %v", scenarioPath, err)
	}
	logPath := filepath.Join(binDir, "ie.log")

	runCmd := exec.Command(
		binaryPath,
		"--log-level", "info",
		"--environment", "local",
		"execute", scenarioPath,
	)
	runCmd.Dir = binDir
	runCmd.Env = append(os.Environ(), "IE_LOG_PATH="+logPath)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ie execute failed: %v\nOutput: %s", err, string(output))
	}
	t.Logf("ie output:\n%s", string(output))

	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected ie.log to be created at %s: %v", logPath, err)
	}

	if !strings.Contains(string(output), "Master test scenario completed.") {
		t.Fatalf("expected command output to contain scenario command, got: %s", string(output))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	return dir
}
