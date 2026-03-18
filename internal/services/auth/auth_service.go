package auth

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// AuthService реализует бизнес-логику авторизации.
type AuthService struct {
	log          interfaces.Logger
	authManager  AuthManager
	userRep      UserRepository
	tokenRep     TokenRepository
	sessionCache SessionCache
}

// NewAuthService создаёт сервис авторизации.
func NewAuthService(
	log interfaces.Logger,
	userRep UserRepository,
	tokenRep TokenRepository,
	manager AuthManager,
	sessionCache SessionCache,
) *AuthService {
	return &AuthService{
		log:          log,
		userRep:      userRep,
		tokenRep:     tokenRep,
		authManager:  manager,
		sessionCache: sessionCache,
	}
}

// Login выполняет аутентификацию: проверяет учётные данные, генерирует токены и кэширует сессию.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ip string) (*dto.User, *dto.TokenPair, error) {
	user, err := s.userRep.GetUserByEmail(ctx, email)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("user not found", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}

	if err != nil {
		s.log.Error("get user by email failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if !user.IsActive {
		s.log.Info("user is inactive", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusForbidden, appErr.ErrCodeForbidden, "пользователь деактивирован")
	}

	if user.PasswordHash == nil || !utils.CheckValuesHash(password, *user.PasswordHash) {
		s.log.Info("invalid credentials", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusUnauthorized, appErr.ErrInvalidCredentials, "неверный email или пароль")
	}

	jwtToken, err := s.authManager.GenerateAccessToken(
		user.ID,
		user.SchoolID,
		user.SessionVersion,
		user.Role,
	)
	if err != nil {
		s.log.Error("generate access token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации токена")
	}

	refreshToken, exp := s.authManager.GenerateRefreshToken()

	refreshTokenHash, err := utils.HashValue(refreshToken)
	if err != nil {
		s.log.Error("hash refresh token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации токена")
	}

	deviceInfo := fmt.Sprintf("%s %s", userAgent, ip)

	err = s.tokenRep.SetTokenByUser(ctx, user.ID, refreshTokenHash, deviceInfo, exp)
	if err != nil {
		s.log.Error("save refresh token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка сохранения сессии")
	}

	if err := s.sessionCache.SetSessionVersion(ctx, user.ID, user.SessionVersion, s.authManager.AccessTTL()); err != nil {
		s.log.Warn("cache session version failed", "err", err, "user_id", user.ID)
	}
	if err := s.sessionCache.SetRefreshTokenHash(ctx, user.ID, refreshTokenHash, s.authManager.RefreshTTL()); err != nil {
		s.log.Warn("cache refresh token failed", "err", err, "user_id", user.ID)
	}

	return user, &dto.TokenPair{
		AccessToken:  jwtToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Registration(ctx context.Context, user *dto.User, password string, userAgent, ip string) (*dto.User, *dto.TokenPair, error) {
	passwordHash, err := utils.HashValue(password)
	if err != nil {
		s.log.Error("hash password failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибки при хэшировании пароля")
	}

	now := time.Now()

	user = &dto.User{
		ID:             utils.GetUniqUUID(),
		Role:           dto.RoleStudent,
		ClassID:        nil,
		SchoolID:       nil,
		Email:          user.Email,
		PasswordHash:   &passwordHash,
		LastName:       user.LastName,
		FirstName:      user.FirstName,
		MiddleName:     user.MiddleName,
		AvatarKey:      nil,
		SessionVersion: 1,
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
		DeletedAt:      nil,
	}

	userDb, err := s.userRep.CreateUser(ctx, user)

	if errors.Is(err, appErr.ErrEmailAlreadyExists) {
		s.log.Info("email already exists", "email", user.Email)
		return nil, nil, appErr.NewAppError(http.StatusConflict, appErr.ErrEmailAlreadyExists, "email уже зарегистрирован")
	}

	if err != nil {
		s.log.Error("create user failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка при создании пользователя")
	}

	jwtToken, err := s.authManager.GenerateAccessToken(
		user.ID,
		user.SchoolID,
		user.SessionVersion,
		user.Role,
	)
	if err != nil {
		s.log.Error("generate access token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации токена")
	}

	refreshToken, exp := s.authManager.GenerateRefreshToken()

	refreshTokenHash, err := utils.HashValue(refreshToken)
	if err != nil {
		s.log.Error("hash refresh token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации токена")
	}

	deviceInfo := fmt.Sprintf("%s %s", userAgent, ip)

	err = s.tokenRep.SetTokenByUser(ctx, user.ID, refreshTokenHash, deviceInfo, exp)
	if err != nil {
		s.log.Error("save refresh token failed", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка сохранения сессии")
	}

	if err := s.sessionCache.SetSessionVersion(ctx, user.ID, user.SessionVersion, s.authManager.AccessTTL()); err != nil {
		s.log.Warn("cache session version failed", "err", err, "user_id", user.ID)
	}
	if err := s.sessionCache.SetRefreshTokenHash(ctx, user.ID, refreshTokenHash, s.authManager.RefreshTTL()); err != nil {
		s.log.Warn("cache refresh token failed", "err", err, "user_id", user.ID)
	}

	return userDb, &dto.TokenPair{AccessToken: jwtToken, RefreshToken: refreshToken}, nil
}
