package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
)

func cycleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cycle",
		Short:   "Open, advance and resolve a ticket's PDCA cycles",
		Aliases: []string{"cycles"},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(cycleListCommand(), cycleOpenCommand(), cycleNextCommand(), cyclePhaseCommand(), cycleResolveCommand())

	return cmd
}

func cycleListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <ticket>",
		Short: "List a ticket's cycles, oldest first",
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
			cycles, err := folio.ListCycles(ticket)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(cycles)
			}
			if len(cycles) == 0 {
				fmt.Println("no cycles")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(out, "ID\tCYCLE\tPHASE\tRESOLUTION")
			for _, c := range cycles {
				state := c.Phase
				if c.ClosedAt != "" {
					state += " (resolved)"
				}
				fmt.Fprintf(out, "%s\t%d\t%s\t%s\n", c.ID, c.Ordinal, state, c.Resolution)
			}
			return out.Flush()
		},
	}
}

func cycleOpenCommand() *cobra.Command {
	var mapTitle string

	cmd := &cobra.Command{
		Use:   "open <ticket>",
		Short: "Open the next cycle on a ticket, starting at plan",
		Long: `Open the next cycle on a ticket. With --map, also create a wayfinder map
under the ticket and link it as this cycle's plan: the cycle cannot leave
plan while a decision on that map is open.

  folio cycle open wayfinder-view --map "Plan the wayfinder view"`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			if mapTitle == "" {
				ticket, err := ticketID(folio, args[0])
				if err != nil {
					return err
				}
				cycle, err := folio.OpenCycle(ticket)
				if err != nil {
					return err
				}
				return renderCycle(cycle)
			}

			ticket, err := bySlugOrID(args[0], folio.GetTicketBySlug, folio.GetTicket)
			if err != nil {
				return err
			}
			cycle, theMap, err := folio.OpenCycleWithMap(ticket.ProjectID, ticket.ID, mapTitle)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(map[string]any{"cycle": cycle, "map": theMap})
			}
			if err := renderCycle(cycle); err != nil {
				return err
			}
			fmt.Printf("map: %s  %s\n", theMap.ID, theMap.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&mapTitle, "map", "", "title of a wayfinder map to create under the ticket and plan this cycle")

	return cmd
}

func cycleNextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "next <ticket>",
		Short: "Advance the ticket's current cycle one phase: plan, do, check, act",
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
			cycle, err := folio.NextPhase(ticket)
			if err != nil {
				return err
			}
			return renderCycle(cycle)
		},
	}
}

func cyclePhaseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "phase <ticket> <plan|do|check|act>",
		Short: "Move the ticket's current cycle to the next phase by name",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			cycle, err := folio.UpdateCurrentCycle(ticket, client.CycleInput{Phase: &args[1]})
			if err != nil {
				return err
			}
			return renderCycle(cycle)
		},
	}
}

func cycleResolveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <ticket> <resolution>",
		Short: "Record what happened and close the ticket's current cycle",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			ticket, err := ticketID(folio, args[0])
			if err != nil {
				return err
			}
			cycle, err := folio.UpdateCurrentCycle(ticket, client.CycleInput{Resolution: &args[1]})
			if err != nil {
				return err
			}
			return renderCycle(cycle)
		},
	}
}

func renderCycle(cycle client.Cycle) error {
	if flagJSON {
		return encode(cycle)
	}
	fmt.Printf("cycle %d  %s\n", cycle.Ordinal, cycle.Phase)
	fmt.Println(cycle.ID)
	if cycle.MapID != "" {
		fmt.Println("map: " + cycle.MapID)
	}
	if cycle.Resolution != "" {
		fmt.Println("resolution: " + cycle.Resolution)
	}
	return nil
}
