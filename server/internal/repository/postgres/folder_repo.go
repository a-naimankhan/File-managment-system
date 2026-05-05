package postgres

import (
	"File-management-system/server/internal/domain"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type folderRepo struct {
	db *sqlx.DB
}

func newFolderRepo(db *sqlx.DB) *folderRepo {
	return &folderRepo{db: db}
}

func (r *folderRepo) Save(ctx context.Context, folder *domain.Folder) error {
	query := `INSERT INTO folders (id , user_id , parent_id , name , created_at) VALUES($1 , $2 , $3 , $4 , %5)`
	_, err := r.db.ExecContext(ctx, query, folder.ID, folder.UserID, folder.ParentID, folder.Name, folder.CreatedAt)
	return err
}

func (r *folderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	query := `SELECT id, user_id, parent_id, name, created_at FROM folders WHERE id = $1`
	folder := &domain.Folder{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&folder.ID, &folder.UserID, &folder.ParentID, &folder.Name, &folder.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return folder, err
}

func (r *folderRepo) ListByParentID(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID) ([]*domain.Folder, error) {
	var rows *sql.Rows
	var err error

	if parentID == nil {
		query := `SELECT id, user_id, parent_id, name, created_at FROM folders WHERE user_id = $1 AND parent_id IS NULL`
		rows, err = r.db.QueryContext(ctx, query, userID)
	} else {
		query := `SELECT id, user_id, parent_id, name, created_at FROM folders WHERE user_id = $1 AND parent_id = $2`
		rows, err = r.db.QueryContext(ctx, query, userID, parentID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*domain.Folder
	for rows.Next() {
		f := &domain.Folder{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.CreatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}

	return folders, nil
}

func (r *folderRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM folders WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *folderRepo) Update(ctx context.Context, folder *domain.Folder) error {
	query := `UPDATE folders SET name = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, folder.Name, folder.ID)
	return err
}
