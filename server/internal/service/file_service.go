package service

import (
	"File-management-system/server/internal/domain"
	"File-management-system/server/internal/worker"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type FileService struct {
	fileRepo    domain.FileRepository
	userRepo    domain.UserRepository
	storagePath string
	wp          *worker.Pool
}

func NewFileService(fRepo domain.FileRepository, uRepo domain.UserRepository, path string, wP *worker.Pool) *FileService {
	return &FileService{
		fileRepo:    fRepo,
		userRepo:    uRepo,
		storagePath: path,
		wp:          wP,
	}
}

func (s *FileService) UploadFile(ctx context.Context, userID uuid.UUID, fileName string, content io.Reader) (*domain.FileMetadata, error) {
	if _, err := os.Stat(s.storagePath); os.IsNotExist(err) {
		err := os.MkdirAll(s.storagePath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create upload dir: %w", err)
		}
	}

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
		Path:       finalPath,
	}

	if err := s.fileRepo.Save(ctx, metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *FileService) DownloadFile(ctx context.Context, id uuid.UUID) (*domain.FileMetadata, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("file not found")
	}

	return file, nil

}

func (s *FileService) DeleteFile(ctx context.Context, id uuid.UUID) error {
	if err := s.fileRepo.DeleteByID(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *FileService) StartImageToPDF(ctx context.Context, fileID uuid.UUID) error {
	if s.wp == nil {
		return errors.New("worker pool is not initialized")
	}

	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}

	task := &ConvertTask{
		service:    s,
		InputPath:  file.Path,
		OutputPath: file.Path + ".pdf",
	}

	s.wp.Submit(task)

	return nil
}

func (s *FileService) ConvertImageToPDF(ctx context.Context, inputPath string, outputPath string) error {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input path does not exist: %s , err : %s", inputPath, err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir err: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.Image(inputPath, 10, 10, 190, 0, false, "", 0, "")

	return pdf.OutputFileAndClose(outputPath)

}
