package commands

import (
	"errors"
	"fmt"

	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/spf13/cobra"
)

const optionBindingFailureMessage = "error binding command options"

type executionOptions struct {
	MarkdownPath         string
	Verbose              bool
	DoNotDelete          bool
	StreamOutput         bool
	Subscription         string
	CorrelationID        string
	WorkingDirectory     string
	Environment          environments.Environment
	RenderValues         bool
	EnvironmentVariables map[string]string
	ReportFile           string
}

type optionBindingError struct {
	user    bool
	message string
	err     error
}

func (e *optionBindingError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *optionBindingError) Unwrap() error {
	return e.err
}

func bindExecutionOptions(cmd *cobra.Command, args []string) (*executionOptions, error) {
	if len(args) == 0 || args[0] == "" {
		return nil, newOptionBindingError(true, "no markdown file specified", nil)
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	doNotDelete, err := cmd.Flags().GetBool("do-not-delete")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	streamOutput, err := cmd.Flags().GetBool("stream-output")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	subscription, err := cmd.Flags().GetString("subscription")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	workingDirectory, err := cmd.Flags().GetString("working-directory")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	correlationID, err := getOptionalStringFlag(cmd, "correlation-id")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	reportFile, err := getOptionalStringFlag(cmd, "report")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	environmentSetting, err := getEnvironmentSetting(cmd)
	if err != nil {
		return nil, newOptionBindingError(false, "error resolving environment", err)
	}

	rawVariables, err := cmd.Flags().GetStringArray("var")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	parsedVariables, err := lib.ParseEnvironmentVariableAssignments(rawVariables)
	if err != nil {
		return nil, newOptionBindingError(true, "invalid --var assignment", err)
	}

	features, err := cmd.Flags().GetStringArray("feature")
	if err != nil {
		return nil, newOptionBindingError(false, optionBindingFailureMessage, err)
	}

	renderValues, err := shouldRenderValues(features)
	if err != nil {
		return nil, err
	}

	return &executionOptions{
		MarkdownPath:         args[0],
		Verbose:              verbose,
		DoNotDelete:          doNotDelete,
		StreamOutput:         streamOutput,
		Subscription:         subscription,
		CorrelationID:        correlationID,
		WorkingDirectory:     workingDirectory,
		Environment:          environmentSetting,
		RenderValues:         renderValues,
		EnvironmentVariables: parsedVariables,
		ReportFile:           reportFile,
	}, nil
}

func handleExecutionOptionError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	var bindingErr *optionBindingError
	if errors.As(err, &bindingErr) {
		return commandError(cmd, bindingErr.err, bindingErr.user, bindingErr.message)
	}
	return commandError(cmd, err, false, optionBindingFailureMessage)
}

func getOptionalStringFlag(cmd *cobra.Command, name string) (string, error) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return "", nil
	}
	return cmd.Flags().GetString(name)
}

func shouldRenderValues(features []string) (bool, error) {
	renderValues := false
	for _, feature := range features {
		switch feature {
		case "render-values":
			renderValues = true
		default:
			return false, newOptionBindingError(true, fmt.Sprintf("invalid feature: %s", feature), nil)
		}
	}
	return renderValues, nil
}

func newOptionBindingError(user bool, message string, err error) *optionBindingError {
	return &optionBindingError{user: user, message: message, err: err}
}
