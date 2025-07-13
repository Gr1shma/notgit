package commands

import (
	"fmt"

	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/spf13/cobra"
)

type ConfigArgs struct {
	userName  string
	userEmail string
}

var configArgs = &ConfigArgs{}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Get and set repository or global options",
	Long:  "Manage configuration values for notgit, including user identity, editor, and default branch.",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run:   getConfigCallback,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run:   setConfigCallback,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key> [<key>...]",
	Short: "Remove one or more configuration values",
	Args:  cobra.MinimumNArgs(1),
	Run:   unsetConfigCallback,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Args:  cobra.NoArgs,
	Run:   listConfigCallback,
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUnsetCmd)

	configCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Root().HelpFunc()(cmd, args)
		fmt.Print(utils.PrintSupportedConfigKeys())
	})

	rootCmd.AddCommand(configCmd)
}

func getConfigCallback(cmd *cobra.Command, args []string) {
	key := args[0]

	cfg, _, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	val, err := utils.GetConfigKeyValue(cfg, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(val)

}

func setConfigCallback(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	cfg, path, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	err = utils.SetConfigKeyValue(cfg, path, key, value)
	if err != nil {
		fmt.Printf("Error setting config: %v\n", err)
		return
	}

	if err := cfg.SaveTo(path); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		return
	}

	fmt.Printf("Set %s = %s\n", key, value)
}

func unsetConfigCallback(cmd *cobra.Command, args []string) {
	cfg, path, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	var anyUnset bool
	for _, key := range args {
		keySetUnsetBool, err := utils.UnsetConfigKey(cfg, path, key)
		if err != nil {
			fmt.Printf("Error unsetting key %s: %v\n", key, err)
			continue
		}
		fmt.Printf("Unset key: %s\n", key)
		anyUnset = keySetUnsetBool
	}

	if anyUnset {
		if err := cfg.SaveTo(path); err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
		}
	}
}

func listConfigCallback(cmd *cobra.Command, args []string) {
	cfg, path, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Listing configuration from: %s\n\n", path)
	utils.PrintAllConfig(cfg)
}
