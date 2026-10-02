package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
	"folio/cli/internal/tui"
)

func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

var (
	flagURL     string
	flagToken   string
	flagProject string
	flagJSON    bool
)

func api() (*client.Client, error) {
	cfg, err := config.Resolve(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	return client.New(cfg.URL, cfg.Token), nil
}

func main() {
	root := &cobra.Command{
		Use:   "folio",
		Short: "Write to the folio project workspace",
		Long: `Write to the folio project workspace.

Exit codes: 0 ok, 1 other failure, 2 not signed in or not allowed, 3 not found,
4 rejected as invalid. With --json a failure prints
{"error","status","fields"} on stderr instead of a line of text.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Bare `folio` opens the browser on a terminal, but a pipe or a
			// CI run gets help rather than an alt-screen program it cannot
			// drive.
			if !isTerminal(os.Stdout) {
				return cmd.Help()
			}

			folio, err := api()
			if err != nil {
				return err
			}
			return tui.Run(folio, config.Project(flagProject))
		},
	}

	root.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "project id or slug")

	root.PersistentFlags().StringVar(&flagURL, "url", "", "folio base URL (default $FOLIO_URL or "+config.DefaultURL+")")
	root.PersistentFlags().StringVar(&flagToken, "token", "", "auth token (default $FOLIO_TOKEN or the cached login)")
	root.PersistentFlags().BoolVar(&flagJSON, "json", false, "output JSON instead of a table")

	root.AddCommand(
		clientCommand(),
		configCommand(),
		cycleCommand(),
		docCommand(),
		domainCommand(),
		interviewCommand(),
		journalCommand(),
		knowledgeCommand(),
		loginCommand(),
		logoutCommand(),
		planCommand(),
		projectCommand(),
		resumeCommand(),
		searchCommand(),
		stopCommand(),
		tagsCommand(),
		ticketCommand(),
		todoCommand(),
		useCommand(),
		workLogCommand(),
	)

	if err := root.Execute(); err != nil {
		if flagJSON {
			fmt.Fprintln(os.Stderr, errorJSON(err))
		} else {
			fmt.Fprintln(os.Stderr, "folio: "+err.Error())
		}
		os.Exit(exitCode(err))
	}
}
