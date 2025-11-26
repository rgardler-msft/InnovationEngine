package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

// commandError logs, surfaces, and wraps CLI errors consistently. When
// showHelp is true, the command's help text is printed after the error.
func commandError(cmd *cobra.Command, err error, showHelp bool, format string, args ...interface{}) error {
	if cmd != nil {
		// Silence Cobra's automatic usage printing; we handle help output explicitly.
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}
	message := fmt.Sprintf(format, args...)
	writer := errorWriter(cmd)

	if err != nil {
		logging.GlobalLogger.Errorf("%s: %v", message, err)
		fmt.Fprintf(writer, "Error: %s\n", message)
		printErrorDetails(writer, err)
		err = fmt.Errorf("%s: %w", message, err)
	} else {
		logging.GlobalLogger.Error(message)
		fmt.Fprintf(writer, "Error: %s\n", message)
		err = errors.New(message)
	}

	if showHelp {
		cmd.PrintErrln()
		_ = cmd.Help()
	}

	return err
}

func printErrorDetails(writer io.Writer, err error) {
	if err == nil {
		return
	}

	trimmed := strings.TrimSpace(err.Error())
	if trimmed == "" {
		return
	}

	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fmt.Fprintf(writer, "  %s\n", line)
	}
}

func errorWriter(cmd *cobra.Command) io.Writer {
	if cmd != nil {
		return cmd.ErrOrStderr()
	}
	return os.Stderr
}
