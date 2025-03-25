package network

import (
	"context"
	"reconya-ai/db"
	"reconya-ai/internal/config"
	"reconya-ai/models"
)

type NetworkService struct {
	Config     *config.Config
	Repository db.NetworkRepository
}

func NewNetworkService(networkRepo db.NetworkRepository, cfg *config.Config) *NetworkService {
	return &NetworkService{
		Config:     cfg,
		Repository: networkRepo,
	}
}

func (s *NetworkService) Create(cidr string) (*models.Network, error) {
	network := &models.Network{CIDR: cidr}
	network, err := s.Repository.CreateOrUpdate(context.Background(), network)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (s *NetworkService) FindOrCreate(cidr string) (*models.Network, error) {
	network, err := s.Repository.FindByCIDR(context.Background(), cidr)
	if err == db.ErrNotFound {
		return s.Create(cidr)
	}
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (s *NetworkService) FindByCIDR(cidr string) (*models.Network, error) {
	network, err := s.Repository.FindByCIDR(context.Background(), cidr)
	if err == db.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (s *NetworkService) FindCurrent() (*models.Network, error) {
	network, err := s.FindByCIDR(s.Config.NetworkCIDR)
	if err != nil {
		return nil, err
	}
	return network, nil
}
