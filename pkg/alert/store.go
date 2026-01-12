package alert

import (
	"sync"
	"time"
)

type Store struct {
	alerts     []Alert
	mu         sync.RWMutex
	expiration time.Duration
}

func NewStore(expiration time.Duration) *Store {
	if expiration == 0 {
		expiration = 7 * 24 * time.Hour
	}
	return &Store{
		alerts:     []Alert{},
		expiration: expiration,
	}
}

func (s *Store) Add(alert Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.alerts {
		if existing.ID == alert.ID {
			return
		}
	}
	s.alerts = append(s.alerts, alert)
}

func (s *Store) MarkSeen(alertID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := range s.alerts {
		if s.alerts[i].ID == alertID {
			s.alerts[i].SeenAt = &now
			return true
		}
	}
	return false
}

func (s *Store) GetUnseen() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var unseen []Alert
	for _, a := range s.alerts {
		if a.SeenAt == nil {
			unseen = append(unseen, a)
		}
	}
	return unseen
}

func (s *Store) GetAll() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Alert, len(s.alerts))
	copy(result, s.alerts)
	return result
}

func (s *Store) Prune() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.expiration)
	var kept []Alert
	removed := 0

	for _, a := range s.alerts {
		if a.CreatedAt.After(cutoff) {
			kept = append(kept, a)
		} else {
			removed++
		}
	}

	s.alerts = kept
	return removed
}

func (s *Store) Load(alerts []Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alerts = make([]Alert, len(alerts))
	copy(s.alerts, alerts)
}

func (s *Store) Export() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Alert, len(s.alerts))
	copy(result, s.alerts)
	return result
}
