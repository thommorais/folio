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

func clientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Short:   "List and update clients",
		Aliases: []string{"clients"},
	}
	cmd.AddCommand(clientListCommand(), clientUpdateCommand())
	return cmd
}

func clientListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the clients you can see",
		RunE: func(_ *cobra.Command, _ []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			clients, err := folio.ListClients()
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(clients)
			}
			if len(clients) == 0 {
				fmt.Println("no clients")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(out, "SLUG\tNAME\tSITE")
			for _, c := range clients {
				fmt.Fprintf(out, "%s\t%s\t%s\n", c.Slug, c.Name, c.Site)
			}
			return out.Flush()
		},
	}
}

func clientUpdateCommand() *cobra.Command {
	var slug, name, site, logo, descr string

	cmd := &cobra.Command{
		Use:   "update <id-or-slug>",
		Short: "Update a client you own",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			in := client.ClientInput{}
			setIf(&in.Slug, slug)
			setIf(&in.Name, name)
			setIf(&in.Site, site)
			setIf(&in.Logo, logo)
			setIf(&in.Descr, descr)
			if in == (client.ClientInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			c, err := folio.UpdateClient(args[0], in)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(c)
			}
			return renderClientDetail(c)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "new slug")
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringVar(&site, "site", "", "new website")
	cmd.Flags().StringVar(&logo, "logo", "", "new logo URL")
	cmd.Flags().StringVar(&descr, "descr", "", "new description")

	return cmd
}

func renderClientDetail(c client.ClientRecord) error {
	fmt.Println(c.Name)
	fmt.Println(strings.Repeat("=", len(c.Name)))
	fmt.Printf("slug: %s\n", c.Slug)
	if c.Site != "" {
		fmt.Printf("site: %s\n", c.Site)
	}
	if c.Logo != "" {
		fmt.Printf("logo: %s\n", c.Logo)
	}
	if c.Descr != "" {
		fmt.Printf("\n%s\n", c.Descr)
	}
	return nil
}
