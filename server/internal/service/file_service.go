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

func NewFileService(fRepo domain.FileRepository, uRepo domain.UserRepository, path string, workerPool *worker.Pool) *FileService {
	return &FileService{
		fileRepo:    fRepo,
		userRepo:    uRepo,
		storagePath: path,
		wp:          workerPool,
	}
}

// нужно ли сделать так что бы один парсил допустим данные и отдавал в функцию ниже ведь в Интерфейсе там именно отдается файл
// or I have to change the interface arguments or I have to just parse it .
// Или парсинг что бы отдавать нужные аргументы идет в другом месте
// ну тут я мог бы изменить на просто файл и через поля структур достигать тоже самое но тут сложность именно с контетом идет

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

func (s *FileService) EnqueuePDFConvertation(ctx context.Context, id uuid.UUID) error {
	meta, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	input := meta.Path
	output := input + "pdf"

	task := &ConvertTask{
		service:    s,
		InputPath:  input,
		OutputPath: output,
	}
	s.wp.Submit(task)

	return nil
}

func (s *FileService) ConvertImageToPDF(inputPath string, outputPath string) error {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input path does not exist: %s , err : %s", inputPath, err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.Image(inputPath, 10, 10, 190, 0, false, "", 0, "")

	err := pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return fmt.Errorf("output file error: %w", err)
	}

	return nil

}
