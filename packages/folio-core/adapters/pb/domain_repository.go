package pb

import (
	"context"

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

func (r *DomainRepository) GetByID(ctx context.Context, id domain.DomainID) (domain.Domain, error) {
	rec, err := r.app.FindRecordById(ColDomains, string(id))
	if err != nil {
		return domain.Domain{}, mapErr(err)
	}
	return r.toDomain(rec)
}

func (r *DomainRepository) toDomain(rec *core.Record) (domain.Domain, error) {
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
		Slug:         rec.GetString("slug"),
		Name:         rec.GetString("name"),
		Descr:        rec.GetString("descr"),
		Members:      toMembers(r.app, roster),
		ClientOwners: clientOwners,
		CreatedAt:    rec.GetDateTime("created").Time(),
		UpdatedAt:    rec.GetDateTime("updated").Time(),
	}, nil
}
