package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
)

func loginCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate and cache a token",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Resolve(flagURL, "")
			if err != nil {
				return err
			}

			if email == "" {
				value, err := prompt("Email: ")
				if err != nil {
					return err
				}
				email = value
			}

			if password == "" {
				value, err := promptPassword("Password: ")
				if err != nil {
					return err
				}
				password = value
			}

			if email == "" || password == "" {
				return errors.New("email and password are required")
			}

			session, err := client.New(cfg.URL, "").Login(email, password)
			if err != nil {
				return err
			}

			if err := config.Save(cfg.URL, session.Token); err != nil {
				return err
			}

			fmt.Printf("signed in as %s\n", session.Email)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "account email, prompted when absent")
	cmd.Flags().StringVar(&password, "password", "", "account password, prompted when absent")

	return cmd
}

func logoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Discard the cached token",
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Println("signed out")
			return nil
		},
	}
}

func prompt(label string) (string, error) {
	fmt.Print(label)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptPassword(label string) (string, error) {
	fmt.Print(label)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return prompt("")
	}

	secret, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(secret)), nil
}
