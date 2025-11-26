package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/InnovationEngine/internal/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(envConfigCommand)
	envConfigCommand.Flags().String(
		"state-file",
		lib.DefaultEnvironmentStateFile,
		"Path to the environment state file to read",
	)
	envConfigCommand.Flags().String(
		"prefix",
		"",
		"Only emit variables that begin with the supplied prefix",
	)
}

var envConfigCommand = &cobra.Command{
	Use:   "env-config",
	Short: "Print stored environment variables as source-able exports",
	Long: `Reads the persisted environment state file (default /tmp/env-vars)
and renders its contents as export statements. Capture the output and source it
later to reproduce the environment from a previous Innovation Engine run.

Examples:
  ie env-config                            # Dump all persisted variables
  ie env-config --prefix EV_               # Limit output to EV_ prefixed vars
  ie env-config --state-file /tmp/custom   # Use a custom state file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stateFile, err := cmd.Flags().GetString("state-file")
		if err != nil {
			return commandError(cmd, err, false, "error parsing --state-file")
		}

		prefix, err := cmd.Flags().GetString("prefix")
		if err != nil {
			return commandError(cmd, err, false, "error parsing --prefix")
		}

		envMap, err := lib.LoadEnvironmentStateFile(stateFile)
		if err != nil {
			return commandError(cmd, err, false, "error loading environment state")
		}

		sanitized := lib.SanitizeEnvironmentMap(envMap)
		exports := buildExportLines(sanitized, prefix)
		writer := cmd.OutOrStdout()
		if len(exports) == 0 {
			fmt.Fprintln(writer, "# No persisted environment variables matched the requested filters.")
			return nil
		}

		for _, line := range exports {
			fmt.Fprintln(writer, line)
		}

		return nil
	},
}

func buildExportLines(values map[string]string, prefix string) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	exports := make([]string, 0, len(keys))
	for _, k := range keys {
		exports = append(exports, fmt.Sprintf("export %s=%s", k, strconv.Quote(values[k])))
	}
	return exports
}
