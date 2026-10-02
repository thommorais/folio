package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
	"folio/cli/internal/git"
)

type checkout struct {
	branch string
	commit string
}

func resumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <ticket>",
		Short: "Pick a ticket back up: last handoff, logs since, open work, body",
		Long: `Everything the next session needs in one call: the latest handoff from
folio stop, work logs written after it, open children, live plans, the cycle
map's next steps, doc pointers and the ticket body. Warns when the checkout
differs from the one handed off.

End the session with folio stop.`,
		Example: `  folio resume $ID
  folio resume landing-page -p geral --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			resume, err := bySlugOrID(args[0], folio.GetTicketResumeBySlug, folio.GetTicketResume)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(resume)
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			return renderResume(resume, checkout{git.Branch(wd), git.Commit(wd)})
		},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")

	return cmd
}

func renderResume(r client.TicketResume, here checkout) error {
	body := r.Ticket.Body
	header := r.Ticket
	header.Body = ""
	if err := renderTicketDetail(header); err != nil {
		return err
	}

	logsTitle := "recent logs"
	if r.Handoff == nil {
		fmt.Println("\nno handoff yet")
	} else {
		logsTitle = "logs since handoff"
		fmt.Printf("\nhandoff %s  %s\n", r.Handoff.CreatedAt, at(r.Handoff.Branch, handoffCommit(*r.Handoff)))
		for _, line := range strings.Split(strings.TrimRight(r.Handoff.Body, "\n"), "\n") {
			fmt.Println("  " + line)
		}
		if paths := uncommitted(*r.Handoff); len(paths) > 0 {
			fmt.Println("uncommitted at handoff: " + strings.Join(paths, ", "))
		}
		if d := drift(*r.Handoff, here); d != "" {
			fmt.Println(d)
		}
	}

	logs := make([]string, 0, len(r.Logs))
	for _, l := range r.Logs {
		lines := strings.Split(strings.TrimRight(l.Body, "\n"), "\n")
		logs = append(logs, fmt.Sprintf("%s  %s  %s", l.ID, l.CreatedAt, lines[0]))
		for _, more := range lines[1:] {
			logs = append(logs, "  "+more)
		}
	}
	section(logsTitle, logs)

	section("cycle plan", mapRows(r.Map))

	open := make([]string, 0, len(r.Open))
	for _, t := range r.Open {
		row := fmt.Sprintf("%s  %-6s %-12s %-6s %s", t.ID, t.Kind, t.Status, t.Priority, t.Title)
		if t.Blocked {
			row += " (blocked)"
		}
		open = append(open, row)
	}
	section("open", open)
	if r.Closed > 0 {
		fmt.Printf("closed: %d\n", r.Closed)
	}

	plans := make([]string, 0, len(r.Plans))
	for _, p := range r.Plans {
		plans = append(plans, fmt.Sprintf("%s  %-8s %d/%d  %s", p.ID, p.Status, p.Progress.Done, p.Progress.Total, p.Title))
	}
	section("plans", plans)

	docs := make([]string, 0, len(r.Docs))
	for _, d := range r.Docs {
		docs = append(docs, fmt.Sprintf("%s  %s  %s", d.ID, d.Slug, d.Title))
	}
	section("docs", docs)

	if body != "" {
		fmt.Println()
		fmt.Println(strings.TrimRight(body, "\n"))
	}
	fmt.Printf("\nstop: folio stop %s - (terse: state, next, traps)\n", r.Ticket.ID)
	return nil
}

func drift(handoff client.JournalEntry, here checkout) string {
	if here.branch == "" && here.commit == "" {
		return ""
	}
	commit := handoffCommit(handoff)
	if here.branch == handoff.Branch && here.commit == commit {
		return ""
	}
	return fmt.Sprintf("drift: here %s, handoff %s", at(here.branch, here.commit), at(handoff.Branch, commit))
}

func at(branch, commit string) string {
	if len(commit) > 7 {
		commit = commit[:7]
	}
	switch {
	case branch == "":
		return commit
	case commit == "":
		return branch
	}
	return branch + " @ " + commit
}

func handoffCommit(e client.JournalEntry) string {
	commit, _ := e.Meta["commit"].(string)
	return commit
}

func uncommitted(e client.JournalEntry) []string {
	raw, _ := e.Meta["uncommitted"].([]any)
	paths := make([]string, 0, len(raw))
	for _, p := range raw {
		if s, ok := p.(string); ok {
			paths = append(paths, s)
		}
	}
	return paths
}
