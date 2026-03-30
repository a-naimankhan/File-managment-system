package service

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type fileService struct {
	fileRepo    domain.FileRepository
	userRepo    domain.UserRepository
	storagePath string
}

func NewFileService(fRepo domain.FileRepository, uRepo domain.UserRepository, path string) *fileService {
	return &fileService{
		fileRepo:    fRepo,
		userRepo:    uRepo,
		storagePath: path,
	}
}

func (s *fileService) UploadFile(ctx context.Context, userID uuid.UUID, fileName string, content io.Reader) (*domain.FileMetadata, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	ext := filepath.Ext(fileName)
	storedName := uuid.New().String() + ext
	finalPath := filepath.Join(s.storagePath, storedName)

	dst, err := os.Create(finalPath)
	if err != nil {
		return nil, err
	}

	defer dst.Close()

	size, err := io.Copy(dst, content)
	if err != nil {
		return nil, err
	}

	metadata := &domain.FileMetadata{
		ID:         uuid.New(),
		UserID:     userID,
		Filename:   fileName,
		StoredName: storedName,
		Size:       size,
	}

	if err := s.fileRepo.Save(ctx, metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *fileService) DownloadFile(ctx context.Context, id uuid.UUID) (*domain.FileMetadata, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("file not found")
	}

	return file, nil

}
