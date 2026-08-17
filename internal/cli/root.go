package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Context server : connect git activity with team context",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "showing detail output")
}
