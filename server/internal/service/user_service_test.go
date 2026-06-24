package service

import (
	"File-management-system/server/internal/domain"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	type Test struct {
		name        string
		mockSetup   func(m *MockUserRepository)
		username    string
		email       string
		password    string
		wantErr     bool
		expectedErr string
	}
	tests := []Test{
		{
			name: "Success Registration",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "alice").Return(nil, nil)
				m.On("GetByEmail", mock.Anything, "alice@test.com").Return(nil, nil)
				m.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
			username: "alice", password: "password123", email: "alice@test.com",
			wantErr: false,
		},
		{
			name:      "Invalid Email",
			mockSetup: func(m *MockUserRepository) {},
			username:  "bob", password: "password123", email: "bad-email",
			wantErr: true, expectedErr: "invalid email format",
		},
		{
			name:      "Nil user",
			mockSetup: func(m *MockUserRepository) {},
			username:  "", password: "", email: "",
			wantErr: true, expectedErr: "user is nil",
		},
		{},
		{
			name:      "Invalid Password with spaces",
			mockSetup: func(m *MockUserRepository) {},
			username:  "blabla", password: "1234567    ", email: "co@mail.ru",
			wantErr: true, expectedErr: "password too short",
		},
		{
			name: "Existing Username",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "alice").Return(&domain.User{}, nil)
			},
			username: "alice", password: "password123", email: "alice@test.com",
			wantErr: true, expectedErr: "username already exists",
		},
		{
			name: "Existing Email",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByUsername", mock.Anything, "alicee").Return(nil, nil)
				m.On("GetByEmail", mock.Anything, "alice@test.com").Return(nil, nil)
			},
			username: "alicee", password: "superpassword123", email: "alice@test.com",
			wantErr: true, expectedErr: "email already in use",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			svc := NewUserService(mockRepo, "mysecret")

			if tc.mockSetup != nil {
				tc.mockSetup(mockRepo)
			}

			user, err := svc.Register(context.Background(), tc.username, tc.password, tc.email)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			//if (err != nil) != tc.wantErr {
			//	assert.Error(t, err)
			//	assert.Contains(t, err.Error(), tc.expectedErr)
			//	//t.Errorf("Register() error = %v, wantErr %v", err, tc.wantErr)
			//} else {
			//	assert.NoError(t, err)
			//	assert.NotNil(t, user)
			//}

			mockRepo.AssertExpectations(t)
		})
	}
}

//func TestUserService_Login(t *testing.T) {
//	type Test struct {
//		name      string
//		mockSetup func(m *MockUserRepository)
//		username  string
//		password  string
//		email     string
//	}
//}
