package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPhoneAlreadyExists = errors.New("phone number already registered")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCampusEmail = errors.New("email must be a valid @student.ub.ac.id address")
)

const campusEmailDomain = "@student.ub.ac.id"

type AuthService struct {
	userRepo   UserRepo
	jwtService *jwtpkg.JWTService
}

func NewAuthService(userRepo UserRepo, jwtService *jwtpkg.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type RegisterInput struct {
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Role     string `json:"role" binding:"required,oneof=passenger driver"`
}

type LoginInput struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User        *UserResponse `json:"user"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   int           `json:"expires_in"`
}

type UserResponse struct {
	ID                uuid.UUID `json:"id"`
	Phone             string    `json:"phone"`
	Email             string    `json:"email"`
	FullName          string    `json:"full_name"`
	Role              string    `json:"role"`
	AvatarURL         *string   `json:"avatar_url"`
	IsActive          bool      `json:"is_active"`
	IsStudentVerified bool      `json:"is_student_verified"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toUserResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:                user.ID,
		Phone:             user.Phone,
		Email:             user.Email,
		FullName:          user.FullName,
		Role:              string(user.Role),
		AvatarURL:         user.AvatarURL,
		IsActive:          user.IsActive,
		IsStudentVerified: user.IsStudentVerified,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

func isCampusEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), campusEmailDomain)
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResponse, string, error) {
	if !isCampusEmail(input.Email) {
		return nil, "", ErrInvalidCampusEmail
	}

	existing, err := s.userRepo.FindByPhone(ctx, input.Phone)
	if err == nil && existing != nil {
		return nil, "", ErrPhoneAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	existing, err = s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, "", ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	role := parseUserRole(input.Role)
	user := &model.User{
		ID:                uuid.New(),
		FullName:          input.FullName,
		Phone:             input.Phone,
		Email:             input.Email,
		PasswordHash:      string(hashedPassword),
		Role:              role,
		IsActive:          true,
		IsStudentVerified: isCampusEmail(input.Email),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	return &AuthResponse{
		User:        toUserResponse(user),
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtService.GetAccessTTL().Seconds()),
	}, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResponse, string, error) {
	user, err := s.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	return &AuthResponse{
		User:        toUserResponse(user),
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtService.GetAccessTTL().Seconds()),
	}, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthResponse, string, error) {
	claims, err := s.jwtService.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	return &AuthResponse{
		User:        toUserResponse(user),
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtService.GetAccessTTL().Seconds()),
	}, newRefreshToken, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return toUserResponse(user), nil
}

type UpdateProfileInput struct {
	FullName  *string `json:"full_name" binding:"omitempty,min=2,max=100"`
	Email     *string `json:"email" binding:"omitempty,email"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url"`
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if input.FullName != nil {
		user.FullName = *input.FullName
	}
	if input.Email != nil {
		// Check email uniqueness
		existing, err := s.userRepo.FindByEmail(ctx, *input.Email)
		if err == nil && existing != nil && existing.ID != userID {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = *input.Email
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

func parseUserRole(s string) model.UserRole {
	switch s {
	case "passenger":
		return model.RolePassenger
	case "driver":
		return model.RoleDriver
	case "admin":
		return model.RoleAdmin
	default:
		return model.RolePassenger
	}
}
