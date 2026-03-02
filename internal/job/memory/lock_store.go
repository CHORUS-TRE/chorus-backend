package memory

import (
	"context"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/job"
)

type lockEntry struct {
	owner     string
	expiresAt time.Time
}

type LockStore struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

func NewLockStore() *LockStore {
	return &LockStore{
		locks: make(map[string]lockEntry),
	}
}

func (s *LockStore) TryAcquire(_ context.Context, name, owner string, timeout time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	if entry, ok := s.locks[name]; ok {
		if entry.expiresAt.After(now) {
			return false, nil
		}
	}

	s.locks[name] = lockEntry{
		owner:     owner,
		expiresAt: now.Add(timeout),
	}

	return true, nil
}

func (s *LockStore) Release(_ context.Context, name, owner string, status job.Status, msg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.locks[name]; ok && entry.owner == owner {
		delete(s.locks, name)
	}

	return nil
}

func (s *LockStore) ReleaseExpired(_ context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	var count int64

	for name, entry := range s.locks {
		if entry.expiresAt.Before(now) {
			delete(s.locks, name)
			count++
		}
	}

	return count, nil
}
