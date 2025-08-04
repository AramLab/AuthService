package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/AramLab/AuthService/internal/models"
	repo "github.com/AramLab/AuthService/internal/repository/postgres"
	"github.com/AramLab/AuthService/pkg/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	log          *zap.SugaredLogger
	userSaver    UserSaver
	userProvider UserProvider
	tokenManager jwt.TokenManager
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserSaver --output=./mocks --case=underscore
type UserSaver interface {
	SaveUser(ctx context.Context, username string, passHash []byte, email, firstName, lastName string) (userID *uuid.UUID, err error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserProvider --output=./mocks --case=underscore
type UserProvider interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
}

func NewAuthService(log *zap.SugaredLogger, saver UserSaver, provider UserProvider, tokenManager jwt.TokenManager) *AuthService {
	if saver == nil {
		panic("UserSaver is nil")
	}
	if provider == nil {
		panic("UserProvider is nil")
	}
	return &AuthService{
		log:          log,
		userSaver:    saver,
		userProvider: provider,
		tokenManager: tokenManager,
	}
}

func (a *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	const op = "Serivce.Login"

	log := a.log.With("op", op, "username", username)
	log.Info("starting login flow")

	user, err := a.userProvider.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			a.log.Warnw("user not found", "op", op, "username", username)

			return "", fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}
		a.log.Errorw("failed to get user", "op", op, "username", username, "error", err)

		return "", fmt.Errorf("internal error")
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		a.log.Warnw("invalid password", "op", op, "username", username)
		return "", fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}
	token, err := a.tokenManager.GenerateToken()
	if err != nil {
		a.log.Errorw("failed to generate token", "op", op, "username", username, "error", err)
	}

	a.log.Infow("user logged in", "op", op, "username", username)

	return token, nil
}

func (a *AuthService) RegisterNewUser(ctx context.Context, username string, password string, email string, firstName string, lastName string) (*uuid.UUID, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With("op", op, "username", username)
	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Errorf("failed to generate password hash: %w", err)

		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}
	userID, err := a.userSaver.SaveUser(ctx, username, passHash, email, firstName, lastName)
	if err != nil {
		a.log.Errorw("failed to save user", "op", op, "username", username, "error", err)
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	log.Info("finish registering new user")

	return userID, nil
}
