package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
	"github.com/wardayadev/ub-mager-api/internal/service/testutil"
)

func newTestAuthService() (*AuthService, *testutil.MockUserRepo) {
	userRepo := new(testutil.MockUserRepo)
	jwtService := jwtpkg.NewJWTService("test-secret-key-32chars-minimum!", 15, 10080)
	svc := NewAuthService(userRepo, jwtService)
	return svc, userRepo
}

func TestRegister_Success(t *testing.T) {
	svc, userRepo := newTestAuthService()
	ctx := context.Background()

	userRepo.On("FindByPhone", ctx, "081234567890").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("FindByEmail", ctx, "test@student.ub.ac.id").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)

	input := RegisterInput{
		Phone:    "081234567890",
		Email:    "test@student.ub.ac.id",
		Password: "password123",
		FullName: "Test User",
		Role:     "passenger",
	}

	resp, refreshToken, err := svc.Register(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, "test@student.ub.ac.id", resp.User.Email)
	assert.True(t, resp.User.IsStudentVerified)
	assert.Equal(t, "Bearer", resp.TokenType)
	userRepo.AssertExpectations(t)
}

func TestRegister_InvalidCampusEmail(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	input := RegisterInput{
		Phone:    "081234567890",
		Email:    "test@gmail.com",
		Password: "password123",
		FullName: "Test User",
		Role:     "passenger",
	}

	resp, _, err := svc.Register(ctx, input)

	assert.ErrorIs(t, err, ErrInvalidCampusEmail)
	assert.Nil(t, resp)
}

func TestRegister_PhoneAlreadyExists(t *testing.T) {
	svc, userRepo := newTestAuthService()
	ctx := context.Background()

	existingUser := &model.User{Phone: "081234567890"}
	userRepo.On("FindByPhone", ctx, "081234567890").Return(existingUser, nil)

	input := RegisterInput{
		Phone:    "081234567890",
		Email:    "test@student.ub.ac.id",
		Password: "password123",
		FullName: "Test User",
		Role:     "passenger",
	}

	resp, _, err := svc.Register(ctx, input)

	assert.ErrorIs(t, err, ErrPhoneAlreadyExists)
	assert.Nil(t, resp)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	svc, userRepo := newTestAuthService()
	ctx := context.Background()

	userRepo.On("FindByPhone", ctx, "081234567890").Return(nil, gorm.ErrRecordNotFound)
	existingUser := &model.User{Email: "test@student.ub.ac.id"}
	userRepo.On("FindByEmail", ctx, "test@student.ub.ac.id").Return(existingUser, nil)

	input := RegisterInput{
		Phone:    "081234567890",
		Email:    "test@student.ub.ac.id",
		Password: "password123",
		FullName: "Test User",
		Role:     "passenger",
	}

	resp, _, err := svc.Register(ctx, input)

	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
	assert.Nil(t, resp)
}

func TestIsCampusEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"user@student.ub.ac.id", true},
		{"USER@STUDENT.UB.AC.ID", true},
		{"test@Student.UB.AC.ID", true},
		{"user@gmail.com", false},
		{"user@ub.ac.id", false},
		{"user@student.ub.ac", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			assert.Equal(t, tt.expected, isCampusEmail(tt.email))
		})
	}
}
