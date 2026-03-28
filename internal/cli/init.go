package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/srijxnnn/localrag/internal/store"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a RAG workspace in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if fi, err := os.Stat(store.RAGDir); err == nil {
			if !fi.IsDir() {
				return fmt.Errorf("%s exists and is not a directory", store.RAGDir)
			}
			return fmt.Errorf("already initialized")
		}
		if err := os.Mkdir(store.RAGDir, 0755); err != nil {
			return err
		}
		s, err := store.Init(store.DBPath())
		if err != nil {
			_ = os.RemoveAll(store.RAGDir)
			return err
		}
		if err := s.Close(); err != nil {
			_ = os.RemoveAll(store.RAGDir)
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Initialized")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
