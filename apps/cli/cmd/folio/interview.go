package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

func interviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "interview",
		Short:   "Run a grilling ticket as an interview answered on a folio page",
		Aliases: []string{"grill"},
		Long: `Run a grilling ticket as an interview. Every command takes the ticket's id,
or its slug with a project selected, and acts on its active interview.

The turn loop: patch a round, hand the user the link, the user answers on the
page, presses Send and says done in the chat; read pending, then patch the next
round with agent.handled set to the last seq read. Nothing here waits.`,
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(interviewStartCommand(), interviewPatchCommand(), interviewPendingCommand(),
		interviewShowCommand(), interviewListCommand(), interviewFinishCommand())

	return cmd
}

func interviewTicket(folio *client.Client, ref string) (client.Ticket, error) {
	return bySlugOrID(ref, folio.GetTicketBySlug, folio.GetTicket)
}

func interviewLink(folio *client.Client, ticket client.Ticket) (string, error) {
	project, err := folio.GetProject(ticket.ProjectID)
	if err != nil {
		return "", err
	}
	if project.DomainID == "" {
		return "", fmt.Errorf("no page link: project %s has no domain", project.Slug)
	}
	domains, err := folio.ListDomains()
	if err != nil {
		return "", err
	}
	for _, d := range domains {
		if d.ID == project.DomainID {
			parts := []string{d.ClientSlug, d.Slug, project.Slug, "tickets", ticket.Slug, "interview"}
			for i, p := range parts {
				parts[i] = url.PathEscape(p)
			}
			return folio.BaseURL() + "/" + strings.Join(parts, "/"), nil
		}
	}
	return "", fmt.Errorf("no page link: domain %s of project %s is not visible to you", project.DomainID, project.Slug)
}

func interviewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <ticket>",
		Short: "Start the ticket's interview, or resume the active one, and print its link",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := interviewTicket(folio, args[0])
			if err != nil {
				return err
			}
			link, err := interviewLink(folio, ticket)
			if err != nil {
				return err
			}
			interview, err := folio.StartInterview(ticket.ID)
			if err != nil {
				return err
			}
			return renderInterview(interview, link)
		},
	}
}

func interviewShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <ticket>",
		Short: "Print the active interview: round, open questions, handled, agent status, link",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := interviewTicket(folio, args[0])
			if err != nil {
				return err
			}
			interview, err := folio.CurrentInterview(ticket.ID)
			if err != nil {
				return err
			}
			link, err := interviewLink(folio, ticket)
			if err != nil {
				return err
			}
			return renderInterview(interview, link)
		},
	}
}

func renderInterview(interview client.Interview, link string) error {
	if flagJSON {
		return encode(map[string]any{"link": link, "interview": interview})
	}
	questions, err := interview.Questions()
	if err != nil {
		return err
	}
	round := 0
	var open []client.InterviewQuestion
	for _, q := range questions {
		round = max(round, q.Round)
		if q.Status == "open" || q.Status == "reopened" {
			open = append(open, q)
		}
	}

	fmt.Println(link)
	fmt.Println(interview.Topic)
	fmt.Printf("round %d  handled %d  agent %s\n", round, interview.Handled, interview.AgentStatus)
	if len(open) == 0 {
		fmt.Println("no open questions")
		return nil
	}
	out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, q := range open {
		fmt.Fprintf(out, "%s\t%s\t%s\n", q.ID, q.Status, q.Title)
	}
	return out.Flush()
}

func interviewPatchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "patch <ticket>",
		Short: "Apply a JSON patch from stdin: the next round, answers and agent.handled in one write",
		Long: `Apply a JSON patch read from stdin to the active interview. Questions merge
by id, so one write can add a round, record answers and replies, and move
agent.handled past the Sends it answers.

  folio interview patch tree-or-graph <<'EOF'
  {"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}}],"agent":{"handled":3}}
  EOF`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			raw = bytes.TrimSpace(raw)
			if len(raw) == 0 {
				return errors.New("the patch is empty: pipe the JSON on stdin")
			}
			if !json.Valid(raw) {
				return errors.New("the patch on stdin is not valid JSON")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			summary, err := folio.PatchInterview(ticket, raw)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(summary)
			}
			fmt.Printf("round %d: %d questions added, %d answered, handled %d\n", summary.Round, summary.Added, summary.Answered, summary.Handled)
			return nil
		},
	}
}

func interviewPendingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pending <ticket>",
		Short: "Print the Sends past handled, one JSON line each, without waiting",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			pending, err := folio.PendingSends(ticket)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(pending)
			}
			if len(pending.Sends) == 0 {
				fmt.Printf("nothing sent since handled %d\n", pending.Handled)
				return nil
			}
			for _, send := range pending.Sends {
				line, err := json.Marshal(send)
				if err != nil {
					return err
				}
				fmt.Println(string(line))
			}
			return nil
		},
	}
}

func interviewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <ticket>",
		Short: "List the ticket's interviews, active and earlier",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			interviews, err := folio.ListInterviews(ticket)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(interviews)
			}
			if len(interviews) == 0 {
				fmt.Println("no interviews")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(out, "ID\tSTATE\tQUESTIONS\tHANDLED\tSTARTED")
			for _, i := range interviews {
				questions, err := i.Questions()
				if err != nil {
					return err
				}
				state := "active"
				if i.FinishedAt != "" {
					state = "finished " + i.FinishedAt
				}
				fmt.Fprintf(out, "%s\t%s\t%d\t%d\t%s\n", i.ID, state, len(questions), i.Handled, i.CreatedAt)
			}
			return out.Flush()
		},
	}
}

func interviewFinishCommand() *cobra.Command {
	var doc string

	cmd := &cobra.Command{
		Use:   "finish <ticket> <answer>",
		Short: "Resolve and close the ticket with the design doc, and lock the interview",
		Long: `Finish the interview: the doc becomes the ticket's resolution entry, the
answer its one-line resolution, and the ticket closes. Every question must be
answered or deferred first.

  folio interview finish tree-or-graph "A graph." --doc - <<'EOF'
  # Tree or graph
  ...
  EOF`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if doc == "" {
				return errors.New("--doc is required: the design doc, - reads stdin")
			}
			text, err := bodyFrom(doc)
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			interview, err := folio.FinishInterview(ticket, args[1], text)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(interview)
			}
			fmt.Printf("interview %s finished; ticket %s resolved: %s\n", interview.ID, ticket, args[1])
			return nil
		},
	}

	cmd.Flags().StringVar(&doc, "doc", "", "the design doc, kept as the resolution entry; - reads stdin")

	return cmd
}
