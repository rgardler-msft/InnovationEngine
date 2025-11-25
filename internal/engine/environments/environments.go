package environments

import "fmt"

// Environment represents the execution environment for the engine and CLI.
// It is a typed wrapper around the underlying string value to reduce the
// chance of passing invalid values around the codebase.
type Environment string

const (
	EnvironmentsLocal        Environment = "local"
	EnvironmentsGithubAction Environment = "github-action"
	EnvironmentsOCD          Environment = "ocd"
	EnvironmentsAzure        Environment = "azure"
)

// Check if the environment is valid.
func IsValidEnvironment(environment string) bool {
	switch Environment(environment) {
	case EnvironmentsLocal,
		EnvironmentsGithubAction,
		EnvironmentsOCD,
		EnvironmentsAzure:
		return true
	default:
		return false
	}
}

// IsAzureEnvironment reports whether the provided environment string
// corresponds to an Azure-based execution environment.
func IsAzureEnvironment(environment string) bool {
	switch Environment(environment) {
	case EnvironmentsAzure, EnvironmentsOCD:
		return true
	default:
		return false
	}
}

// Parse converts a raw string into a typed Environment. It returns an error
// if the provided value is not one of the known environments.
func Parse(environment string) (Environment, error) {
	if !IsValidEnvironment(environment) {
		return "", fmt.Errorf("invalid environment: %s", environment)
	}
	return Environment(environment), nil
}
