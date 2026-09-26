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
		out = append(out, domain.Client{
			ID:        domain.ClientID(rec.Id),
			Slug:      rec.GetString("slug"),
			Name:      rec.GetString("name"),
			Site:      rec.GetString("site"),
			Logo:      rec.GetString("logo"),
			Descr:     rec.GetString("descr"),
			CreatedAt: rec.GetDateTime("created").Time(),
			UpdatedAt: rec.GetDateTime("updated").Time(),
		})
	}
	return out, nil
}
