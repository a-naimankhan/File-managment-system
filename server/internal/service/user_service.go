package service

import (
	"File-management-system/server/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{
		userRepo: repo,
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
	//tut еще отдаю токен и разрешаю входить в систему
	//ну а пока временный костыль
	return user.ID.String(), nil

}
