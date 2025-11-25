package commands

import "github.com/spf13/cobra"

// addCommonExecutionFlags adds flags that are shared across execution-style
// commands such as execute, interactive, and test.
func addCommonExecutionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().
		Bool("verbose", false, "Show extra console context (working dirs, full command output). For deeper persisted diagnostics use --log-level")
	cmd.PersistentFlags().
		Bool("do-not-delete", false, "Do not delete the Azure resources created by the Azure CLI commands executed.")
	cmd.PersistentFlags().
		Bool("stream-output", true, "Stream command output in real-time as it's generated (default). Use --stream-output=false to show spinner and display output after completion.")

	cmd.PersistentFlags().
		String("subscription", "", "Sets the subscription ID used by a scenarios azure-cli commands. Will rely on the default subscription if not set.")
	cmd.PersistentFlags().
		String("working-directory", ".", "Sets the working directory for innovation engine to operate out of. Restores the current working directory when finished.")

	cmd.PersistentFlags().
		StringArray("var", []string{}, "Sets an environment variable for the scenario. Format: --var <key>=<value>")
}

// addCorrelationFlag adds the correlation-id flag used by some commands.
func addCorrelationFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().
		String("correlation-id", "", "Adds a correlation ID to the user agent used by a scenarios azure-cli commands.")
}
