package memory

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"

	"sync"

	"github.com/google/uuid"
)

type fileRepository struct {
	mu    sync.RWMutex
	files map[uuid.UUID]*domain.FileMetadata
}

func NewFileRepository() domain.FileRepository {
	return &fileRepository{
		files: make(map[uuid.UUID]*domain.FileMetadata),
	}
}

func (r *fileRepository) Save(ctx context.Context, f *domain.FileMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}

	r.files[f.ID] = f
	return nil
}

func (r *fileRepository) GetUserByID(ctx context.Context, userID uuid.UUID) ([]*domain.FileMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userFiles []*domain.FileMetadata
	for _, f := range r.files {
		if f.UserID == userID {
			userFiles = append(userFiles, f)
		}
	}
	return userFiles, nil

}

func (r *fileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if f, ok := r.files[id]; ok {
		return f, nil
	}
	return nil, nil
}

func (r *fileRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.files[id] != nil {
		delete(r.files, id)
		return nil
	} else {
		return errors.New("File not found")
	}

}

func (r *fileRepository) ListByUserId(ctx context.Context, userID uuid.UUID) ([]*domain.FileMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userFiles []*domain.FileMetadata

	_, exists := r.files[userID]
	if !exists {
		return userFiles, nil
	}

	for _, f := range r.files {
		if f.UserID == userID {
			userFiles = append(userFiles, f)
		}
	}

	if len(userFiles) == 0 {
		return []*domain.FileMetadata{}, errors.New("empty list of files") //или эта хуйня должна быть в сервисе наверное хз?
	}
	return userFiles, nil
}
