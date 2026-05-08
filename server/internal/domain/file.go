package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	FolderID   *uuid.UUID `db:"folder_id" json:"folder_id"`
	Filename   string     `db:"filename" json:"filename"`
	StoredName string     `db:"stored_name" json:"-"`
	Path       string     `db:"path" json:"path"`
	Size       int64      `db:"size" json:"size"`
	MimeType   string     `db:"mime_type" json:"mime_type"`
	Checksum   string     `db:"checksum" json:"checksum"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type FileRepository interface {
	Save(ctx context.Context, file *FileMetadata) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileMetadata, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*FileMetadata, error)
}

type FileService interface {
	UploadFile(ctx context.Context, userID uuid.UUID, fileName string, folderID *uuid.UUID, content io.Reader) (*FileMetadata, error)
	DownloadFile(ctx context.Context, userId, id uuid.UUID) (*FileMetadata, error)
	DeleteFile(ctx context.Context, userId, fileId uuid.UUID) error
	StartImageToPDF(ctx context.Context, userId, id uuid.UUID) error
	ListFiles(ctx context.Context, userID uuid.UUID) ([]*FileMetadata, error)
}
