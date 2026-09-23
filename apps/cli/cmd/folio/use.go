package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"folio/cli/internal/config"
)

func useCommand() *cobra.Command {
	var clear, here bool

	cmd := &cobra.Command{
		Use:   "use [project]",
		Short: "Print the shell export that selects a project",
		Long: "Prints an export statement for " + config.EnvProject + `, so the selection
lives in the shell rather than on disk and two terminals can work on
different projects.

  eval "$(folio use folio)"
  eval "$(folio use --clear)"

--here binds the current directory and everything below it instead, stored
in the config dir so nothing is written to the repository. A shell export
still wins over a binding.

  folio use folio --here
  folio use --clear --here`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if clear && here {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				if err := config.Unbind(wd); err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "unbound "+wd)
				return nil
			}

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

			if here {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				if err := config.Bind(wd, project.Slug); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "bound %s to %s\n", wd, project.Name)
				if env := os.Getenv(config.EnvProject); env != "" && env != project.Slug {
					fmt.Fprintf(os.Stderr, "%s=%s in this shell still takes precedence\n", config.EnvProject, env)
				}
				return nil
			}

			fmt.Printf("export %s=%s\n", config.EnvProject, project.Slug)
			fmt.Fprintln(os.Stderr, "now working on "+project.Name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "unset the selected project")
	cmd.Flags().BoolVar(&here, "here", false, "bind the current directory instead of printing an export")

	return cmd
}
