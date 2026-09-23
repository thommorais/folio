package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

// The knowledge base is the one command with no --project: a note is readable
// and writable by every signed-in user, and the project it names is a label
// saying where it came from rather than a scope.
func knowledgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kb",
		Short:   "Save and find durable knowledge: tips, snippets, fixes worth keeping",
		Aliases: []string{"knowledge"},
		Long: `Knowledge outlives the project it was learned on, so it is not scoped to one
and any signed-in user can read and write it.

  folio kb add "Enable realtime in PocketBase on Railway" --body - --tags pocketbase,railway
  folio kb list --tags pocketbase
  folio kb get pocketbase-realtime-on-railway
  folio search railway

A note is also indexed for search, so it answers from inside any project.`,
	}

	cmd.AddCommand(
		knowledgeListCommand(), knowledgeGetCommand(), knowledgeAddCommand(),
		knowledgeUpdateCommand(), knowledgeDeleteCommand(),
	)
	return cmd
}

func knowledgeListCommand() *cobra.Command {
	var filter client.KnowledgeFilter
	var tags string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List knowledge, newest first",
		Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if tags != "" {
				filter.Tags = strings.Split(tags, ",")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			notes, err := folio.ListKnowledge(filter)
			if err != nil {
				return err
			}
			noteIfPaged(len(notes), filter.Limit)
			return renderKnowledgeList(notes)
		},
	}

	cmd.Flags().StringVar(&tags, "tags", "", "comma separated tags")
	cmd.Flags().StringVar(&filter.ProjectID, "project", "", "only notes filed under this project")
	cmd.Flags().BoolVar(&filter.Unattached, "unattached", false, "only notes filed under no project")
	cmd.Flags().StringVarP(&filter.Search, "query", "q", "", "match the title and body")
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, "maximum rows")
	cmd.Flags().IntVar(&filter.Offset, "offset", 0, "rows to skip")

	return cmd
}

func knowledgeGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-slug>",
		Short: "Show one note with its body",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			note, err := folio.GetKnowledge(args[0])
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(note)
			}
			return renderKnowledgeDetail(note)
		},
	}
}

func knowledgeAddCommand() *cobra.Command {
	var slug, body, project, tags string

	cmd := &cobra.Command{
		Use:     "add <title>",
		Short:   "Save a note, with --body - to read markdown from stdin",
		Aliases: []string{"create", "new"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(body)
			if err != nil {
				return err
			}

			in := client.KnowledgeInput{Title: &args[0]}
			setIf(&in.Slug, slug)
			setIf(&in.Body, text)
			setIf(&in.ProjectID, project)
			setFreeTags(&in.Tags, tags)

			folio, err := api()
			if err != nil {
				return err
			}

			note, err := folio.CreateKnowledge(in)
			if err != nil {
				return err
			}
			return renderKnowledge(note)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "derived from the title when omitted")
	cmd.Flags().StringVar(&body, "body", "", "markdown body, or - for stdin")
	cmd.Flags().StringVar(&project, "project", "", "note where it was learned; does not restrict who can read it")
	cmd.Flags().StringVar(&tags, "tags", "", "comma separated; free-form, not the project tag vocabulary")

	return cmd
}

func knowledgeUpdateCommand() *cobra.Command {
	var title, slug, body, project, tags string

	cmd := &cobra.Command{
		Use:   "update <id-or-slug>",
		Short: "Update a note, leaving unset fields alone",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(body)
			if err != nil {
				return err
			}

			in := client.KnowledgeInput{}
			setIf(&in.Title, title)
			setIf(&in.Slug, slug)
			setIf(&in.Body, text)
			setIf(&in.ProjectID, project)
			setFreeTags(&in.Tags, tags)

			if in == (client.KnowledgeInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			// Updates address the record by id, so a slug is resolved first.
			id := args[0]
			if note, err := folio.GetKnowledge(id); err == nil {
				id = note.ID
			}

			note, err := folio.UpdateKnowledge(id, in)
			if err != nil {
				return err
			}
			return renderKnowledge(note)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&slug, "slug", "", "new slug")
	cmd.Flags().StringVar(&body, "body", "", "replace the body, or - for stdin")
	cmd.Flags().StringVar(&project, "project", "", "file it under this project")
	cmd.Flags().StringVar(&tags, "tags", "", "replace the tags; free-form, comma separated")

	return cmd
}

func knowledgeDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <id-or-slug>",
		Short:   "Delete a note",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			id := args[0]
			if note, err := folio.GetKnowledge(id); err == nil {
				id = note.ID
			}

			if err := folio.DeleteKnowledge(id); err != nil {
				return err
			}
			if !flagJSON {
				fmt.Println("deleted " + id)
			}
			return nil
		},
	}
}

// setFreeTags splits a tag list without checking it against a vocabulary.
// Tags elsewhere are rows on a domain, so the CLI can validate them; knowledge
// belongs to no domain and its tags are plain strings on the record.
func setFreeTags(target **[]string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	if len(tags) > 0 {
		*target = &tags
	}
}

func renderKnowledge(note client.Knowledge) error {
	if flagJSON {
		return encode(note)
	}
	return renderKnowledgeList([]client.Knowledge{note})
}

func renderKnowledgeList(notes []client.Knowledge) error {
	if flagJSON {
		return encode(notes)
	}

	if len(notes) == 0 {
		fmt.Println("no knowledge")
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "SLUG\tTITLE\tTAGS")
	for _, note := range notes {
		fmt.Fprintf(out, "%s\t%s\t%s\n", note.Slug, note.Title, strings.Join(note.Tags, ","))
	}
	return out.Flush()
}

func renderKnowledgeDetail(note client.Knowledge) error {
	fmt.Println(note.Title)
	fmt.Println(strings.Repeat("=", len(note.Title)))
	fmt.Printf("slug: %s\n", note.Slug)
	if len(note.Tags) > 0 {
		fmt.Printf("tags: %s\n", strings.Join(note.Tags, ", "))
	}
	if note.Body != "" {
		fmt.Printf("\n%s\n", note.Body)
	}
	return nil
}
