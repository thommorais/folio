package pb

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// ProjectRepository stores projects and their grants. It runs
// in-process against core.App, so it bypasses PocketBase's own API rules;
// authorisation is the service layer's job (see services.ProjectGuard).
type ProjectRepository struct {
	app core.App
}

func NewProjectRepository(app core.App) *ProjectRepository {
	return &ProjectRepository{app: app}
}

var _ ports.ProjectRepository = (*ProjectRepository)(nil)

func (r *ProjectRepository) toProject(rec *core.Record) (domain.Project, error) {
	p := domain.Project{
		ID:        domain.ProjectID(rec.Id),
		DomainID:  domain.DomainID(rec.GetString("domain")),
		Slug:      rec.GetString("slug"),
		Name:      rec.GetString("name"),
		Descr:     rec.GetString("descr"),
		Archived:  rec.GetBool("archived"),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}

	owners, err := r.clientOwnersOf(rec.GetString("domain"))
	if err != nil {
		return domain.Project{}, err
	}
	p.ClientOwners = owners

	grants, err := r.app.FindAllRecords(ColProjectGrants, dbx.HashExp{"project": rec.Id})
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	var personal []*core.Record
	for _, g := range grants {
		if g.GetString("user") != "" {
			personal = append(personal, g)
			continue
		}
		roster, err := r.rosterOf(g.GetString("domain"))
		if err != nil {
			return domain.Project{}, err
		}
		p.DomainGrants = append(p.DomainGrants, domain.DomainGrant{
			DomainID: domain.DomainID(g.GetString("domain")),
			Role:     domain.Role(g.GetString("role")),
			Members:  roster,
		})
	}
	p.Members = toMembers(r.app, personal)
	return p, nil
}

func (r *ProjectRepository) clientOwnersOf(domainID string) ([]domain.UserID, error) {
	if domainID == "" {
		return nil, nil
	}
	d, err := r.app.FindRecordById(ColDomains, domainID)
	if err != nil {
		return nil, mapErr(err)
	}
	rows, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"client": d.GetString("client"), "role": "owner"})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.UserID, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.UserID(row.GetString("user")))
	}
	return out, nil
}

func (r *ProjectRepository) rosterOf(domainID string) ([]domain.UserID, error) {
	rows, err := r.app.FindAllRecords(ColMembers, dbx.HashExp{"domain": domainID})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.UserID, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.UserID(row.GetString("user")))
	}
	return out, nil
}

// A user that cannot be expanded (a deleted account) leaves that member's
// email blank rather than failing the whole roster.
func toMembers(app core.App, rows []*core.Record) []domain.Member {
	app.ExpandRecords(rows, []string{"user"}, nil)
	out := make([]domain.Member, 0, len(rows))
	for _, row := range rows {
		m := domain.Member{
			UserID: domain.UserID(row.GetString("user")),
			Role:   domain.Role(row.GetString("role")),
		}
		if user := row.ExpandedOne("user"); user != nil {
			m.Email = user.GetString("email")
			m.Name = user.GetString("name")
		}
		out = append(out, m)
	}
	return out
}

func (r *ProjectRepository) List(ctx context.Context, actor domain.UserID, includeArchived bool) ([]domain.Project, error) {
	ids, err := r.reachable(string(actor))
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []domain.Project{}, nil
	}

	exprs := []dbx.Expression{dbx.In("id", ids...)}
	if !includeArchived {
		exprs = append(exprs, dbx.HashExp{"archived": false})
	}
	records, err := r.app.FindAllRecords(ColProjects, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}

	out := make([]domain.Project, 0, len(records))
	for _, rec := range records {
		p, err := r.toProject(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *ProjectRepository) reachable(user string) ([]any, error) {
	seen := map[string]bool{}
	var ids []any
	add := func(id string) {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	owned, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"user": user, "role": "owner"})
	if err != nil {
		return nil, mapErr(err)
	}
	if len(owned) > 0 {
		clients := make([]any, 0, len(owned))
		for _, row := range owned {
			clients = append(clients, row.GetString("client"))
		}
		domains, err := r.app.FindAllRecords(ColDomains, dbx.In("client", clients...))
		if err != nil {
			return nil, mapErr(err)
		}
		if len(domains) > 0 {
			domainIDs := make([]any, 0, len(domains))
			for _, d := range domains {
				domainIDs = append(domainIDs, d.Id)
			}
			projects, err := r.app.FindAllRecords(ColProjects, dbx.In("domain", domainIDs...))
			if err != nil {
				return nil, mapErr(err)
			}
			for _, p := range projects {
				add(p.Id)
			}
		}
	}

	rosters, err := r.app.FindAllRecords(ColMembers, dbx.HashExp{"user": user})
	if err != nil {
		return nil, mapErr(err)
	}
	grantFilters := []dbx.Expression{dbx.HashExp{"user": user}}
	if len(rosters) > 0 {
		domainIDs := make([]any, 0, len(rosters))
		for _, row := range rosters {
			domainIDs = append(domainIDs, row.GetString("domain"))
		}
		grantFilters = append(grantFilters, dbx.In("domain", domainIDs...))
	}
	grants, err := r.app.FindAllRecords(ColProjectGrants, dbx.Or(grantFilters...))
	if err != nil {
		return nil, mapErr(err)
	}
	for _, g := range grants {
		add(g.GetString("project"))
	}
	return ids, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	rec, err := r.app.FindRecordById(ColProjects, string(id))
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	return r.toProject(rec)
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	rec, err := r.app.FindFirstRecordByData(ColProjects, "slug", slug)
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	return r.toProject(rec)
}

// Create writes the project and its grants in one transaction: a project
// without an owner would be unreachable and unadministrable.
func (r *ProjectRepository) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	if existing, err := r.app.FindFirstRecordByData(ColProjects, "slug", p.Slug); err == nil && existing != nil {
		return domain.Project{}, domain.ErrConflict
	}

	collection, err := r.app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	grantCollection, err := r.app.FindCollectionByNameOrId(ColProjectGrants)
	if err != nil {
		return domain.Project{}, mapErr(err)
	}

	rec := core.NewRecord(collection)
	rec.Id = string(p.ID)
	rec.Set("slug", p.Slug)
	rec.Set("name", p.Name)
	rec.Set("descr", p.Descr)
	rec.Set("archived", p.Archived)

	err = r.app.RunInTransaction(func(tx core.App) error {
		domainID := string(p.DomainID)
		var clientID string
		if domainID == "" {
			var err error
			if clientID, domainID, err = newOwnDomain(tx, p); err != nil {
				return err
			}
		} else if _, err := tx.FindRecordById(ColDomains, domainID); err != nil {
			return mapErr(err)
		}
		rec.Set("domain", domainID)

		if err := tx.Save(rec); err != nil {
			return err
		}
		if clientID != "" {
			if err := ownOwnDomain(tx, clientID, domainID, rec.Id, p.Owners()); err != nil {
				return err
			}
		}

		byDomain := core.NewRecord(grantCollection)
		byDomain.Set("project", rec.Id)
		byDomain.Set("domain", domainID)
		byDomain.Set("role", string(domain.RoleEditor))
		if err := tx.Save(byDomain); err != nil {
			return err
		}
		for _, m := range p.Members {
			row := core.NewRecord(grantCollection)
			row.Set("project", rec.Id)
			row.Set("user", string(m.UserID))
			row.Set("role", string(m.Role))
			if err := tx.Save(row); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	return r.GetByID(ctx, domain.ProjectID(rec.Id))
}

// newOwnDomain gives a project created without a domain its own client and
// domain.
func newOwnDomain(tx core.App, p domain.Project) (string, string, error) {
	clients, err := tx.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return "", "", err
	}
	domains, err := tx.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return "", "", err
	}

	clientSlugs, err := existingSlugs(tx, ColClients)
	if err != nil {
		return "", "", err
	}
	client := core.NewRecord(clients)
	client.Set("slug", uniqueSlug(p.Slug, clientSlugs))
	client.Set("name", p.Name)
	if err := tx.Save(client); err != nil {
		return "", "", err
	}

	domainSlugs, err := existingSlugs(tx, ColDomains)
	if err != nil {
		return "", "", err
	}
	d := core.NewRecord(domains)
	d.Set("client", client.Id)
	d.Set("slug", uniqueSlug(p.Slug, domainSlugs))
	d.Set("name", p.Name)
	d.Set("descr", p.Descr)
	if err := tx.Save(d); err != nil {
		return "", "", err
	}
	return client.Id, d.Id, nil
}

func ownOwnDomain(tx core.App, clientID, domainID, projectID string, owners []domain.Member) error {
	clientMembers, err := tx.FindCollectionByNameOrId(ColClientMembers)
	if err != nil {
		return err
	}
	roster, err := tx.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	for _, m := range owners {
		owner := core.NewRecord(clientMembers)
		owner.Set("client", clientID)
		owner.Set("user", string(m.UserID))
		owner.Set("role", "owner")
		if err := tx.Save(owner); err != nil {
			return err
		}
		row := core.NewRecord(roster)
		row.Set("domain", domainID)
		row.Set("project", projectID)
		row.Set("user", string(m.UserID))
		row.Set("role", "owner")
		if err := tx.Save(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectRepository) Update(ctx context.Context, p domain.Project) (domain.Project, error) {
	rec, err := r.app.FindRecordById(ColProjects, string(p.ID))
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	rec.Set("name", p.Name)
	rec.Set("descr", p.Descr)
	rec.Set("archived", p.Archived)
	if err := r.app.Save(rec); err != nil {
		return domain.Project{}, mapErr(err)
	}
	return r.GetByID(ctx, p.ID)
}

func (r *ProjectRepository) Delete(ctx context.Context, id domain.ProjectID) error {
	rec, err := r.app.FindRecordById(ColProjects, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func (r *ProjectRepository) AddMember(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error {
	if row, err := r.grantRow(id, user); err == nil {
		row.Set("role", string(role))
		return mapErr(r.app.Save(row))
	} else if !isNotFound(err) {
		return err
	}

	if _, err := r.app.FindRecordById(ColProjects, string(id)); err != nil {
		return mapErr(err)
	}
	collection, err := r.app.FindCollectionByNameOrId(ColProjectGrants)
	if err != nil {
		return mapErr(err)
	}
	row := core.NewRecord(collection)
	row.Set("project", string(id))
	row.Set("user", string(user))
	row.Set("role", string(role))
	return mapErr(r.app.Save(row))
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, id domain.ProjectID, user domain.UserID) error {
	row, err := r.grantRow(id, user)
	if err != nil {
		return err
	}
	return mapErr(r.app.Delete(row))
}

func (r *ProjectRepository) SetMemberRole(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error {
	row, err := r.grantRow(id, user)
	if err != nil {
		return err
	}
	row.Set("role", string(role))
	return mapErr(r.app.Save(row))
}

func (r *ProjectRepository) grantRow(id domain.ProjectID, user domain.UserID) (*core.Record, error) {
	row, err := r.app.FindFirstRecordByFilter(
		ColProjectGrants,
		"project = {:project} && user = {:user}",
		dbx.Params{"project": string(id), "user": string(user)},
	)
	if err != nil {
		return nil, mapErr(err)
	}
	return row, nil
}

func (r *ProjectRepository) FindUserByEmail(ctx context.Context, email string) (domain.UserID, error) {
	rec, err := r.app.FindAuthRecordByEmail(ColUsers, email)
	if err != nil {
		return "", mapErr(err)
	}
	return domain.UserID(rec.Id), nil
}
