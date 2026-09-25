package pb

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain/rules"
)

// Register installs the folio schema. It is idempotent: an existing
// collection is left alone, so the app can call it on every boot.
//
// Access rules are expressed in PocketBase rule syntax and enforced by
// PocketBase itself for any direct REST access to these collections. The
// folio API additionally checks permissions in the service layer, so the
// rules here are the second line of defence rather than the only one.
// The rules reference journ_members, which cannot be resolved while that
// collection does not yet exist, so structure is created first and the access
// rules are applied in a second pass once every collection is present.
func Register(app core.App) error {
	if err := ensureClients(app); err != nil {
		return fmt.Errorf("clients: %w", err)
	}
	if err := ensureDomains(app); err != nil {
		return fmt.Errorf("domains: %w", err)
	}
	if err := ensureProjects(app); err != nil {
		return fmt.Errorf("projects: %w", err)
	}
	if err := ensureMembers(app); err != nil {
		return fmt.Errorf("members: %w", err)
	}
	// Plans come before issues, cycles before entries: each carries a
	// relation to the one above it, so the target has to exist first.
	if err := ensurePlans(app); err != nil {
		return fmt.Errorf("plans: %w", err)
	}
	if err := ensureIssues(app); err != nil {
		return fmt.Errorf("issues: %w", err)
	}
	if err := ensureIssueLinks(app); err != nil {
		return fmt.Errorf("issue links: %w", err)
	}
	if err := ensurePlanIssue(app); err != nil {
		return fmt.Errorf("plan issue: %w", err)
	}
	if err := ensureShares(app); err != nil {
		return fmt.Errorf("shares: %w", err)
	}
	if err := ensureCycles(app); err != nil {
		return fmt.Errorf("cycles: %w", err)
	}
	if err := ensureCycleMap(app); err != nil {
		return fmt.Errorf("cycle map: %w", err)
	}
	if err := ensureEntries(app); err != nil {
		return fmt.Errorf("entries: %w", err)
	}
	if err := ensureEntryKinds(app); err != nil {
		return fmt.Errorf("entry kinds: %w", err)
	}
	if err := ensureIssueResolution(app); err != nil {
		return fmt.Errorf("issue resolution: %w", err)
	}
	if err := ensureKnowledge(app); err != nil {
		return fmt.Errorf("knowledge: %w", err)
	}
	if err := ensureTags(app); err != nil {
		return fmt.Errorf("tags: %w", err)
	}
	if err := ensureTagJoins(app); err != nil {
		return fmt.Errorf("tag joins: %w", err)
	}
	if err := backfillDomains(app); err != nil {
		return fmt.Errorf("backfill domains: %w", err)
	}
	if err := backfillMemberDomains(app); err != nil {
		return fmt.Errorf("backfill member domains: %w", err)
	}
	if err := backfillTags(app); err != nil {
		return fmt.Errorf("backfill tags: %w", err)
	}
	// Last of the data steps: it refuses unless every legacy row was copied.
	if err := dropLegacy(app); err != nil {
		return fmt.Errorf("drop legacy: %w", err)
	}
	if err := applyRules(app); err != nil {
		return fmt.Errorf("rules: %w", err)
	}
	// Last: the index mirrors every collection above, so they all have to
	// exist before its triggers can reference them.
	if err := ensureSearchIndex(app); err != nil {
		return fmt.Errorf("search index: %w", err)
	}
	return nil
}

// backfillDomains gives every existing project a client and a domain. Each
// project becomes its own client rather than sharing one placeholder: merging
// by hand can undo that, a shared placeholder could not.
func backfillDomains(app core.App) error {
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	if projects.Fields.GetByName("domain") == nil {
		projects.Fields.Add(&core.RelationField{Name: "domain", CollectionId: domains.Id, CascadeDelete: false, MaxSelect: 1})
		if err := app.Save(projects); err != nil {
			return fmt.Errorf("add domain field: %w", err)
		}
	}

	records, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	clientSlugs, err := existingSlugs(app, ColClients)
	if err != nil {
		return err
	}
	domainSlugs, err := existingSlugs(app, ColDomains)
	if err != nil {
		return err
	}

	clients, err := app.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return err
	}

	for _, project := range records {
		if project.GetString("domain") != "" {
			continue
		}

		base := project.GetString("slug")
		if base == "" {
			base = rules.Slugify(project.GetString("name"))
		}

		clientSlug := uniqueSlug(base, clientSlugs)
		clientSlugs[clientSlug] = true
		client := core.NewRecord(clients)
		client.Set("slug", clientSlug)
		client.Set("name", project.GetString("name"))
		if err := app.Save(client); err != nil {
			return fmt.Errorf("create client for %s: %w", project.Id, err)
		}

		domainSlug := uniqueSlug(base, domainSlugs)
		domainSlugs[domainSlug] = true
		domain := core.NewRecord(domains)
		domain.Set("client", client.Id)
		domain.Set("slug", domainSlug)
		domain.Set("name", project.GetString("name"))
		domain.Set("descr", project.GetString("descr"))
		if err := app.Save(domain); err != nil {
			return fmt.Errorf("create domain for %s: %w", project.Id, err)
		}

		project.Set("domain", domain.Id)
		if err := app.Save(project); err != nil {
			return fmt.Errorf("link project %s to domain: %w", project.Id, err)
		}
	}

	return nil
}

// backfillMemberDomains moves membership from the project to the domain. The
// project column is kept until the readers stop using it, so a rollback still
// finds its memberships. Two roles collapsing into one row keep the stronger:
// silently demoting someone is the worse failure.
func backfillMemberDomains(app core.App) error {
	members, err := app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}

	if members.Fields.GetByName("domain") == nil {
		members.Fields.Add(&core.RelationField{Name: "domain", CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1})
		// (project, user) would reject a user joining a second project in the
		// same domain.
		members.RemoveIndex("idx_journ_members_unique")
		if err := app.Save(members); err != nil {
			return fmt.Errorf("add domain field: %w", err)
		}
	}

	rows, err := app.FindAllRecords(ColMembers)
	if err != nil {
		return fmt.Errorf("load members: %w", err)
	}

	projectDomain := make(map[string]string)
	projects, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	for _, p := range projects {
		projectDomain[p.Id] = p.GetString("domain")
	}

	rank := map[string]int{"viewer": 0, "editor": 1, "owner": 2}
	type key struct{ domain, user string }
	seen := make(map[key]*core.Record)

	for _, row := range rows {
		domainID := row.GetString("domain")
		if domainID == "" {
			domainID = projectDomain[row.GetString("project")]
			if domainID == "" {
				return fmt.Errorf("member %s: project %s has no domain", row.Id, row.GetString("project"))
			}
			row.Set("domain", domainID)
			if err := app.Save(row); err != nil {
				return fmt.Errorf("set domain on member %s: %w", row.Id, err)
			}
		}

		k := key{domainID, row.GetString("user")}
		kept, ok := seen[k]
		if !ok {
			seen[k] = row
			continue
		}
		if rank[row.GetString("role")] > rank[kept.GetString("role")] {
			kept.Set("role", row.GetString("role"))
			if err := app.Save(kept); err != nil {
				return fmt.Errorf("promote member %s: %w", kept.Id, err)
			}
		}
		if err := app.Delete(row); err != nil {
			return fmt.Errorf("delete duplicate member %s: %w", row.Id, err)
		}
	}

	members, err = app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	if field, ok := members.Fields.GetByName("domain").(*core.RelationField); ok && !field.Required {
		field.Required = true
		members.AddIndex("idx_journ_members_domain_unique", true, "domain, user", "")
		if err := app.Save(members); err != nil {
			return fmt.Errorf("require domain: %w", err)
		}
	}

	return nil
}

var ticketStatusToIssue = map[string]string{
	"open": "open", "in_progress": "in_progress", "blocked": "blocked",
	"closed": "done", "cancelled": "cancelled",
}

var todoStatusToIssue = map[string]string{
	"pending": "open", "in_progress": "in_progress", "blocked": "blocked",
	"done": "done", "cancelled": "cancelled",
}

func backfillTags(app core.App) error {
	tags, err := app.FindCollectionByNameOrId(ColTags)
	if err != nil {
		return err
	}

	known := make(map[string]string)
	rows, err := app.FindAllRecords(ColTags)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	for _, row := range rows {
		known[row.GetString("domain")+"/"+row.GetString("slug")] = row.Id
	}

	tagID := func(domainID, name string) (string, error) {
		slug := rules.Slugify(name)
		if slug == "" {
			return "", nil
		}
		key := domainID + "/" + slug
		if id, ok := known[key]; ok {
			return id, nil
		}
		rec := core.NewRecord(tags)
		rec.Set("domain", domainID)
		rec.Set("slug", slug)
		rec.Set("name", name)
		if err := app.Save(rec); err != nil {
			return "", fmt.Errorf("create tag %q: %w", name, err)
		}
		known[key] = rec.Id
		return rec.Id, nil
	}

	for _, src := range []struct{ collection, join, field string }{
		{ColIssues, ColIssueTags, "issue"},
		{ColEntries, ColEntryTags, "entry"},
	} {
		join, err := app.FindCollectionByNameOrId(src.join)
		if err != nil {
			return err
		}
		linked := make(map[string]bool)
		existing, err := app.FindAllRecords(src.join)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.join, err)
		}
		for _, row := range existing {
			linked[row.GetString(src.field)+"/"+row.GetString("tag")] = true
		}

		records, err := app.FindAllRecords(src.collection)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.collection, err)
		}
		for _, rec := range records {
			names := strSlice(rec, "tags")
			if len(names) == 0 {
				continue
			}
			domainID := rec.GetString("domain")
			if domainID == "" {
				return fmt.Errorf("%s %s has no domain", src.collection, rec.Id)
			}
			for _, name := range names {
				id, err := tagID(domainID, name)
				if err != nil {
					return err
				}
				if id == "" || linked[rec.Id+"/"+id] {
					continue
				}
				row := core.NewRecord(join)
				row.Set(src.field, rec.Id)
				row.Set("tag", id)
				if err := app.Save(row); err != nil {
					return fmt.Errorf("link tag on %s: %w", rec.Id, err)
				}
				linked[rec.Id+"/"+id] = true
			}
		}
	}

	return nil
}

func existingSlugs(app core.App, collection string) (map[string]bool, error) {
	records, err := app.FindAllRecords(collection)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", collection, err)
	}
	taken := make(map[string]bool, len(records))
	for _, record := range records {
		if slug := record.GetString("slug"); slug != "" {
			taken[slug] = true
		}
	}
	return taken, nil
}

// find returns the collection if it already exists.
func find(app core.App, name string) (*core.Collection, bool) {
	c, err := app.FindCollectionByNameOrId(name)
	if err != nil || c == nil {
		return nil, false
	}
	return c, true
}

// uniqueSlug takes the lowest free numeric suffix. A title that slugifies to
// nothing (emoji, CJK) falls back to "entry".
func uniqueSlug(base string, taken map[string]bool) string {
	if base == "" {
		base = "entry"
	}
	// Leaves room for a suffix inside the 60-character column.
	if len(base) > 50 {
		base = strings.TrimRight(base[:50], "-")
	}
	if !taken[base] {
		return base
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !taken[candidate] {
			return candidate
		}
	}
}
