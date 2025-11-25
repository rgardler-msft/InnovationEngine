package commands

import (
	"encoding/json"
	"fmt"
	"strings"

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
		markdownFile := args[0]
		if markdownFile == "" {
			return commandError(cmd, nil, true, "no markdown file specified")
		}

		environmentVariables, _ := cmd.Flags().GetStringArray("var")

		environmentSetting, err := getEnvironmentSetting(cmd)
		if err != nil {
			return commandError(cmd, err, false, "error resolving environment")
		}

		// Parse the environment variables
		cliEnvironmentVariables := make(map[string]string)
		for _, environmentVariable := range environmentVariables {
			keyValuePair := strings.SplitN(environmentVariable, "=", 2)
			if len(keyValuePair) != 2 {
				return commandError(cmd, nil, true, "invalid environment variable format: %s", environmentVariable)
			}

			cliEnvironmentVariables[keyValuePair[0]] = keyValuePair[1]
		}

		// Parse the markdown file and create a scenario
		scenario, err := common.CreateScenarioFromMarkdown(
			markdownFile,
			[]string{"bash", "azurecli", "azurecli-interactive", "terraform"},
			cliEnvironmentVariables)
		if err != nil {
			return commandError(cmd, err, false, "error creating scenario")
		}

		// If within cloudshell, we need to wrap the script in a json object to
		// communicate it to the portal.
		if environmentSetting.IsAzureLike() {
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
	toBashCommand.PersistentFlags().
		StringArray("var", []string{}, "Sets an environment variable for the scenario. Format: --var <key>=<value>")
}
