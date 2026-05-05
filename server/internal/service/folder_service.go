package service

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type FolderService struct {
	folderRepo domain.FolderRepository
	fileRepo   domain.FileRepository
}

func NewFolderService(fRepo domain.FolderRepository, fileRepo domain.FileRepository) *FolderService {
	return &FolderService{
		folderRepo: fRepo,
		fileRepo:   fileRepo,
	}
}

func (s *FolderService) CreateFolder(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, name string) (*domain.Folder, error) {
	if len(name) == 0 {
		return nil, errors.New("folder name cannot be empty")
	}

	if parentID != nil {
		parent, err := s.folderRepo.GetByID(ctx, *parentID)
		if err != nil || parent == nil {
			return nil, errors.New("parent folder not found")
		}
		if parent.UserID != userID {
			return nil, errors.New("access denied")
		}
	}

	folder := &domain.Folder{
		ID:        uuid.New(),
		UserID:    userID,
		ParentID:  parentID,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.folderRepo.Save(ctx, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) error {
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil || folder == nil {
		return errors.New("folder not found")
	}

	if folderID != userID {
		return errors.New("access denied")
	}

	return s.folderRepo.DeleteByID(ctx, folderID)
}

func (s *FolderService) RenameFolder(ctx context.Context, userID, folderID uuid.UUID, name string) error {
	if len(name) == 0 {
		return errors.New("folder name cannot be empty")
	}

	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil || folder == nil {
		return errors.New("folder not found")
	}

	if folder.UserID != userID {
		return errors.New("access denied")
	}

	folder.Name = name
	return s.folderRepo.Update(ctx, folder)
}

func (s *FolderService) ListContents(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]*domain.Folder, []*domain.FileMetadata, error) {
	folders, err := s.folderRepo.ListByParentID(ctx, userID, parentID)
	if err != nil {
		return nil, nil, err
	}

	files, err := s.fileRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var filtered []*domain.FileMetadata
	for _, f := range files {
		if parentID == nil && f.FolderID == nil {
			filtered = append(filtered, f)
		} else if parentID != nil && f.FolderID != nil && *f.FolderID == *parentID {
			filtered = append(filtered, f)
		}
	}

	return folders, filtered, nil
}
