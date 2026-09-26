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

func domainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domain",
		Short:   "List and update domains",
		Aliases: []string{"domains"},
	}
	cmd.AddCommand(domainListCommand(), domainUpdateCommand())
	return cmd
}

func splitDomainRef(ref string) (string, string, error) {
	clientRef, domainRef, ok := strings.Cut(ref, "/")
	if !ok || clientRef == "" || domainRef == "" || strings.Contains(domainRef, "/") {
		return "", "", fmt.Errorf("domain %q: want <client>/<domain>", ref)
	}
	return clientRef, domainRef, nil
}

func findDomain(folio *client.Client, clientRef, domainRef string) (string, error) {
	domains, err := folio.ListDomains()
	if err != nil {
		return "", err
	}
	for _, d := range domains {
		if (d.ClientSlug == clientRef || d.ClientID == clientRef) && (d.Slug == domainRef || d.ID == domainRef) {
			return d.ID, nil
		}
	}
	return "", fmt.Errorf("no domain %s/%s among the domains you can see", clientRef, domainRef)
}

func domainListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the domains you can see",
		RunE: func(_ *cobra.Command, _ []string) error {
			folio, err := api()
			if err != nil {
				return err
			}
			domains, err := folio.ListDomains()
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(domains)
			}
			if len(domains) == 0 {
				fmt.Println("no domains")
				return nil
			}
			out := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintln(out, "DOMAIN\tNAME\tMEMBERS")
			for _, d := range domains {
				fmt.Fprintf(out, "%s/%s\t%s\t%d\n", d.ClientSlug, d.Slug, d.Name, len(d.Members))
			}
			return out.Flush()
		},
	}
}

func domainUpdateCommand() *cobra.Command {
	var slug, name, descr string

	cmd := &cobra.Command{
		Use:   "update <client>/<domain>",
		Short: "Update a domain you own, or one under a client you own",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			clientRef, domainRef, err := splitDomainRef(args[0])
			if err != nil {
				return err
			}
			in := client.DomainInput{}
			setIf(&in.Slug, slug)
			setIf(&in.Name, name)
			setIf(&in.Descr, descr)
			if in == (client.DomainInput{}) {
				return errors.New("nothing to update: pass at least one field")
			}

			folio, err := api()
			if err != nil {
				return err
			}
			d, err := folio.UpdateDomain(clientRef, domainRef, in)
			if err != nil {
				return err
			}
			if flagJSON {
				return encode(d)
			}
			return renderDomainDetail(d)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "new slug, unique within the client")
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringVar(&descr, "descr", "", "new description")

	return cmd
}

func renderDomainDetail(d client.DomainRecord) error {
	fmt.Println(d.Name)
	fmt.Println(strings.Repeat("=", len(d.Name)))
	fmt.Printf("domain: %s/%s\n", d.ClientSlug, d.Slug)
	if d.Descr != "" {
		fmt.Printf("\n%s\n", d.Descr)
	}
	renderMembers(d.Members)
	return nil
}
