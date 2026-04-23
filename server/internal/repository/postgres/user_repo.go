package postgres

import (
	"File-management-system/server/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Save(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
              VALUES (:id, :username, :email, :password_hash, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	var u domain.User
	query := `SELECT * FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &u, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	query := `SELECT * FROM users WHERE username = $1`

	err := r.db.GetContext(ctx, &u, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRowContext(ctx,
		"SELECT id, username, password_hash, email FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &user, err
}
