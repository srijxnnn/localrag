package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a file or directory to the index",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Adding: %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
