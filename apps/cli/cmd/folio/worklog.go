package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

func workLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worklog",
		Short:   "Record what happened while working a ticket, plan or todo",
		Aliases: []string{"worklogs", "wl"},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(workLogListCommand(), workLogWriteCommand(), workLogDeleteCommand())

	return cmd
}

type workLogTarget struct {
	ticket string
	plan   string
	todo   string
}

func (t workLogTarget) resolve() (kind, id string, err error) {
	switch {
	case t.ticket != "" && t.plan == "" && t.todo == "":
		return "ticket", t.ticket, nil
	case t.plan != "" && t.ticket == "" && t.todo == "":
		return "plan", t.plan, nil
	case t.todo != "" && t.ticket == "" && t.plan == "":
		return "todo", t.todo, nil
	case t.ticket == "" && t.plan == "" && t.todo == "":
		return "", "", errors.New("pass exactly one of --ticket, --plan or --todo")
	default:
		return "", "", errors.New("pass only one of --ticket, --plan or --todo")
	}
}

func (t *workLogTarget) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&t.ticket, "ticket", "", "ticket id")
	cmd.Flags().StringVar(&t.plan, "plan", "", "plan id")
	cmd.Flags().StringVar(&t.todo, "todo", "", "todo id")
}

func workLogListCommand() *cobra.Command {
	var target workLogTarget
	var filter client.WorkLogFilter

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the work log of one ticket, plan or todo",
		RunE: func(_ *cobra.Command, _ []string) error {
			kind, id, err := target.resolve()
			if err != nil {
				return err
			}
			folio, err := api()
			if err != nil {
				return err
			}

			project, err := resolveProject()
			if err != nil {
				return err
			}

			var entries []client.WorkLog
			switch kind {
			case "plan":
				entries, err = folio.ListPlanLogs(project, id, filter)
			default:
				entries, err = folio.ListIssueLogs(project, id, filter)
			}
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(entries)
			}
			if len(entries) == 0 {
				fmt.Println("no work log entries")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(out, "ID\tCYCLE\tWHEN\tBODY")
			for _, e := range entries {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", e.ID, e.CycleID, e.CreatedAt, firstLine(e.Body))
			}
			return out.Flush()
		},
	}

	target.register(cmd)
	cmd.Flags().StringVar(&filter.Cycle, "cycle", "", "only entries from this cycle; tickets only")
	cmd.Flags().StringVarP(&filter.Search, "query", "q", "", "match the body")
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, "maximum rows")

	return cmd
}

func workLogWriteCommand() *cobra.Command {
	var target workLogTarget
	var body string

	cmd := &cobra.Command{
		Use:   "write <body>",
		Short: "Add an entry, with - to read from stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			kind, id, err := target.resolve()
			if err != nil {
				return err
			}
			text := body
			if len(args) == 1 {
				text = args[0]
			}
			text, err = bodyFrom(text)
			if err != nil {
				return err
			}
			if text == "" {
				return errors.New("nothing to write: pass a body or -")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			project, err := resolveProject()
			if err != nil {
				return err
			}

			var entry client.WorkLog
			switch kind {
			case "plan":
				entry, err = folio.WritePlanLog(project, id, text)
			default:
				entry, err = folio.WriteIssueLog(project, id, text)
			}
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(entry)
			}
			fmt.Println(entry.ID)
			if entry.CycleID != "" {
				fmt.Println("cycle: " + entry.CycleID)
			}
			return nil
		},
	}

	target.register(cmd)
	cmd.Flags().StringVar(&body, "body", "", "entry body, or - to read stdin")

	return cmd
}

func workLogDeleteCommand() *cobra.Command {
	var target workLogTarget

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a work log entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			return folio.DeleteWorkLog(args[0])
		},
	}

	target.register(cmd)

	return cmd
}

func firstLine(body string) string {
	for i, r := range body {
		if r == '\n' {
			return body[:i] + "…"
		}
	}
	return body
}
