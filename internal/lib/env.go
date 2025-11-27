package lib

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/InnovationEngine/internal/lib/fs"
)

// Get environment variables from the current process.
func GetEnvironmentVariables() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}

	return envMap
}

// ParseEnvironmentVariableAssignments converts CLI-provided KEY=VALUE strings into a
// map. It returns an error if any entry is not in the expected KEY=VALUE
// format. This helper is intended for use by commands that accept repeated
// environment variable flags such as --var.
func ParseEnvironmentVariableAssignments(assignments []string) (map[string]string, error) {
	env := make(map[string]string)

	for _, assignment := range assignments {
		pair := strings.SplitN(assignment, "=", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid environment variable format: %s", assignment)
		}

		key := strings.TrimSpace(pair[0])
		value := pair[1]

		if key == "" {
			return nil, fmt.Errorf("environment variable name is empty in assignment: %s", assignment)
		}

		env[key] = value
	}

	return env, nil
}

// Location where the environment and working directory state from commands are
// to be captured and sent to for being able to share state across commands.
var DefaultEnvironmentStateFile = "/tmp/ie-env-vars"
var DefaultWorkingDirectoryStateFile = "/tmp/working-dir"

// BaselineEnvironmentStateFile returns the file path that stores the original
// process environment used to filter persisted values. Callers can pass a
// custom state file path; otherwise the default is used.
func BaselineEnvironmentStateFile(stateFile string) string {
	if strings.TrimSpace(stateFile) == "" {
		stateFile = DefaultEnvironmentStateFile
	}
	return stateFile + ".baseline"
}

// SaveEnvironmentBaselineFile captures the provided environment map and writes
// it to the baseline file corresponding to the supplied state file path.
func SaveEnvironmentBaselineFile(stateFile string, env map[string]string) error {
	baselinePath := BaselineEnvironmentStateFile(stateFile)
	return writeEnvironmentStateFile(baselinePath, filterInvalidKeys(env))
}

// FilterEnvironmentStateFile keeps only the variables whose values differ from
// the baseline environment. This ensures we persist/export only the variables
// introduced or modified by the executable document itself.
func FilterEnvironmentStateFile(stateFile, baselineFile string) error {
	if strings.TrimSpace(stateFile) == "" {
		stateFile = DefaultEnvironmentStateFile
	}

	if !fs.FileExists(stateFile) {
		return nil
	}

	currentEnv, err := LoadEnvironmentStateFile(stateFile)
	if err != nil {
		return err
	}

	baselineEnv := make(map[string]string)
	if strings.TrimSpace(baselineFile) != "" && fs.FileExists(baselineFile) {
		baselineEnv, err = LoadEnvironmentStateFile(baselineFile)
		if err != nil {
			return err
		}
	}

	filtered := filterAgainstBaseline(currentEnv, baselineEnv)
	filtered = filterInvalidKeys(filtered)
	return writeEnvironmentStateFile(stateFile, filtered)
}

// Loads a file that contains environment variables
func LoadEnvironmentStateFile(path string) (map[string]string, error) {
	if !fs.FileExists(path) {
		return nil, fmt.Errorf("env file '%s' does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file '%s': %w", path, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	env := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2) // Split at the first "=" only
			value := parts[1]
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				// Remove leading and trailing quotes
				value = value[1 : len(value)-1]
			}
			env[parts[0]] = value
		}
	}
	return env, nil
}

func CleanEnvironmentStateFile(path string) error {
	env, err := LoadEnvironmentStateFile(path)
	if err != nil {
		return err
	}

	env = filterInvalidKeys(env)
	return writeEnvironmentStateFile(path, env)
}

var environmentVariableName = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func filterInvalidKeys(envMap map[string]string) map[string]string {
	validEnvMap := make(map[string]string)
	for key, value := range envMap {
		if environmentVariableName.MatchString(key) {
			validEnvMap[key] = value
		}
	}
	return validEnvMap
}

func filterAgainstBaseline(current map[string]string, baseline map[string]string) map[string]string {
	if len(current) == 0 {
		return current
	}

	filtered := make(map[string]string)
	for key, value := range current {
		if baseValue, exists := baseline[key]; !exists || baseValue != value {
			filtered[key] = value
		}
	}

	return filtered
}

func writeEnvironmentStateFile(path string, env map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := fmt.Fprintf(writer, "%s=\"%s\"\n", key, env[key]); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// SanitizeEnvironmentMap removes any keys that are not valid shell environment
// variable names. Callers that surface persisted environment data should run it
// before emitting shell-friendly output.
func SanitizeEnvironmentMap(envMap map[string]string) map[string]string {
	return filterInvalidKeys(envMap)
}

// Deletes the stored environment variables file.
func DeleteEnvironmentStateFile(path string) error {
	return os.Remove(path)
}

// Loads a file that contains the working directory
func LoadWorkingDirectoryStateFile(path string) (string, error) {
	if !fs.FileExists(path) {
		return "", fmt.Errorf("working directory file '%s' does not exist", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read working directory file '%s': %w", path, err)
	}

	workingDir := strings.TrimSpace(string(content))
	return workingDir, nil
}

// Deletes the stored working directory file.
func DeleteWorkingDirectoryStateFile(path string) error {
	return os.Remove(path)
}

// SaveWorkingDirectoryStateFile writes the provided working directory path to
// the state file, creating or truncating it as needed. This allows callers to
// explicitly override any previously persisted working directory before the
// first command executes.
func SaveWorkingDirectoryStateFile(path string, workingDir string) error {
	if workingDir == "" {
		return fmt.Errorf("working directory is empty")
	}
	// Ensure trailing newline for consistency with pwd > file behavior.
	content := workingDir
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}
