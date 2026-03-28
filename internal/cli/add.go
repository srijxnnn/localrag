package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/srijxnnn/localrag/internal/chunk"
	"github.com/srijxnnn/localrag/internal/ollama"
	"github.com/srijxnnn/localrag/internal/store"
)

const (
	defaultOllamaURL = "http://localhost:11434"
	embedModel       = "nomic-embed-text"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a file or directory to the index",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(store.DBPath()); err != nil {
			return fmt.Errorf("workspace not initialized; run: rag init")
		}
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		chunks, err := chunk.FromFile(absPath)
		if err != nil {
			return err
		}
		s, err := store.Init(store.DBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		docID, err := s.SaveDocument(absPath)
		if err != nil {
			return err
		}

		client := ollama.New(defaultOllamaURL)
		for _, c := range chunks {
			vec, err := client.Embed(embedModel, c)
			if err != nil {
				return err
			}
			if err := s.SaveChunk(docID, c, vec); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
