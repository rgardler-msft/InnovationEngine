package commands

import (
	"errors"
	"fmt"

	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

// commandError logs, surfaces, and wraps CLI errors consistently. When
// showHelp is true, the command's help text is printed after the error.
func commandError(cmd *cobra.Command, err error, showHelp bool, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)

	if err != nil {
		logging.GlobalLogger.Errorf("%s: %v", message, err)
		cmd.PrintErrf("Error: %s: %v\n", message, err)
		err = fmt.Errorf("%s: %w", message, err)
	} else {
		logging.GlobalLogger.Error(message)
		cmd.PrintErrf("Error: %s\n", message)
		err = errors.New(message)
	}

	if showHelp {
		cmd.PrintErrln()
		_ = cmd.Help()
	}

	return err
}
