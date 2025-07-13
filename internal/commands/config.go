package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
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

	configData, _, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	section, subkey, err := utils.SplitKey(key)
	if err != nil {
		fmt.Printf("Invalid config key: %v\n", err)
		return
	}

	value := configData.Section(section).Key(subkey).String()
	if value == "" {
		fmt.Printf("Key %s not set.\n", key)
		return
	}

	fmt.Println(value)
}

func setConfigCallback(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	section, subkey, err := utils.SplitKey(key)
	if err != nil {
		fmt.Printf("Error while splitting the keys: %v\n", err)
		return
	}

	configData, configPath, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error while loading config: %v\n", err)
		return
	}

	configData.Section(section).Key(subkey).SetValue(value)
	if err := configData.SaveTo(configPath); err != nil {
		fmt.Printf("failed to save config: %v", err)
	}
	fmt.Printf("set %s = %s\n", key, value)
}

func unsetConfigCallback(cmd *cobra.Command, args []string) {
	configData, configPath, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	var anyUnset bool

	for _, key := range args {
		section, subkey, err := utils.SplitKey(key)
		if err != nil {
			fmt.Printf("Invalid config key %q: %v\n", key, err)
			continue
		}

		sec := configData.Section(section)
		if sec == nil || !sec.HasKey(subkey) {
			fmt.Printf("Key %s is not set.\n", key)
			continue
		}

		sec.DeleteKey(subkey)
		fmt.Printf("Unset key: %s\n", key)
		anyUnset = true

		if len(sec.Keys()) == 0 && !strings.EqualFold(sec.Name(), ini.DefaultSection) {
			configData.DeleteSection(section)
			fmt.Printf("Removed empty section: [%s]\n", section)
		}
	}

	if anyUnset {
		if err := configData.SaveTo(configPath); err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
		}
	} else {
		fmt.Println("No keys were unset.")
	}
}

func listConfigCallback(cmd *cobra.Command, args []string) {
	configData, configPath, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Listing configuration from: %s\n\n", configPath)

	for _, section := range configData.Sections() {
		if strings.EqualFold(section.Name(), ini.DefaultSection) {
			continue
		}

		for _, key := range section.Keys() {
			fmt.Printf("%s.%s = %s\n", section.Name(), key.Name(), key.Value())
		}
	}
}
