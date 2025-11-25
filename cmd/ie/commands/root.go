package commands

import (
	"fmt"
	"os"

	"github.com/Azure/InnovationEngine/internal/engine/environments"
	"github.com/Azure/InnovationEngine/internal/logging"
	"github.com/spf13/cobra"
)

// The root command for the CLI. Currently initializes the logging for all other
// commands.
var rootCommand = &cobra.Command{
	Use:   "ie",
	Short: "The innovation engine.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevel, err := cmd.Flags().GetString("log-level")
		if err != nil {
			fmt.Printf("Error getting log level: %s", err)
			return fmt.Errorf("error getting log level: %w", err)
		}
		logging.Init(logging.LevelFromString(logLevel))

		// Check environment
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			fmt.Printf("Error getting environment: %s", err)
			logging.GlobalLogger.Errorf("Error getting environment: %s", err)
			return fmt.Errorf("error getting environment: %w", err)
		}

		if !environments.IsValidEnvironment(environment) {
			fmt.Printf("Invalid environment: %s", environment)
			logging.GlobalLogger.Errorf("Invalid environment: %s", environment)
			return fmt.Errorf("invalid environment: %s", environment)
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
			"environment",
			environments.EnvironmentsLocal,
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
