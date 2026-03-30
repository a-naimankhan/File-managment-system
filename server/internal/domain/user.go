package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository interface {
	Save(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
}

type UserService interface {
	Register(ctx context.Context, username, password string) (*User, error)
	Login(ctx context.Context, username, password string) (string, error)
}
