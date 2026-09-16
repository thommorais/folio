package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
)

func projectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "List and inspect projects",
		Aliases: []string{"projects"},
	}

	cmd.AddCommand(
		projectListCommand(),
		projectGetCommand(),
		projectCreateCommand(),
		projectUpdateCommand(),
		projectDeleteCommand(),
		projectMemberCommand(),
	)

	return cmd
}

func projectCreateCommand() *cobra.Command {
	var slug, descr string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a project, becoming its first owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.CreateProject(client.CreateProjectInput{
				Name:  args[0],
				Slug:  slug,
				Descr: descr,
			})
			if err != nil {
				return err
			}

			if flagJSON {
				return encode(project)
			}
			if err := renderProjectDetail(project); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "\nselect it with: eval \"$(folio use %s)\"\n", project.Slug)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "derived from the name when omitted")
	cmd.Flags().StringVar(&descr, "descr", "", "what the project is")

	return cmd
}

func projectUpdateCommand() *cobra.Command {
	var name, descr string
	var archive, unarchive bool

	cmd := &cobra.Command{
		Use:   "update [id-or-slug]",
		Short: "Update a project, defaulting to the selected one",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := projectRef(args)
			if err != nil {
				return err
			}

			if archive && unarchive {
				return errors.New("--archive and --unarchive are mutually exclusive")
			}

			in := client.ProjectInput{}
			setIf(&in.Name, name)
			setIf(&in.Descr, descr)
			if archive || unarchive {
				in.Archived = &archive
			}

			if in == (client.ProjectInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.UpdateProject(ref, in)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(project)
			}
			return renderProjectDetail(project)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringVar(&descr, "descr", "", "new description")
	cmd.Flags().BoolVar(&archive, "archive", false, "archive the project")
	cmd.Flags().BoolVar(&unarchive, "unarchive", false, "restore an archived project")

	return cmd
}

func projectDeleteCommand() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <id-or-slug>",
		Short: "Delete a project and everything under it",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Deleting a project cascades to its plans, todos, logs and docs,
			// so it takes an explicit flag rather than a prompt an agent
			// cannot answer.
			if !confirm {
				return errors.New("refusing to delete a project and its contents without --yes")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			if err := folio.DeleteProject(args[0]); err != nil {
				return err
			}
			if !flagJSON {
				fmt.Println("deleted " + args[0])
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "yes", false, "confirm the cascading delete")

	return cmd
}

func projectMemberCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member",
		Short:   "Add, promote and remove project members",
		Aliases: []string{"members"},
	}

	var role string

	add := &cobra.Command{
		Use:   "add <email>",
		Short: "Add a member by email",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := projectRef(nil)
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.AddMember(ref, args[0], role)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(project)
			}
			return renderProjectDetail(project)
		},
	}
	add.Flags().StringVar(&role, "role", "", "owner, editor or viewer; defaults to editor")

	var newRole string
	promote := &cobra.Command{
		Use:   "set-role <user-id>",
		Short: "Change a member's role",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := projectRef(nil)
			if err != nil {
				return err
			}
			if newRole == "" {
				return errors.New("--role is required")
			}

			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.SetMemberRole(ref, args[0], newRole)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(project)
			}
			return renderProjectDetail(project)
		},
	}
	promote.Flags().StringVar(&newRole, "role", "", "owner, editor or viewer")

	remove := &cobra.Command{
		Use:   "remove <user-id>",
		Short: "Remove a member",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := projectRef(nil)
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.RemoveMember(ref, args[0])
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(project)
			}
			return renderProjectDetail(project)
		},
	}

	cmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")
	cmd.AddCommand(add, promote, remove)

	return cmd
}

// projectRef takes a positional argument when given, else the selection.
func projectRef(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	return resolveProject()
}

func projectListCommand() *cobra.Command {
	var archived bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the projects you are a member of",
		RunE: func(_ *cobra.Command, _ []string) error {
			folio, err := api()
			if err != nil {
				return err
			}

			projects, err := folio.ListProjects(archived)
			if err != nil {
				return err
			}
			return renderProjects(projects)
		},
	}

	cmd.Flags().BoolVar(&archived, "archived", false, "include archived projects")

	return cmd
}

func projectGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id-or-slug]",
		Short: "Show one project, defaulting to the selected one",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := projectRef(args)
			if err != nil {
				return err
			}

			folio, err := api()
			if err != nil {
				return err
			}

			project, err := folio.GetProject(ref)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(project)
			}
			return renderProjectDetail(project)
		},
	}
}

func renderProjects(projects []client.Project) error {
	if flagJSON {
		return encode(projects)
	}

	if len(projects) == 0 {
		fmt.Println("no projects")
		return nil
	}

	selected := config.Project(flagProject)

	out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(out, "\tSLUG\tNAME\tMEMBERS")
	for _, project := range projects {
		marker := " "
		if project.Slug == selected || project.ID == selected {
			marker = "*"
		}
		name := project.Name
		if project.Archived {
			name += " (archived)"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%d\n", marker, project.Slug, name, len(project.Members))
	}
	return out.Flush()
}

func renderProjectDetail(project client.Project) error {
	fmt.Println(project.Name)
	fmt.Println(strings.Repeat("=", len(project.Name)))
	fmt.Printf("slug: %s\n", project.Slug)
	if project.Descr != "" {
		fmt.Printf("\n%s\n", project.Descr)
	}
	if len(project.Members) > 0 {
		fmt.Println("\nmembers:")
		for _, member := range project.Members {
			who := member.Email
			if member.Name != "" {
				who = member.Name + " <" + member.Email + ">"
			}
			fmt.Printf("  %-8s %s\n", member.Role, who)
		}
	}
	return nil
}
