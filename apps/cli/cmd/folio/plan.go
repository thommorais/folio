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

func planCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Create, read, update and delete plans",
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(planListCommand(), planGetCommand(), planCreateCommand(), planUpdateCommand(), planDeleteCommand())

	return cmd
}

func planListCommand() *cobra.Command {
	var filter client.PlanFilter
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a project's plans",
		RunE: func(_ *cobra.Command, _ []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}
			if status != "" {
				filter.Status = strings.Split(status, ",")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			plans, err := folio.ListPlans(project, filter)
			if err != nil {
				return err
			}
			return renderPlans(plans)
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "comma separated: draft,active,done,abandoned")
	cmd.Flags().StringVar(&filter.TicketID, "ticket", "", "only plans under this ticket")
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, "maximum rows")
	cmd.Flags().IntVar(&filter.Offset, "offset", 0, "rows to skip")

	return cmd
}

func planGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			plan, err := folio.GetPlan(args[0])
			if err != nil {
				return err
			}
			return renderPlan(plan)
		},
	}
}

func planCreateCommand() *cobra.Command {
	var goal, status, ticket, tags string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			project, err := resolveProject()
			if err != nil {
				return err
			}

			in := client.PlanInput{Title: &args[0]}
			setIf(&in.Goal, goal)
			setIf(&in.Status, status)
			setIf(&in.TicketID, ticket)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			plan, err := folio.CreatePlan(project, in)
			if err != nil {
				return err
			}
			return renderPlan(plan)
		},
	}

	cmd.Flags().StringVar(&goal, "goal", "", "what the plan is for")
	cmd.Flags().StringVar(&status, "status", "", "defaults to draft")
	cmd.Flags().StringVar(&ticket, "ticket", "", "ticket id to file it under")
	cmd.Flags().StringVar(&tags, "tags", "", tagHelp())
	registerTagCompletion(cmd)

	return cmd
}

func planUpdateCommand() *cobra.Command {
	var title, goal, status, ticket, tags string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a plan, leaving unset fields alone",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			in := client.PlanInput{}
			setIf(&in.Title, title)
			setIf(&in.Goal, goal)
			setIf(&in.Status, status)
			setIf(&in.TicketID, ticket)
			if err := setTags(&in.Tags, tags); err != nil {
				return err
			}

			if in == (client.PlanInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			plan, err := folio.UpdatePlan(args[0], in)
			if err != nil {
				return err
			}
			return renderPlan(plan)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&goal, "goal", "", "new goal")
	cmd.Flags().StringVar(&status, "status", "", "draft, active, done or abandoned")
	cmd.Flags().StringVar(&ticket, "ticket", "", "move under this ticket")
	cmd.Flags().StringVar(&tags, "tags", "", "replace the tags; "+tagHelp())
	registerTagCompletion(cmd)

	return cmd
}

func planDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a plan, detaching its todos",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			if err := folio.DeletePlan(args[0]); err != nil {
				return err
			}
			if !flagJSON {
				fmt.Println("deleted " + args[0])
			}
			return nil
		},
	}
}

func renderPlan(plan client.Plan) error {
	if flagJSON {
		return encode(plan)
	}
	return renderPlans([]client.Plan{plan})
}

func renderPlans(plans []client.Plan) error {
	if flagJSON {
		return encode(plans)
	}

	if len(plans) == 0 {
		fmt.Println("no plans")
		return nil
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "ID\tSTATUS\tPROGRESS\tTITLE\tTAGS")
	for _, plan := range plans {
		progress := fmt.Sprintf("%d/%d (%d%%)", plan.Progress.Done, plan.Progress.Total, plan.Progress.Percent)
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n", plan.ID, plan.Status, progress, plan.Title, strings.Join(plan.Tags, ","))
	}
	return out.Flush()
}
