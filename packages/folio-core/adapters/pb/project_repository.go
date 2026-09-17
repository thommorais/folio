package pb

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// ProjectRepository stores projects and their membership rows. It runs
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
	members, err := r.membersOf(rec.GetString("domain"))
	if err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:        domain.ProjectID(rec.Id),
		DomainID:  domain.DomainID(rec.GetString("domain")),
		Slug:      rec.GetString("slug"),
		Name:      rec.GetString("name"),
		Descr:     rec.GetString("descr"),
		Archived:  rec.GetBool("archived"),
		Members:   members,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

// membersOf loads the roster, expanding each row's user so callers get an
// email and name without a second round trip.
func (r *ProjectRepository) membersOf(domainID string) ([]domain.Member, error) {
	if domainID == "" {
		return nil, nil
	}
	rows, err := r.app.FindAllRecords(ColMembers, dbx.HashExp{"domain": domainID})
	if err != nil {
		return nil, mapErr(err)
	}
	// A user that cannot be expanded (a deleted account) leaves that member's
	// email blank rather than failing the whole roster.
	r.app.ExpandRecords(rows, []string{"user"}, nil)
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
	return out, nil
}

// List returns the projects the actor belongs to, resolved through their
// membership rows so a non-member never appears in anyone's listing.
func (r *ProjectRepository) List(ctx context.Context, actor domain.UserID, includeArchived bool) ([]domain.Project, error) {
	memberships, err := r.app.FindAllRecords(ColMembers, dbx.HashExp{"user": string(actor)})
	if err != nil {
		return nil, mapErr(err)
	}
	if len(memberships) == 0 {
		return []domain.Project{}, nil
	}
	ids := make([]any, 0, len(memberships))
	for _, m := range memberships {
		ids = append(ids, m.GetString("domain"))
	}

	exprs := []dbx.Expression{dbx.In("domain", ids...)}
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

// Create writes the project and its first membership row in one transaction:
// a project without an owner would be unreachable and unadministrable.
func (r *ProjectRepository) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	if existing, err := r.app.FindFirstRecordByData(ColProjects, "slug", p.Slug); err == nil && existing != nil {
		return domain.Project{}, domain.ErrConflict
	}

	collection, err := r.app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return domain.Project{}, mapErr(err)
	}
	memberCollection, err := r.app.FindCollectionByNameOrId(ColMembers)
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
		// A project without a domain has no roster and would be
		// unadministrable, so one is created for it.
		domainID := string(p.DomainID)
		if domainID == "" {
			var err error
			if domainID, err = ensureOwnDomain(tx, p); err != nil {
				return err
			}
		} else if _, err := tx.FindRecordById(ColDomains, domainID); err != nil {
			return mapErr(err)
		}
		rec.Set("domain", domainID)

		if err := tx.Save(rec); err != nil {
			return err
		}
		// Joining an existing domain must not grant access through its own
		// member list.
		for _, m := range p.Members {
			existing, err := tx.FindFirstRecordByFilter(
				ColMembers,
				"domain = {:domain} && user = {:user}",
				dbx.Params{"domain": domainID, "user": string(m.UserID)},
			)
			if err == nil && existing != nil {
				continue
			}
			row := core.NewRecord(memberCollection)
			row.Set("domain", domainID)
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

func ensureOwnDomain(tx core.App, p domain.Project) (string, error) {
	clients, err := tx.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return "", err
	}
	domains, err := tx.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return "", err
	}

	clientSlugs, err := existingSlugs(tx, ColClients)
	if err != nil {
		return "", err
	}
	client := core.NewRecord(clients)
	client.Set("slug", uniqueSlug(p.Slug, clientSlugs))
	client.Set("name", p.Name)
	if err := tx.Save(client); err != nil {
		return "", err
	}

	domainSlugs, err := existingSlugs(tx, ColDomains)
	if err != nil {
		return "", err
	}
	d := core.NewRecord(domains)
	d.Set("client", client.Id)
	d.Set("slug", uniqueSlug(p.Slug, domainSlugs))
	d.Set("name", p.Name)
	d.Set("descr", p.Descr)
	if err := tx.Save(d); err != nil {
		return "", err
	}
	return d.Id, nil
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
	if row, err := r.memberRow(id, user); err == nil {
		row.Set("role", string(role))
		return mapErr(r.app.Save(row))
	} else if !isNotFound(err) {
		return err
	}

	collection, err := r.app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return mapErr(err)
	}
	domainID, err := r.domainOf(id)
	if err != nil {
		return err
	}
	row := core.NewRecord(collection)
	row.Set("domain", domainID)
	row.Set("project", string(id))
	row.Set("user", string(user))
	row.Set("role", string(role))
	return mapErr(r.app.Save(row))
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, id domain.ProjectID, user domain.UserID) error {
	row, err := r.memberRow(id, user)
	if err != nil {
		return err
	}
	return mapErr(r.app.Delete(row))
}

func (r *ProjectRepository) SetMemberRole(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error {
	row, err := r.memberRow(id, user)
	if err != nil {
		return err
	}
	row.Set("role", string(role))
	return mapErr(r.app.Save(row))
}

func (r *ProjectRepository) memberRow(id domain.ProjectID, user domain.UserID) (*core.Record, error) {
	domainID, err := r.domainOf(id)
	if err != nil {
		return nil, err
	}
	row, err := r.app.FindFirstRecordByFilter(
		ColMembers,
		"domain = {:domain} && user = {:user}",
		dbx.Params{"domain": domainID, "user": string(user)},
	)
	if err != nil {
		return nil, mapErr(err)
	}
	return row, nil
}

func (r *ProjectRepository) domainOf(id domain.ProjectID) (string, error) {
	rec, err := r.app.FindRecordById(ColProjects, string(id))
	if err != nil {
		return "", mapErr(err)
	}
	return rec.GetString("domain"), nil
}

func (r *ProjectRepository) FindUserByEmail(ctx context.Context, email string) (domain.UserID, error) {
	rec, err := r.app.FindAuthRecordByEmail(ColUsers, email)
	if err != nil {
		return "", mapErr(err)
	}
	return domain.UserID(rec.Id), nil
}
