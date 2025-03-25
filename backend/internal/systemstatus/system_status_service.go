package systemstatus

import (
	"context"
	"reconya-ai/db"
	"reconya-ai/models"
	"time"
)

type SystemStatusService struct {
	repository db.SystemStatusRepository
}

func NewSystemStatusService(repository db.SystemStatusRepository) *SystemStatusService {
	return &SystemStatusService{
		repository: repository,
	}
}

func (s *SystemStatusService) GetLatest() (*models.SystemStatus, error) {
	systemStatus, err := s.repository.FindLatest(context.Background())
	if err == db.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return systemStatus, nil
}

func (s *SystemStatusService) CreateOrUpdate(systemStatus *models.SystemStatus) (*models.SystemStatus, error) {
	now := time.Now()
	systemStatus.CreatedAt = now
	systemStatus.UpdatedAt = now

	// Create the system status
	if err := s.repository.Create(context.Background(), systemStatus); err != nil {
		return nil, err
	}

	// Return the latest system status
	return s.GetLatest()
}
