package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "notgit",
	Short: "A Git clone in Go — it's like Git, but worse.",
	Long: `notgit is a command-line tool that reimplements Git's core features in Go.
It’s a hands-on project built during the journey of learning Go — simple, clear, and intentionally minimal.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
