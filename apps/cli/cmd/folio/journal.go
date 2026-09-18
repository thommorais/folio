package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
	"folio/cli/internal/git"
)

func bodyFrom(value string) (string, error) {
	if value != "-" {
		return value, nil
	}
	piped, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(piped), nil
}

func journalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "journal",
		Short:   "Write and read journal entries",
		Aliases: []string{"log", "logs"},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(
		journalListCommand(),
		journalGetCommand(),
		journalWriteCommand(),
		journalUpdateCommand(),
		journalAppendCommand(),
		journalDeleteCommand(),
	)

	return cmd
}

func journalListCommand() *cobra.Command {
	var filter client.JournalFilter
	var tags string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a project's journal entries, newest first",
		RunE: func(_ *cobra.Command, _ []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}
			if tags != "" {
				filter.Tags = strings.Split(tags, ",")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			entries, err := folio.ListJournal(project, filter)
			if err != nil {
				return err
			}
			return renderLogs(entries)
		},
	}

	cmd.Flags().StringVar(&filter.Branch, "branch", "", "entries anchored to this branch")
	cmd.Flags().StringVar(&filter.TicketID, "ticket", "", "entries filed under this ticket")
	cmd.Flags().StringVar(&filter.ExternalRef, "external-ref", "", "entries carrying this external tracker key")
	cmd.Flags().StringVar(&filter.PlanID, "plan", "", "entries under this plan")
	cmd.Flags().StringVar(&filter.TicketID, "todo", "", "entries under this todo")
	cmd.Flags().StringVar(&tags, "tags", "", "comma separated tags")
	registerTagCompletion(cmd)
	cmd.Flags().StringVarP(&filter.Search, "query", "q", "", "match the title and body")
	cmd.Flags().StringVar(&filter.Since, "since", "", "RFC 3339 lower bound")
	cmd.Flags().StringVar(&filter.Until, "until", "", "RFC 3339 upper bound")
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, "maximum rows")
	cmd.Flags().IntVar(&filter.Offset, "offset", 0, "rows to skip")

	return cmd
}

func journalGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-slug>",
		Short: "Show one entry with its body; a slug needs --project",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			// A slug is unique only within a project, so a slug needs the
			// project route; ids resolve without one.
			get := folio.GetJournalEntry
			if project := config.Project(flagProject); project != "" {
				get = func(ref string) (client.JournalEntry, error) {
					return folio.GetJournalEntryBySlug(project, ref)
				}
			}

			entry, err := get(args[0])
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(entry)
			}
			return renderLogDetail(entry)
		},
	}
}

func journalWriteCommand() *cobra.Command {
	var slug, body, branch, pr, ticket, externalRef, plan, todo, tags string

	cmd := &cobra.Command{
		Use:   "write <title>",
		Short: "Write a journal entry, with --body - to read markdown from stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}

			text, err := bodyFrom(body)
			if err != nil {
				return err
			}

			// Changed rather than an empty check, so --branch "" records no
			// branch instead of being refilled by detection.
			wd, _ := os.Getwd()
			if !cmd.Flags().Changed("branch") {
				branch = git.Branch(wd)
			}
			if !cmd.Flags().Changed("pr") {
				pr = git.PR(wd)
			}

			in := client.LogInput{Title: &args[0]}
			setIf(&in.Slug, slug)
			setIf(&in.Body, text)
			setIf(&in.Branch, branch)
			setIf(&in.PR, pr)
			setIf(&in.TicketID, ticket)
			setIf(&in.ExternalRef, externalRef)
			setIf(&in.PlanID, plan)
			setIf(&in.TicketID, todo)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			entry, err := folio.WriteJournalEntry(project, in)
			if err != nil {
				return err
			}
			return renderLog(entry)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "derived from the title when omitted")
	cmd.Flags().StringVar(&body, "body", "", "markdown body, or - for stdin")
	cmd.Flags().StringVar(&branch, "branch", "", "git branch")
	cmd.Flags().StringVar(&pr, "pr", "", "pull request url")
	cmd.Flags().StringVar(&ticket, "ticket", "", "ticket this belongs to")
	cmd.Flags().StringVar(&externalRef, "external-ref", "", "key in another tracker, e.g. JIRA-123")
	cmd.Flags().StringVar(&plan, "plan", "", "plan this documents")
	cmd.Flags().StringVar(&todo, "todo", "", "todo this documents")
	cmd.Flags().StringVar(&tags, "tags", "", tagHelp())
	registerTagCompletion(cmd)

	return cmd
}

func journalUpdateCommand() *cobra.Command {
	var title, body, branch, pr, ticket, externalRef, tags string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an entry, leaving unset fields alone",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(body)
			if err != nil {
				return err
			}

			in := client.LogInput{}
			setIf(&in.Title, title)
			setIf(&in.Body, text)
			setIf(&in.Branch, branch)
			setIf(&in.PR, pr)
			setIf(&in.TicketID, ticket)
			setIf(&in.ExternalRef, externalRef)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			if in == (client.LogInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			entry, err := folio.UpdateJournalEntry(args[0], in)
			if err != nil {
				return err
			}
			return renderLog(entry)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&body, "body", "", "replace the body, or - for stdin")
	cmd.Flags().StringVar(&branch, "branch", "", "git branch")
	cmd.Flags().StringVar(&pr, "pr", "", "pull request url")
	cmd.Flags().StringVar(&ticket, "ticket", "", "ticket this belongs to")
	cmd.Flags().StringVar(&externalRef, "external-ref", "", "key in another tracker")
	cmd.Flags().StringVar(&tags, "tags", "", "replace the tags; "+tagHelp())
	registerTagCompletion(cmd)

	return cmd
}

func journalAppendCommand() *cobra.Command {
	var section string

	cmd := &cobra.Command{
		Use:   "append <id>",
		Short: "Append a section to an entry's body",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(section)
			if err != nil {
				return err
			}
			if strings.TrimSpace(text) == "" {
				return errors.New("--section is required (use - to read stdin)")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			entry, err := folio.AppendJournalEntry(args[0], text)
			if err != nil {
				return err
			}
			return renderLog(entry)
		},
	}

	cmd.Flags().StringVar(&section, "section", "", "markdown to append, or - for stdin")

	return cmd
}

func journalDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a journal entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			if err := folio.DeleteJournalEntry(args[0]); err != nil {
				return err
			}
			if !flagJSON {
				fmt.Println("deleted " + args[0])
			}
			return nil
		},
	}
}

func renderLog(entry client.JournalEntry) error {
	if flagJSON {
		return encode(entry)
	}
	return renderLogs([]client.JournalEntry{entry})
}

func renderLogs(entries []client.JournalEntry) error {
	if flagJSON {
		return encode(entries)
	}

	if len(entries) == 0 {
		fmt.Println("no journal entries")
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "ID\tDATE\tBRANCH\tTITLE\tTAGS")
	for _, entry := range entries {
		date := entry.CreatedAt
		if len(date) >= 10 {
			date = date[:10]
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n", entry.ID, date, entry.Branch, entry.Title, strings.Join(entry.Tags, ","))
	}
	return out.Flush()
}

func renderLogDetail(entry client.JournalEntry) error {
	fmt.Println(entry.Title)
	fmt.Println(strings.Repeat("=", len(entry.Title)))
	fmt.Printf("slug: %s\n", entry.Slug)

	for label, value := range map[string]string{"branch": entry.Branch, "pr": entry.PR, "external ref": entry.ExternalRef} {
		if value != "" {
			fmt.Printf("%s: %s\n", label, value)
		}
	}
	if len(entry.Tags) > 0 {
		fmt.Printf("tags: %s\n", strings.Join(entry.Tags, ", "))
	}

	if entry.Body != "" {
		fmt.Printf("\n%s\n", entry.Body)
	}
	return nil
}
