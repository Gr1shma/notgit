package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/spf13/cobra"
)

var (
	globalConfigFlag bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Get and set repository or global options",
	Long:  "Manage configuration values for notgit, including user identity, editor, and default branch.",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  getConfigCallback,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  setConfigCallback,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key> [<key>...]",
	Short: "Remove one or more configuration values",
	Args:  cobra.MinimumNArgs(1),
	RunE:  unsetConfigCallback,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Args:  cobra.NoArgs,
	RunE:  listConfigCallback,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in the editor",
	Args:  cobra.NoArgs,
	RunE:  editConfigCallback,
}

func init() {
	configCmd.PersistentFlags().BoolVarP(&globalConfigFlag, "global", "g", false, "Use global configuration file")
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configEditCmd)

	configCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Root().HelpFunc()(cmd, args)
		fmt.Print(utils.PrintSupportedConfigKeys())
	})

	rootCmd.AddCommand(configCmd)
}

func getConfigCallback(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, _, err := utils.LoadConfig(globalConfigFlag)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	val, err := utils.GetConfigKeyValue(cfg, key)
	if err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), val)
	return nil
}

func setConfigCallback(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, path, err := utils.LoadConfig(globalConfigFlag)
	if err != nil {
		return fmt.Errorf("failed loading config: %w", err)
	}

	err = utils.SetConfigKeyValue(cfg, path, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config key: %w", err)
	}

	if err := cfg.SaveTo(path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
	return nil
}

func unsetConfigCallback(cmd *cobra.Command, args []string) error {
	cfg, path, err := utils.LoadConfig(globalConfigFlag)
	if err != nil {
		return fmt.Errorf("Error loading config: %w", err)

	}

	var anyUnset bool
	for _, key := range args {
		unset, err := utils.UnsetConfigKey(cfg, path, key)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error unsetting key %s: %v\n", key, err)
			continue
		}
		if unset {
			fmt.Fprintf(cmd.OutOrStdout(), "Unset key: %s\n", key)
			anyUnset = true
		}
	}

	if anyUnset {
		if err := cfg.SaveTo(path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	return nil
}

func listConfigCallback(cmd *cobra.Command, args []string) error {
	cfg, path, err := utils.LoadConfig(globalConfigFlag)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Listing configuration from: %s\n\n", path)
	utils.PrintAllConfig(cfg)
	return nil
}

func editConfigCallback(cmd *cobra.Command, args []string) error {
	cfg, configPath, err := utils.LoadConfig(globalConfigFlag)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	editor, err := utils.GetConfigKeyValue(cfg, "core.editor")
	if err != nil || editor == "" {
		editor = os.Getenv("EDITOR")
	}

	if editor == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "No editor configured. Set core.editor or $EDITOR environment variable.")
		return nil
	}

	tryEditor := func(ed string) error {
		cmdExec := exec.Command(ed, configPath)
		cmdExec.Stdin = os.Stdin
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		return cmdExec.Run()
	}

	if err := tryEditor(editor); err != nil {
		fallbackEditor := os.Getenv("EDITOR")
		if fallbackEditor != "" && fallbackEditor != editor {
			fmt.Printf("Failed to open editor %q, falling back to $EDITOR=%q\n", editor, fallbackEditor)
			if err := tryEditor(fallbackEditor); err != nil {
				return fmt.Errorf("failed to open fallback editor: %w", err)
			}
		} else {
			return fmt.Errorf("failed to open editor %q: %w", editor, err)
		}
	}
	return nil
}
