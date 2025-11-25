package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/InnovationEngine/internal/engine/common"
	"github.com/spf13/cobra"
)

type AzureScript struct {
	Script string `json:"script"`
}

var toBashCommand = &cobra.Command{
	Use:   "to-bash",
	Short: "Convert a markdown scenario into a bash script.",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := bindExecutionOptions(cmd, args)
		if err != nil {
			return handleExecutionOptionError(cmd, err)
		}

		// Parse the markdown file and create a scenario
		scenario, err := common.CreateScenarioFromMarkdown(
			opts.MarkdownPath,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			opts.EnvironmentVariables)
		if err != nil {
			return commandError(cmd, err, false, "error creating scenario")
		}

		// If within cloudshell, we need to wrap the script in a json object to
		// communicate it to the portal.
		if opts.Environment.IsAzureLike() {
			script := AzureScript{Script: scenario.ToShellScript()}
			scriptJson, err := json.Marshal(script)
			if err != nil {
				return commandError(cmd, err, false, "error converting to json")
			}

			fmt.Printf("ie_us%sie_ue\n", scriptJson)
		} else {
			fmt.Printf("%s", scenario.ToShellScript())
		}

		return nil
	},
}

func init() {
	rootCommand.AddCommand(toBashCommand)
	addCommonExecutionFlags(toBashCommand)
}
