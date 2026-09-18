package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

func searchCommand() *cobra.Command {
	var query client.SearchQuery
	var kinds, tags string

	cmd := &cobra.Command{
		Use:   "search [term]",
		Short: "Search a project's journal, docs, todos and plans",
		Long: `Search across all four kinds at once, newest first, each hit carrying a
snippet so it is judgeable without a second call.

  folio search fts5
  folio search rules --kind journal,doc
  folio search --tags decision
  folio search "index strategy" --limit 5 --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				query.Text = args[0]
			}
			if kinds != "" {
				query.Kinds = strings.Split(kinds, ",")
			}
			if tags != "" {
				query.Tags = strings.Split(tags, ",")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			hits, err := folio.Search(project, query)
			if err != nil {
				return err
			}
			noteIfPaged(len(hits), query.Limit)
			return renderHits(hits)
		},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.Flags().StringVar(&kinds, "kind", "", "comma separated: journal,doc,todo,plan")
	cmd.Flags().StringVar(&tags, "tags", "", "comma separated tags")
	registerTagCompletion(cmd)
	cmd.Flags().IntVar(&query.Limit, "limit", 0, "maximum hits")
	cmd.Flags().IntVar(&query.Offset, "offset", 0, "hits to skip")

	return cmd
}

func renderHits(hits []client.SearchHit) error {
	if flagJSON {
		return encode(hits)
	}

	if len(hits) == 0 {
		fmt.Println("no matches")
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "KIND\tID\tTITLE")
	for _, hit := range hits {
		fmt.Fprintf(out, "%s\t%s\t%s\n", hit.Kind, hit.ID, oneLine(hit.Title))
	}
	if err := out.Flush(); err != nil {
		return err
	}

	// Snippets go under the table rather than in a column: they are long
	// enough that a fifth column would wrap and break the alignment.
	for _, hit := range hits {
		if hit.Snippet != "" {
			fmt.Printf("\n%s\n  %s\n", oneLine(hit.Title), oneLine(hit.Snippet))
		}
	}
	return nil
}

func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
