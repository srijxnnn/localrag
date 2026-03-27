package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a RAG workspace in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initialized")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
