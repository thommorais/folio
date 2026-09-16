package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"folio/cli/internal/config"
)

func useCommand() *cobra.Command {
	var clear bool

	cmd := &cobra.Command{
		Use:   "use [project]",
		Short: "Print the shell export that selects a project",
		Long: "Prints an export statement for " + config.EnvProject + `, so the selection
lives in the shell rather than on disk and two terminals can work on
different projects.

  eval "$(folio use folio)"
  eval "$(folio use --clear)"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if clear {
				fmt.Printf("unset %s\n", config.EnvProject)
				fmt.Fprintln(os.Stderr, "project cleared")
				return nil
			}

			if len(args) == 0 {
				if current := config.Project(""); current != "" {
					fmt.Fprintln(os.Stderr, "current project: "+current)
					return nil
				}
				fmt.Fprintln(os.Stderr, "no project selected")
				return nil
			}

			folio, err := api()
			if err != nil {
				return err
			}

			// Resolve it now so a typo fails here rather than on the next
			// command, and so the export holds the canonical slug.
			project, err := folio.GetProject(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("export %s=%s\n", config.EnvProject, project.Slug)
			fmt.Fprintln(os.Stderr, "now working on "+project.Name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "unset the selected project")

	return cmd
}
