package services

import (
	"context"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type ClientService struct {
	repo ports.ClientRepository
}

func NewClientService(repo ports.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

var _ ports.ClientUseCase = (*ClientService)(nil)

func (s *ClientService) ListClients(ctx context.Context, actor ports.Actor) ([]domain.Client, error) {
	return s.repo.List(ctx, actor.UserID)
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
