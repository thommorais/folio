package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
	"folio/cli/internal/git"
)

var errUncommitted = errors.New("uncommitted changes")

func stopCommand() *cobra.Command {
	var allowDirty bool

	cmd := &cobra.Command{
		Use:   "stop <ticket> <note|->",
		Short: "End a session on a ticket with a handoff note for folio resume",
		Long: `Records a handoff on the ticket: the note plus the branch, HEAD commit and PR.
Refuses while the tree has uncommitted changes; commit first, or pass
--allow-dirty to record the paths instead.

Write the note for the next session: state, next step, traps. Terse.`,
		Example: `  folio stop $ID - <<'EOF'
  State: coverage widget renders, no entitlement lock yet.
  Next: hasFeature gate on product rows.
  Trap: subscribedCountries is empty in dev seed.
  EOF`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			note, err := bodyFrom(args[1])
			if err != nil {
				return err
			}
			if strings.TrimSpace(note) == "" {
				return errors.New("nothing to hand off: pass a note or -")
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			meta := map[string]any{}
			if commit := git.Commit(wd); commit != "" {
				meta["commit"] = commit
			}
			if changes := git.Changes(wd); len(changes) > 0 {
				if !allowDirty {
					return fmt.Errorf("%w: %s; commit, then rerun, or pass --allow-dirty", errUncommitted, strings.Join(changes, ", "))
				}
				meta["uncommitted"] = changes
			}

			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := bySlugOrID(args[0], folio.GetTicketBySlug, folio.GetTicket)
			if err != nil {
				return err
			}

			entry, err := folio.WriteHandoff(ticket.ProjectID, ticket.ID, client.Handoff{
				Body:   note,
				Branch: git.Branch(wd),
				PR:     git.PR(wd),
				Meta:   meta,
			})
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(entry)
			}
			fmt.Printf("%s\nresume: folio resume %s\n", entry.ID, ticket.ID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&allowDirty, "allow-dirty", false, "hand off with uncommitted changes, recording their paths")

	return cmd
}
