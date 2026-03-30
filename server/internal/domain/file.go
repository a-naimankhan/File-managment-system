package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	Filename   string    `db:"filename" json:"filename"`
	StoredName string    `db:"stored_name" json:"-"`
	Size       int64     `db:"size" json:"size"`
	MimeType   string    `db:"mime_type" json:"mime_type"`
	Checksum   string    `db:"checksum" json:"checksum"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type FileRepository interface {
	Save(ctx context.Context, file *FileMetadata) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileMetadata, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type FileService interface {
	UploadFile(ctx context.Context, file *FileMetadata) (string, error)
	DownloadFile(ctx context.Context, id uuid.UUID) (*FileMetadata, error)
}
