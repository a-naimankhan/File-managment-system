package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	ParentID  *uuid.UUID `db:"parent_id" json:"parent_id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type FolderRepository interface {
	Save(ctx context.Context, folder *Folder) error
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	ListByParentID(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]*Folder, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
	Update(ctx context.Context, folder *Folder) error
}

type FolderService interface {
	CreateFolder(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, Name string) (*Folder, error)
	DeleteFolder(ctx context.Context, userID, folderID uuid.UUID) error
	RenameFolder(ctx context.Context, userID, folderID uuid.UUID, name string) error
	ListContents(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]*Folder, []*FileMetadata, error)
}
