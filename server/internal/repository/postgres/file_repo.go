package postgres

import (
	"File-management-system/server/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type fileRepo struct {
	db *sqlx.DB
}

func NewFileRepo(db *sqlx.DB) *fileRepo {
	return &fileRepo{db: db}
}

func (r *fileRepo) Save(ctx context.Context, meta *domain.FileMetadata) error {
	query := `INSERT INTO file_metadata (id, user_id, filename, stored_name, path, size, mime_type, checksum, created_at) 
              VALUES (:id, :user_id, :filename, :stored_name, :path, :size, :mime_type, :checksum, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, meta)
	return err
}

func (r *fileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileMetadata, error) {
	var f domain.FileMetadata
	query := `SELECT * FROM file_metadata WHERE id = $1`

	err := r.db.GetContext(ctx, &f, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	return &f, nil
}

func (r *fileRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM file_metadata WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
