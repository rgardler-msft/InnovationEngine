package commands

import (
	"fmt"
	"os"

	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

const logPathEnvVar = "IE_LOG_PATH"

// The root command for the CLI. Currently initializes the logging for all other
// commands.
var rootCommand = &cobra.Command{
	Use:   "ie",
	Short: "The innovation engine.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevel, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return commandError(cmd, err, false, "error getting log level")
		}

		logPath, err := cmd.Flags().GetString("log-path")
		if err != nil {
			return commandError(cmd, err, false, "error getting log path")
		}
		if logPath == "" {
			if envPath := os.Getenv(logPathEnvVar); envPath != "" {
				logPath = envPath
			} else {
				logPath = logging.DefaultLogFile
			}
		}

		logging.Init(logging.LevelFromString(logLevel), logPath)

		if _, err := getEnvironmentSetting(cmd); err != nil {
			return commandError(cmd, err, false, "error resolving environment")
		}

		return nil
	},
}

// Entrypoint into the Innovation Engine CLI.
func ExecuteCLI() {
	rootCommand.PersistentFlags().
		String(
			"log-level",
			string(logging.Debug),
			"Set file logging level (trace|debug|info|warn|error|fatal). Controls entries written to ie.log; --verbose enriches interactive console separately",
		)
	rootCommand.PersistentFlags().
		String(
			"log-path",
			"",
			fmt.Sprintf("Path to ie log output (default %s, overridable via %s)", logging.DefaultLogFile, logPathEnvVar),
		)
	rootCommand.PersistentFlags().
		String(
			"environment",
			string(environments.EnvironmentsLocal),
			"The environment that the CLI is running in. Valid options are 'local', 'github-action'. For running ie in your standard terminal, local will work just fine. If using IE inside a github action, use github-action.",
		)

	rootCommand.PersistentFlags().
		StringArray(
			"feature",
			[]string{},
			"Enables the specified feature. Format: --feature <feature>",
		)

	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		logging.GlobalLogger.Errorf("Failed to execute ie: %s", err)
		os.Exit(1)
	}
}
