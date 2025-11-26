package commands

import (
	"strings"
	"testing"
)

func TestInspectFailsWhenDescriptionMissing(t *testing.T) {
	content := "# Scenario\n\n## Step\n\n```bash\necho missing description\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err == nil {
		t.Fatalf("expected inspect to fail when code block description is missing")
	}
	if !strings.Contains(stderr.String(), "descriptive text") {
		t.Fatalf("expected description error, got %q", stderr.String())
	}
}

func TestInspectFailsWhenLanguageTagMissing(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nThis command lacks a language tag.\n\n```\necho no language\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err == nil {
		t.Fatalf("expected inspect to fail when language tag is missing")
	}
	if !strings.Contains(stderr.String(), "language tag") {
		t.Fatalf("expected language error, got %q", stderr.String())
	}
}

func TestInspectFailsWhenPrerequisiteMissingExpectedResults(t *testing.T) {
	content := "# Scenario\n\n## Prerequisites\n\nEnsure the environment is ready.\n\n```bash\necho verifying prereq\n```\n\n## Step\n\nRun the main step.\n\n```bash\necho done\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err == nil {
		t.Fatalf("expected inspect to fail when prerequisite lacks expected_results")
	}
	if !strings.Contains(stderr.String(), "expected_results") {
		t.Fatalf("expected expected_results error, got %q", stderr.String())
	}
}

func TestInspectAllowsPrerequisiteExportOnlyBlockWithoutExpectedResults(t *testing.T) {
	content := "# Scenario\n\n## Prerequisites\n\nConfigure environment variables.\n\n```bash\nexport AZ_APP_NAME=my-app\nexport AZ_LOCATION=eastus\n```\n\n## Step\n\nProceed with main execution.\n\n```bash\necho ready\n```\n"
	path := writeScenarioWithContent(t, content)
	_, _, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to allow export-only prerequisite blocks, got %v", err)
	}
}

func TestInspectAllowsMultipleEnvPrefixes(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nSet deployment values.\n\n```bash\nexport APP_NAME=demo\nexport OTHER_LOCATION=eastus\naz group create --name $APP_NAME --location $OTHER_LOCATION\n```\n"
	path := writeScenarioWithContent(t, content)
	_, _, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to allow mixed prefixes, got %v", err)
	}
}

func TestInspectWarnsWhenSimilarityOutOfRange(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nShow output.\n\n```bash\necho hi\n```\n\n<!-- expected_similarity=2.0 -->\n```text\nhi\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to succeed with warning, got %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "warnings detected") {
		t.Fatalf("expected pre-outline warning notice, got %q", output)
	}
	if !strings.Contains(output, "Similarity") && !strings.Contains(output, "expected_similarity") {
		t.Fatalf("expected detailed warning about similarity, got %q", output)
	}
}

func TestInspectFailsWhenPrerequisiteMissingFile(t *testing.T) {
	content := "# Scenario\n\n## Prerequisites\n\n- [Missing](./does-not-exist.md)\n\n## Step\n\nExecute main step.\n\n```bash\necho hi\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err == nil {
		t.Fatalf("expected inspect to fail for missing prerequisite file")
	}
	if !strings.Contains(stderr.String(), "Prerequisite") {
		t.Fatalf("expected missing prerequisite error, got %q", stderr.String())
	}
}

func TestInspectWarnsWhenExportUnused(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nExport without usage.\n\n```bash\nexport APP_NAME=demo\necho $APP_NAME\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to succeed with warning, got %v", err)
	}
	if !strings.Contains(stderr.String(), "never referenced") {
		t.Fatalf("expected unused export warning, got %q", stderr.String())
	}
}

func TestInspectFailsWhenEnvReferencedWithoutExport(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nUse undefined variable.\n\n```bash\necho $MISSING_VAR\n```\n"
	path := writeScenarioWithContent(t, content)
	_, stderr, err := runRootWithArgsCapturing(t, "inspect", path)
	if err == nil {
		t.Fatalf("expected inspect to fail when environment variable is referenced without export")
	}
	if !strings.Contains(stderr.String(), "is referenced but never exported") {
		t.Fatalf("expected undefined env error, got %q", stderr.String())
	}
}

func TestInspectAllowsLocalAssignmentsForEnvReferences(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nAssign and read variable.\n\n```bash\nAI_ENABLED=$(echo true)\nif [ \"$AI_ENABLED\" != \"true\" ]; then\n  exit 1\nfi\n```\n"
	path := writeScenarioWithContent(t, content)
	_, _, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to allow locally assigned variables, got %v", err)
	}
}

func TestInspectAllowsLowercaseLocalVariables(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nLoop with lowercase iterator.\n\n```bash\nfor i in 1 2 3; do\n  echo $i\ndone\n```\n"
	path := writeScenarioWithContent(t, content)
	_, _, err := runRootWithArgsCapturing(t, "inspect", path)
	if err != nil {
		t.Fatalf("expected inspect to allow lowercase local variables, got %v", err)
	}
}

func TestInspectReportsSuccessMessageWhenNoFindings(t *testing.T) {
	content := "# Scenario\n\n## Step\n\nSay hello.\n\n```bash\necho hello\n```\n"
	stdout, stderr, err := runRootWithArgsCapturing(t, "inspect", writeScenarioWithContent(t, content))
	if err != nil {
		t.Fatalf("expected inspect to succeed for clean scenario, got %v", err)
	}
	if strings.Contains(stdout.String(), "Scenario") {
		t.Fatalf("did not expect scenario outline when not verbose, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Inspection passed: no validation issues found.") {
		t.Fatalf("expected success message, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
