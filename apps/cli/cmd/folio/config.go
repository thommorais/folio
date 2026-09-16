package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"folio/cli/internal/config"
)

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write CLI configuration",
	}

	cmd.AddCommand(configSetCommand(), configGetCommand(), configPathCommand())

	return cmd
}

func configSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set url <value>",
		Short: "Persist the API base URL",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if args[0] != "url" {
				return fmt.Errorf("unknown key %q: only url is settable (project comes from %s)", args[0], config.EnvProject)
			}

			if err := config.SetURL(args[1]); err != nil {
				return err
			}

			cfg, err := config.Resolve("", "")
			if err != nil {
				return err
			}
			fmt.Println("url = " + cfg.URL)
			return nil
		},
	}
}

func configGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Print the resolved configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.Resolve(flagURL, flagToken)
			if err != nil {
				return err
			}

			values := map[string]string{
				"url":     cfg.URL,
				"project": config.Project(flagProject),
				"token":   redact(cfg.Token),
			}

			if len(args) == 1 {
				value, ok := values[args[0]]
				if !ok {
					return errors.New("unknown key " + args[0] + ": try url, project or token")
				}
				fmt.Println(value)
				return nil
			}

			for _, key := range []string{"url", "project", "token"} {
				fmt.Printf("%-8s %s\n", key, values[key])
			}
			return nil
		},
	}
}

func configPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print where the configuration is stored",
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
}

// The value is a bearer credential, so only its presence is reported.
func redact(token string) string {
	if token == "" {
		return "(not set)"
	}
	return "(set)"
}
