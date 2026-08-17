package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new context",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	message := args[0]
	fmt.Println("TODO: simpan entery:", message)
	return nil
}
