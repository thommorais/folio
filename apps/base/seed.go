package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	folio "folio/folio-core"
	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

const (
	seedEmail    = "user@test.com"
	seedPassword = "pass@test"
)

func newSeedCommand(app *pocketbase.PocketBase) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Create the demo user and fill the workspace with example data",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}
			if err := folio.Migrate(app); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			// Refuse to add demo data to a database that is already in use.
			if !force {
				count, err := app.CountRecords("journ_projects")
				if err != nil {
					return err
				}
				if count > 0 {
					return fmt.Errorf("database already has %d project(s); re-run with --force if you meant to add the demo data anyway", count)
				}
			}

			user, err := ensureSeedUser(app)
			if err != nil {
				return fmt.Errorf("seed user: %w", err)
			}

			useCases := folio.New(app, nil)
			report, err := services.Seed(context.Background(), services.SeedUseCases{
				Projects: useCases.Projects,
				Plans:    useCases.Plans,
				Issues:   useCases.Issues,
				Journal:  useCases.Journal,
				Docs:     useCases.Docs,
			}, ports.Actor{UserID: domain.UserID(user.Id), Email: user.Email()})
			if err != nil {
				return err
			}

			fmt.Printf("seeded %s\n", report)
			fmt.Printf("sign in as %s / %s\n", seedEmail, seedPassword)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "seed even when the database already has projects")

	return cmd
}

func ensureSeedUser(app core.App) (*core.Record, error) {
	existing, err := app.FindAuthRecordByEmail("users", seedEmail)
	if err == nil {
		existing.SetPassword(seedPassword)
		existing.SetVerified(true)
		if err := app.Save(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}
	record := core.NewRecord(collection)
	record.SetEmail(seedEmail)
	record.SetPassword(seedPassword)
	record.SetVerified(true)
	record.Set("name", "Demo User")
	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}
