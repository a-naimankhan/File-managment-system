package memory

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

type userRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

// can't understand y it doesn't see the UserRepo
func NewUserRepository() domain.UserRepository {
	return &userRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *userRepository) Save(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	r.users[u.ID] = u
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if u, ok := r.users[id]; ok {
		return u, nil
	}
	//somehow have to think about returning err if it nil for example
	return nil, errors.New("user not found")
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}
