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
	cmd.AddCommand(cycleListCommand(), cycleOpenCommand(), cyclePhaseCommand(), cycleResolveCommand())

	return cmd
}

func cycleListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "List a ticket's cycles, oldest first",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			cycles, err := folio.ListCycles(args[0])
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
	return &cobra.Command{
		Use:   "open <ticket-id>",
		Short: "Open the next cycle on a ticket, starting at plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			cycle, err := folio.OpenCycle(args[0])
			if err != nil {
				return err
			}
			return renderCycle(cycle)
		},
	}
}

func cyclePhaseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "phase <cycle-id> <plan|do|check|act>",
		Short: "Advance a cycle one phase",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			cycle, err := folio.UpdateCycle(args[0], client.CycleInput{Phase: &args[1]})
			if err != nil {
				return err
			}
			return renderCycle(cycle)
		},
	}
}

func cycleResolveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <cycle-id> <resolution>",
		Short: "Record what happened and close the cycle",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			cycle, err := folio.UpdateCycle(args[0], client.CycleInput{Resolution: &args[1]})
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
	if cycle.Resolution != "" {
		fmt.Println("resolution: " + cycle.Resolution)
	}
	return nil
}
