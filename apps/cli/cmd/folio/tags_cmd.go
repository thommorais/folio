package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

type tagCount struct {
	Tag     string `json:"tag"`
	Total   int    `json:"total"`
	Todos   int    `json:"todos"`
	Plans   int    `json:"plans"`
	Journal int    `json:"journal"`
	Docs    int    `json:"docs"`
	Known   bool   `json:"known"`
}

func tagsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Show the tags in use across a project",
		Long: `Tally every tag on the project's todos, plans, logs and docs, so the
vocabulary already in use is visible before adding to it.

Tags carry the context a title cannot. A tag says where the work lives
(frontend, backend, cli) and what sort of work it is (bug, refactor, decision),
which is what makes ` + "`--tags frontend`" + ` narrow a whole project to one surface.
Write commands accept only known tags, so a typo is caught where it is typed
instead of becoming a tag nothing will query again.

  folio tags
  folio tags --unused
  folio tags --json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			counts, err := collectTags(folio, project)
			if err != nil {
				return err
			}
			return renderTags(counts)
		},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.Flags().BoolVar(&flagUnusedTags, "unused", false, "also list known tags nothing carries yet")

	return cmd
}

var flagUnusedTags bool

// collectTags tallies client-side: there is no counts endpoint, and a tag
// census is a whole-project question by nature, so the four lists are the
// honest cost rather than four aggregate calls that do not exist.
func collectTags(folio *client.Client, project string) ([]tagCount, error) {
	byTag := map[string]*tagCount{}

	add := func(tags []string, bump func(*tagCount)) {
		for _, tag := range tags {
			entry, seen := byTag[tag]
			if !seen {
				entry = &tagCount{Tag: tag, Known: isKnownTag(tag)}
				byTag[tag] = entry
			}
			entry.Total++
			bump(entry)
		}
	}

	todos, err := folio.ListTodos(project, client.TodoFilter{})
	if err != nil {
		return nil, err
	}
	for _, todo := range todos {
		add(todo.Tags, func(c *tagCount) { c.Todos++ })
	}

	plans, err := folio.ListPlans(project, client.PlanFilter{})
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		add(plan.Tags, func(c *tagCount) { c.Plans++ })
	}

	logs, err := folio.ListJournal(project, client.JournalFilter{})
	if err != nil {
		return nil, err
	}
	for _, entry := range logs {
		add(entry.Tags, func(c *tagCount) { c.Journal++ })
	}

	docs, err := folio.ListDocs(project, client.DocFilter{})
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		add(doc.Tags, func(c *tagCount) { c.Docs++ })
	}

	if flagUnusedTags {
		for _, known := range knownTags() {
			if _, seen := byTag[known]; !seen {
				byTag[known] = &tagCount{Tag: known, Known: true}
			}
		}
	}

	counts := make([]tagCount, 0, len(byTag))
	for _, entry := range byTag {
		counts = append(counts, *entry)
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Total != counts[j].Total {
			return counts[i].Total > counts[j].Total
		}
		return counts[i].Tag < counts[j].Tag
	})
	return counts, nil
}

func renderTags(counts []tagCount) error {
	if flagJSON {
		return encode(counts)
	}

	if len(counts) == 0 {
		fmt.Println("no tags yet")
		fmt.Fprintln(os.Stderr, "\ncontext: "+strings.Join(contextTags, ", "))
		fmt.Fprintln(os.Stderr, "kind:    "+strings.Join(kindTags, ", "))
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "TAG\tTOTAL\tTODOS\tPLANS\tLOGS\tDOCS")
	for _, entry := range counts {
		tag := entry.Tag
		// An unknown tag predates the vocabulary or came in through the API;
		// marking it is how it gets noticed and retired.
		if !entry.Known {
			tag += " *"
		}
		fmt.Fprintf(out, "%s\t%d\t%d\t%d\t%d\t%d\n", tag, entry.Total, entry.Todos, entry.Plans, entry.Journal, entry.Docs)
	}
	if err := out.Flush(); err != nil {
		return err
	}

	for _, entry := range counts {
		if !entry.Known {
			fmt.Fprintln(os.Stderr, "\n* not in the known vocabulary; `folio tags --unused` lists what is")
			break
		}
	}
	return nil
}
