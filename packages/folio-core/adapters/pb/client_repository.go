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

type ClientRepository struct {
	app core.App
}

func NewClientRepository(app core.App) *ClientRepository {
	return &ClientRepository{app: app}
}

var _ ports.ClientRepository = (*ClientRepository)(nil)

func (r *ClientRepository) List(ctx context.Context, user domain.UserID) ([]domain.Client, error) {
	var ids idSet
	memberships, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"user": string(user)})
	if err != nil {
		return nil, mapErr(err)
	}
	for _, m := range memberships {
		ids.add(m.GetString("client"))
	}
	for _, reach := range []func(core.App, string) ([]*core.Record, error){rosterDomains, grantedDomains} {
		domains, err := reach(r.app, string(user))
		if err != nil {
			return nil, mapErr(err)
		}
		for _, d := range domains {
			ids.add(d.GetString("client"))
		}
	}
	if len(ids.ids) == 0 {
		return []domain.Client{}, nil
	}

	records, err := r.app.FindAllRecords(ColClients, dbx.In("id", ids.ids...))
	if err != nil {
		return nil, mapErr(err)
	}
	slices.SortFunc(records, func(a, b *core.Record) int { return cmp.Compare(a.GetString("slug"), b.GetString("slug")) })
	out := make([]domain.Client, 0, len(records))
	for _, rec := range records {
		c, err := r.toClient(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *ClientRepository) GetByID(ctx context.Context, id domain.ClientID) (domain.Client, error) {
	rec, err := r.app.FindRecordById(ColClients, string(id))
	if err != nil {
		return domain.Client{}, mapErr(err)
	}
	return r.toClient(rec)
}

func (r *ClientRepository) GetBySlug(ctx context.Context, slug string) (domain.Client, error) {
	rec, err := r.app.FindFirstRecordByData(ColClients, "slug", slug)
	if err != nil {
		return domain.Client{}, mapErr(err)
	}
	return r.toClient(rec)
}

func (r *ClientRepository) Update(ctx context.Context, c domain.Client) (domain.Client, error) {
	rec, err := r.app.FindRecordById(ColClients, string(c.ID))
	if err != nil {
		return domain.Client{}, mapErr(err)
	}
	if taken, err := r.app.FindFirstRecordByData(ColClients, "slug", c.Slug); err == nil && taken.Id != rec.Id {
		return domain.Client{}, domain.ErrConflict
	}
	rec.Set("slug", c.Slug)
	rec.Set("name", c.Name)
	rec.Set("site", c.Site)
	rec.Set("logo", c.Logo)
	rec.Set("descr", c.Descr)
	if err := r.app.Save(rec); err != nil {
		return domain.Client{}, mapErr(err)
	}
	return r.toClient(rec)
}

func (r *ClientRepository) toClient(rec *core.Record) (domain.Client, error) {
	owners, err := r.app.FindAllRecords(ColClientMembers, dbx.HashExp{"client": rec.Id, "role": "owner"})
	if err != nil {
		return domain.Client{}, mapErr(err)
	}
	ids := make([]domain.UserID, 0, len(owners))
	for _, o := range owners {
		ids = append(ids, domain.UserID(o.GetString("user")))
	}
	return domain.Client{
		ID:        domain.ClientID(rec.Id),
		Slug:      rec.GetString("slug"),
		Name:      rec.GetString("name"),
		Site:      rec.GetString("site"),
		Logo:      rec.GetString("logo"),
		Descr:     rec.GetString("descr"),
		Owners:    ids,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}
