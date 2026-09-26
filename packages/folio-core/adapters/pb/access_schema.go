package pb

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func ensureClientMembers(app core.App) error {
	if _, ok := find(app, ColClientMembers); ok {
		return nil
	}
	clients, err := app.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColClientMembers)
	c.Fields.Add(
		&core.RelationField{Name: "client", Required: true, CollectionId: clients.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "user", Required: true, CollectionId: users.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "role", Required: true, MaxSelect: 1, Values: []string{"owner", "member"}},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_client_members_unique", true, "client, user", "")
	c.AddIndex("idx_journ_client_members_user", false, "user", "")
	if err := app.Save(c); err != nil {
		return err
	}

	return seedClientOwners(app, c)
}

func seedClientOwners(app core.App, c *core.Collection) error {
	owners, err := app.FindAllRecords(ColMembers)
	if err != nil {
		return fmt.Errorf("load domain members: %w", err)
	}
	seen := make(map[[2]string]bool)
	for _, m := range owners {
		if m.GetString("role") != "owner" {
			continue
		}
		d, err := app.FindRecordById(ColDomains, m.GetString("domain"))
		if err != nil {
			return fmt.Errorf("domain of member %s: %w", m.Id, err)
		}
		key := [2]string{d.GetString("client"), m.GetString("user")}
		if seen[key] {
			continue
		}
		seen[key] = true

		row := core.NewRecord(c)
		row.Set("client", key[0])
		row.Set("user", key[1])
		row.Set("role", "owner")
		if err := app.Save(row); err != nil {
			return fmt.Errorf("seed client owner: %w", err)
		}
	}
	return nil
}

func ensureProjectGrants(app core.App) error {
	if _, ok := find(app, ColProjectGrants); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColProjectGrants)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "domain", CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "user", CollectionId: users.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "role", Required: true, MaxSelect: 1, Values: []string{"owner", "editor", "viewer"}},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_project_grants_domain", true, "project, domain", "domain != ''")
	c.AddIndex("idx_journ_project_grants_user", true, "project, user", "user != ''")
	if err := app.Save(c); err != nil {
		return err
	}

	return seedDomainGrants(app, c)
}

func seedDomainGrants(app core.App, c *core.Collection) error {
	projects, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	for _, p := range projects {
		row := core.NewRecord(c)
		row.Set("project", p.Id)
		row.Set("domain", p.GetString("domain"))
		row.Set("role", "editor")
		if err := app.Save(row); err != nil {
			return fmt.Errorf("seed grant for project %s: %w", p.Id, err)
		}
	}
	return nil
}
