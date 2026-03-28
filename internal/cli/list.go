package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/srijxnnn/localrag/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List indexed documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(store.DBPath()); err != nil {
			return fmt.Errorf("workspace not initialized; run: rag init")
		}
		s, err := store.Init(store.DBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		docs, err := s.ListDocuments()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if len(docs) == 0 {
			fmt.Fprintln(out, "No indexed documents. Run: rag add <file>")
			return nil
		}

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tCHUNKS\tADDED")
		for _, d := range docs {
			added := time.Unix(d.AddedAt, 0).Format(time.RFC3339)
			fmt.Fprintf(w, "%s\t%d\t%s\n", d.Path, d.ChunkCount, added)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
