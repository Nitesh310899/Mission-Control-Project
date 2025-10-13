package commander

import (
	"mission-control/pkg/models"
	"sync"
	"time"
)

type MissionStore struct {
	missions map[string]*models.Mission
	lock     sync.RWMutex
}

func NewMissionStore() *MissionStore {
	return &MissionStore{
		missions: make(map[string]*models.Mission),
	}
}

func (s *MissionStore) CreateMission(id, payload string) *models.Mission {
	m := &models.Mission{
		ID:        id,
		Payload:   payload,
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.lock.Lock()
	s.missions[id] = m
	s.lock.Unlock()
	return m
}

func (s *MissionStore) UpdateStatus(id string, status models.MissionStatus) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if m, ok := s.missions[id]; ok {
		m.Status = status
		m.UpdatedAt = time.Now()
	}
}

func (s *MissionStore) GetMission(id string) (*models.Mission, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	m, ok := s.missions[id]
	return m, ok
}
