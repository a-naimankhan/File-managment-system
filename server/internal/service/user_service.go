package service

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo  domain.UserRepository
	jwtSecret string
}

func NewUserService(repo domain.UserRepository, jwtS string) domain.UserService {
	return &userService{
		userRepo:  repo,
		jwtSecret: jwtS,
	}
}

func (s *userService) Register(ctx context.Context, username, password string) (*domain.User, error) {
	if len(username) < 3 {
		return nil, errors.New("username too short")
	}

	if len(password) < 8 {
		return nil, errors.New("password too short")
	}

	existing, _ := s.userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	err = s.userRepo.Save(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil

}
