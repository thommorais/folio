package pb

import (
	"cmp"
	"context"
	"slices"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type DomainRepository struct {
	app core.App
}

func NewDomainRepository(app core.App) *DomainRepository {
	return &DomainRepository{app: app}
}

var _ ports.DomainRepository = (*DomainRepository)(nil)

func (r *DomainRepository) List(ctx context.Context, user domain.UserID) ([]domain.Domain, error) {
	var ids idSet
	owned, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"user": string(user), "role": "owner"})
	if err != nil {
		return nil, mapErr(err)
	}
	if len(owned) > 0 {
		domains, err := r.app.FindAllRecords(ColDomains, dbx.In("client", column(owned, "client")...))
		if err != nil {
			return nil, mapErr(err)
		}
		for _, d := range domains {
			ids.add(d.Id)
		}
	}
	for _, reach := range []func(core.App, string) ([]*core.Record, error){rosterDomains, grantedDomains} {
		domains, err := reach(r.app, string(user))
		if err != nil {
			return nil, mapErr(err)
		}
		for _, d := range domains {
			ids.add(d.Id)
		}
	}
	if len(ids.ids) == 0 {
		return []domain.Domain{}, nil
	}

	records, err := r.app.FindAllRecords(ColDomains, dbx.In("id", ids.ids...))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Domain, 0, len(records))
	for _, rec := range records {
		d, err := r.toDomain(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	slices.SortFunc(out, func(a, b domain.Domain) int {
		return cmp.Or(cmp.Compare(a.ClientSlug, b.ClientSlug), cmp.Compare(a.Slug, b.Slug))
	})
	return out, nil
}

func (r *DomainRepository) GetByID(ctx context.Context, id domain.DomainID) (domain.Domain, error) {
	rec, err := r.app.FindRecordById(ColDomains, string(id))
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	return r.toDomain(rec)
}

func (r *DomainRepository) GetBySlug(ctx context.Context, client domain.ClientID, slug string) (domain.Domain, error) {
	rec, err := r.app.FindFirstRecordByFilter(ColDomains, "client = {:client} && slug = {:slug}",
		dbx.Params{"client": string(client), "slug": slug})
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	return r.toDomain(rec)
}

func (r *DomainRepository) Update(ctx context.Context, d domain.Domain) (domain.Domain, error) {
	rec, err := r.app.FindRecordById(ColDomains, string(d.ID))
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	taken, err := r.app.FindFirstRecordByFilter(ColDomains, "client = {:client} && slug = {:slug}",
		dbx.Params{"client": rec.GetString("client"), "slug": d.Slug})
	if err == nil && taken.Id != rec.Id {
		return domain.Domain{}, domain.ErrConflict
	}
	rec.Set("slug", d.Slug)
	rec.Set("name", d.Name)
	rec.Set("descr", d.Descr)
	if err := r.app.Save(rec); err != nil {
		return domain.Domain{}, mapErr(err)
	}
	return r.toDomain(rec)
}

func (r *DomainRepository) toDomain(rec *core.Record) (domain.Domain, error) {
	client, err := r.app.FindRecordById(ColClients, rec.GetString("client"))
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	roster, err := r.app.FindAllRecords(ColMembers, dbx.HashExp{"domain": rec.Id})
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	owners, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"client": rec.GetString("client"), "role": "owner"})
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	clientOwners := make([]domain.UserID, 0, len(owners))
	for _, o := range owners {
		clientOwners = append(clientOwners, domain.UserID(o.GetString("user")))
	}
	return domain.Domain{
		ID:           domain.DomainID(rec.Id),
		ClientID:     domain.ClientID(rec.GetString("client")),
		ClientSlug:   client.GetString("slug"),
		Slug:         rec.GetString("slug"),
		Name:         rec.GetString("name"),
		Descr:        rec.GetString("descr"),
		Members:      toMembers(r.app, roster),
		ClientOwners: clientOwners,
		CreatedAt:    rec.GetDateTime("created").Time(),
		UpdatedAt:    rec.GetDateTime("updated").Time(),
	}, nil
}
