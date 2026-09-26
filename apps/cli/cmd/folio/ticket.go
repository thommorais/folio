package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

func ticketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket",
		Short:   "Create, read, update and delete tickets",
		Aliases: []string{"tickets"},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(
		ticketListCommand(), ticketGetCommand(), ticketBriefCommand(), ticketCreateCommand(),
		ticketUpdateCommand(), ticketResolveCommand(), ticketDeleteCommand(), ticketFrontierCommand(),
	)

	return cmd
}

func ticketListCommand() *cobra.Command {
	var filter client.TicketFilter
	var status, tags string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a project's tickets",
		RunE: func(_ *cobra.Command, _ []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}
			if status != "" {
				filter.Status = strings.Split(status, ",")
			}
			if tags != "" {
				filter.Tags = strings.Split(tags, ",")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			if err := resolveMe(folio, &filter.Assignee); err != nil {
				return err
			}

			tickets, err := folio.ListTickets(project, filter)
			if err != nil {
				return err
			}
			noteIfPaged(len(tickets), filter.Limit)
			return renderTickets(tickets)
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "comma separated: open,in_progress,blocked,done,cancelled")
	cmd.Flags().StringVar(&filter.Priority, "priority", "", "low, medium or high")
	cmd.Flags().StringVar(&filter.Assignee, "assignee", "", assigneeHelp)
	cmd.Flags().StringVar(&tags, "tags", "", "comma separated tags")
	registerTagCompletion(cmd)
	cmd.Flags().StringVarP(&filter.Search, "query", "q", "", "match the title and body")
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, "maximum rows")
	cmd.Flags().IntVar(&filter.Offset, "offset", 0, "rows to skip")

	return cmd
}

func ticketGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-slug>",
		Short: "Show one ticket with its body; a slug needs --project",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			ticket, err := bySlugOrID(args[0], folio.GetTicketBySlug, folio.GetTicket)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(ticket)
			}
			return renderTicketDetail(ticket)
		},
	}
}

func ticketBriefCommand() *cobra.Command {
	var recentJournal int

	cmd := &cobra.Command{
		Use:   "brief <id-or-slug>",
		Short: "Show a ticket with its plans, todos, journal and docs; a slug needs --project",
		Long: `Everything filed under a ticket in one call, for opening a session on it.

Todos come back open first, so the next step is the first row. Logs are the
most recent only.

  folio ticket brief mobile-nav
  folio ticket brief $ID --recent-journal 3 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			brief, err := bySlugOrID(args[0],
				func(project, slug string) (client.TicketBrief, error) {
					return folio.GetTicketBriefBySlug(project, slug, recentJournal)
				},
				func(id string) (client.TicketBrief, error) {
					return folio.GetTicketBrief(id, recentJournal)
				},
			)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(brief)
			}
			return renderBrief(brief)
		},
	}

	cmd.Flags().IntVar(&recentJournal, "recent-journal", 0, "how many journal entries to carry")

	return cmd
}

func renderBrief(b client.TicketBrief) error {
	if err := renderTicketDetail(b.Ticket); err != nil {
		return err
	}

	section := func(title string, rows []string) {
		if len(rows) == 0 {
			return
		}
		fmt.Printf("\n%s\n", title)
		for _, row := range rows {
			fmt.Println("  " + row)
		}
	}

	if b.Map != nil {
		state := fmt.Sprintf("%d open", b.Map.Open)
		if b.Map.Open == 0 {
			state = "the way is clear"
		}
		rows := []string{fmt.Sprintf("%s  %s  %s", b.Map.Ticket.ID, b.Map.Ticket.Title, state)}
		for _, next := range b.Map.Frontier {
			rows = append(rows, fmt.Sprintf("next  %s  %s  %s", next.ID, next.Wayfinder, next.Title))
		}
		section("cycle plan", rows)
	}

	plans := make([]string, 0, len(b.Plans))
	for _, p := range b.Plans {
		plans = append(plans, fmt.Sprintf("%s  %-8s %d/%d  %s", p.ID, p.Status, p.Progress.Done, p.Progress.Total, p.Title))
	}
	section("plans", plans)

	children := make([]string, 0, len(b.Children))
	for _, t := range b.Children {
		row := fmt.Sprintf("%s  %-6s %-12s %-6s %s", t.ID, t.Kind, t.Status, t.Priority, t.Title)
		if t.Resolution != "" {
			row += ": " + t.Resolution
		}
		children = append(children, row)
	}
	section("children", children)

	journal := make([]string, 0, len(b.Journal))
	for _, e := range b.Journal {
		journal = append(journal, fmt.Sprintf("%s  %s  %s", e.ID, e.CreatedAt, e.Title))
	}
	section("journal", journal)

	cycles := make([]string, 0, len(b.Cycles))
	for _, c := range b.Cycles {
		state := c.Phase
		if c.ClosedAt != "" {
			state += " (resolved)"
		}
		cycles = append(cycles, fmt.Sprintf("%s  %d  %-16s %s", c.ID, c.Ordinal, state, c.Resolution))
	}
	section("cycles", cycles)

	docs := make([]string, 0, len(b.Docs))
	for _, d := range b.Docs {
		docs = append(docs, fmt.Sprintf("%s  %-24s %s", d.ID, d.Slug, d.Title))
	}
	section("docs", docs)

	return nil
}

func ticketCreateCommand() *cobra.Command {
	var slug, body, status, priority, assignee, externalRef, tags string
	var size string
	var parent, wayfinder, dependsOn string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a ticket",
		Example: `  folio ticket create "Mobile nav" --body "No nav below md." --tags frontend,bug
  folio ticket create "The map" --wayfinder map
  folio ticket create "Build it" --parent $MAP_ID --depends-on $ID1,$ID2 --size 3`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}

			text, err := bodyFrom(body)
			if err != nil {
				return err
			}
			in := client.TicketInput{Title: &args[0]}
			setIf(&in.Slug, slug)
			setIf(&in.Body, text)
			setIf(&in.Status, status)
			setIf(&in.Priority, priority)
			if err := setSize(&in.Size, size); err != nil {
				return err
			}
			setIf(&in.Assignee, assignee)
			setIf(&in.ExternalRef, externalRef)
			setIf(&in.ParentID, parent)
			setIf(&in.Wayfinder, wayfinder)
			setList(&in.DependsOn, dependsOn)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}
			if err := resolveMe(folio, in.Assignee); err != nil {
				return err
			}

			ticket, err := folio.CreateTicket(project, in)
			if err != nil {
				return err
			}
			return renderTicket(ticket)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "defaults to a slug derived from the title")
	cmd.Flags().StringVar(&body, "body", "", "what the ticket is about; - reads stdin")
	cmd.Flags().StringVar(&status, "status", "", "defaults to open")
	cmd.Flags().StringVar(&priority, "priority", "", "low, medium or high; defaults to medium")
	cmd.Flags().StringVar(&size, "size", "", "effort: 1, 2, 3, 5 or 8")
	cmd.Flags().StringVar(&assignee, "assignee", "", assigneeHelp)
	cmd.Flags().StringVar(&externalRef, "external-ref", "", "key in another tracker, e.g. JIRA-123")
	cmd.Flags().StringVar(&tags, "tags", "", tagHelp())
	cmd.Flags().StringVar(&parent, "parent", "", "id of the ticket this one sits under")
	cmd.Flags().StringVar(&wayfinder, "wayfinder", "", wayfinderHelp)
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "comma separated ticket ids that block this one")
	registerTagCompletion(cmd)

	return cmd
}

func ticketUpdateCommand() *cobra.Command {
	var slug, title, body, status, priority, assignee, externalRef, tags string
	var size string
	var parent, wayfinder, dependsOn string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a ticket, leaving unset fields alone",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(body)
			if err != nil {
				return err
			}
			in := client.TicketInput{}
			setIf(&in.Slug, slug)
			setIf(&in.Title, title)
			setIf(&in.Body, text)
			setIf(&in.Status, status)
			setIf(&in.Priority, priority)
			if err := setSize(&in.Size, size); err != nil {
				return err
			}
			setIf(&in.Assignee, assignee)
			setIf(&in.ExternalRef, externalRef)
			setIf(&in.ParentID, parent)
			setIf(&in.Wayfinder, wayfinder)
			setList(&in.DependsOn, dependsOn)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			if in == (client.TicketInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			if err := resolveMe(folio, in.Assignee); err != nil {
				return err
			}

			ticket, err := folio.UpdateTicket(args[0], in)
			if err != nil {
				return err
			}
			return renderTicket(ticket)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "new slug")
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&body, "body", "", "new body; - reads stdin")
	cmd.Flags().StringVar(&status, "status", "", "open, in_progress, blocked, done or cancelled")
	cmd.Flags().StringVar(&priority, "priority", "", "low, medium or high")
	cmd.Flags().StringVar(&size, "size", "", "effort: 1, 2, 3, 5 or 8")
	cmd.Flags().StringVar(&assignee, "assignee", "", assigneeHelp)
	cmd.Flags().StringVar(&externalRef, "external-ref", "", "key in another tracker")
	cmd.Flags().StringVar(&tags, "tags", "", "replace the tags; "+tagHelp())
	cmd.Flags().StringVar(&parent, "parent", "", "id of the ticket this one sits under")
	cmd.Flags().StringVar(&wayfinder, "wayfinder", "", wayfinderHelp)
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "replace the blockers; comma separated ticket ids")
	registerTagCompletion(cmd)

	return cmd
}

func ticketResolveCommand() *cobra.Command {
	var detail string
	var cancel bool

	cmd := &cobra.Command{
		Use:   "resolve <id-or-slug> <answer>",
		Short: "Close a ticket with its answer; a slug needs --project",
		Long: `Record the answer and close the ticket in one call. A research, prototype,
grilling or task ticket cannot close without one.

The answer is the one line a map lists; --detail holds the reasoning and is
kept as a resolution entry linked from the ticket. --cancel rules the ticket
out of scope instead of marking it done.

  folio ticket resolve tree-or-graph "A graph." --detail - <<'EOF'
  The tree hides blockers, the board hides depth.
  EOF
  folio ticket resolve $ID "Out of scope: no map needs it." --cancel`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			text, err := bodyFrom(detail)
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			ticket, err := bySlugOrID(args[0], folio.GetTicketBySlug, folio.GetTicket)
			if err != nil {
				return err
			}

			status := "done"
			if cancel {
				status = "cancelled"
			}
			resolved, err := folio.ResolveTicket(ticket.ProjectID, ticket.ID, client.Resolution{
				Answer: args[1], Detail: text, Status: status,
			})
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(resolved)
			}
			return renderTicketDetail(resolved)
		},
	}

	cmd.Flags().StringVar(&detail, "detail", "", "the reasoning behind the answer; - reads stdin")
	cmd.Flags().BoolVar(&cancel, "cancel", false, "close as cancelled: ruled out of scope")

	return cmd
}

func ticketFrontierCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "frontier <id>",
		Short: "List a map's takeable children: open, unblocked and unassigned",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			tickets, err := folio.TicketFrontier(args[0])
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(tickets)
			}
			if len(tickets) == 0 {
				fmt.Println("no takeable tickets")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(out, "ID\tWAYFINDER\tTITLE")
			for _, t := range tickets {
				fmt.Fprintf(out, "%s\t%s\t%s\n", t.ID, t.Wayfinder, t.Title)
			}
			return out.Flush()
		},
	}
}

const wayfinderHelp = "map, research, prototype, grilling or task"

func setList(dst **[]string, value string) {
	if value == "" {
		return
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	*dst = &out
}

func ticketDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a ticket, detaching its plans, todos, journal and docs",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			if err := folio.DeleteTicket(args[0]); err != nil {
				return err
			}
			if !flagJSON {
				fmt.Println("deleted " + args[0])
			}
			return nil
		},
	}
}

func renderTicket(ticket client.Ticket) error {
	if flagJSON {
		return encode(ticket)
	}
	return renderTickets([]client.Ticket{ticket})
}

func renderTickets(tickets []client.Ticket) error {
	if flagJSON {
		return encode(tickets)
	}

	if len(tickets) == 0 {
		fmt.Println("no tickets")
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "ID\tSLUG\tSTATUS\tPRIORITY\tPROGRESS\tTITLE\tTAGS")
	for _, ticket := range tickets {
		progress := fmt.Sprintf("%d/%d (%d%%)", ticket.Progress.Done, ticket.Progress.Total, ticket.Progress.Percent)
		status := ticket.Status
		if ticket.Blocked {
			status += " (blocked)"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.ID, ticket.Slug, status, ticket.Priority,
			progress, ticket.Title, strings.Join(ticket.Tags, ","))
	}
	return out.Flush()
}

func renderTicketDetail(ticket client.Ticket) error {
	fmt.Println(ticket.Title)
	fmt.Printf("%s  %s  %s  %d/%d done\n",
		ticket.ID, ticket.Status, ticket.Priority, ticket.Progress.Done, ticket.Progress.Total)
	if ticket.Cycle > 0 {
		fmt.Printf("cycle %d: %s\n", ticket.Cycle, ticket.Phase)
	}

	for label, value := range map[string]string{
		"assignee": ticket.Assignee, "external ref": ticket.ExternalRef,
		"parent": ticket.ParentID, "wayfinder": ticket.Wayfinder,
	} {
		if value != "" {
			fmt.Printf("%s: %s\n", label, value)
		}
	}
	if len(ticket.DependsOn) > 0 {
		blocked := ""
		if ticket.Blocked {
			blocked = " (blocked)"
		}
		fmt.Println("depends on: " + strings.Join(ticket.DependsOn, ", ") + blocked)
	}
	if len(ticket.Tags) > 0 {
		fmt.Println("tags: " + strings.Join(ticket.Tags, ", "))
	}
	if ticket.Resolution != "" {
		fmt.Println("resolution: " + ticket.Resolution)
	}
	if ticket.Body != "" {
		fmt.Println()
		fmt.Println(ticket.Body)
	}
	return nil
}

func setSize(target **int, raw string) error {
	if raw == "" {
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || !validSize(n) {
		return errors.New("size must be one of 1, 2, 3, 5, 8")
	}
	*target = &n
	return nil
}

func validSize(n int) bool {
	for _, allowed := range []int{1, 2, 3, 5, 8} {
		if n == allowed {
			return true
		}
	}
	return false
}
