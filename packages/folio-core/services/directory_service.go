package services

import (
	"context"
	"slices"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type ClientService struct {
	repo  ports.ClientRepository
	clock ports.Clock
}

func NewClientService(repo ports.ClientRepository, clock ports.Clock) *ClientService {
	return &ClientService{repo: repo, clock: clock}
}

var _ ports.ClientUseCase = (*ClientService)(nil)

func (s *ClientService) ListClients(ctx context.Context, actor ports.Actor) ([]domain.Client, error) {
	return s.repo.List(ctx, actor.UserID)
}

func (s *ClientService) UpdateClient(ctx context.Context, actor ports.Actor, ref string, in ports.UpdateClientInput) (domain.Client, error) {
	c, err := s.resolve(ctx, ref)
	if err != nil {
		return domain.Client{}, err
	}
	if !actor.Superuser && !rules.CanAdminClient(c, actor.UserID) {
		visible, err := s.repo.List(ctx, actor.UserID)
		if err != nil {
			return domain.Client{}, err
		}
		if slices.ContainsFunc(visible, func(v domain.Client) bool { return v.ID == c.ID }) {
			return domain.Client{}, domain.ErrForbidden
		}
		return domain.Client{}, domain.ErrNotFound
	}

	if in.Slug != nil {
		c.Slug = strings.TrimSpace(*in.Slug)
	}
	if in.Name != nil {
		c.Name = strings.TrimSpace(*in.Name)
	}
	if in.Site != nil {
		c.Site = strings.TrimSpace(*in.Site)
	}
	if in.Logo != nil {
		c.Logo = strings.TrimSpace(*in.Logo)
	}
	if in.Descr != nil {
		c.Descr = *in.Descr
	}
	if err := rules.ValidateClient(c); err != nil {
		return domain.Client{}, err
	}
	c.UpdatedAt = s.clock.Now()
	return s.repo.Update(ctx, c)
}

func (s *ClientService) resolve(ctx context.Context, ref string) (domain.Client, error) {
	c, err := s.repo.GetByID(ctx, domain.ClientID(ref))
	if err == nil || !notFound(err) {
		return c, err
	}
	return s.repo.GetBySlug(ctx, ref)
}

type DomainService struct {
	repo ports.DomainRepository
}

func NewDomainService(repo ports.DomainRepository) *DomainService {
	return &DomainService{repo: repo}
}

var _ ports.DomainUseCase = (*DomainService)(nil)

func (s *DomainService) ListDomains(ctx context.Context, actor ports.Actor) ([]domain.Domain, error) {
	return s.repo.List(ctx, actor.UserID)
}
