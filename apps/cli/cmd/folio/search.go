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
	var all bool

	cmd := &cobra.Command{
		Use:   "search [term]",
		Short: "Search tickets, todos, plans, docs, journal, work logs and knowledge",
		Long: `Search every kind at once, ranked by relevance, each hit carrying a snippet
so it is judgeable without a second call.

Knowledge is not scoped to a project, so it answers from wherever you are.

  folio search fts5
  folio search rules --kind journal,doc
  folio search railway --kind knowledge
  folio search --tags decision
  folio search "index strategy" --all --limit 5 --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
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

			// --all ranks one result set over every project the caller can
			// read; per-project queries could not be compared, since bm25
			// scores are only meaningful within a single query.
			search := func() ([]client.SearchHit, error) {
				if all {
					return folio.SearchAll(query)
				}
				project, err := resolveProject()
				if err != nil {
					return nil, err
				}
				return folio.Search(project, query)
			}

			hits, err := search()
			if err != nil {
				return err
			}
			noteIfPaged(len(hits), query.Limit)
			return renderHits(hits)
		},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.Flags().BoolVar(&all, "all", false, "search every project you can read, not just the current one")
	cmd.Flags().StringVar(&kinds, "kind", "", "comma separated: ticket,todo,plan,doc,journal,worklog,resolution,decision,knowledge")
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
	// The project column earns its place only when a result can span more
	// than one, which is what --all does.
	spansProjects := false
	for _, hit := range hits {
		if hit.ProjectSlug != hits[0].ProjectSlug {
			spansProjects = true
			break
		}
	}

	if spansProjects {
		fmt.Fprintln(out, "KIND\tPROJECT\tID\tTITLE")
	} else {
		fmt.Fprintln(out, "KIND\tID\tTITLE")
	}
	for _, hit := range hits {
		if spansProjects {
			project := hit.ProjectSlug
			if project == "" {
				project = "-"
			}
			fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", hit.Kind, project, hit.ID, oneLine(hit.Title))
			continue
		}
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
