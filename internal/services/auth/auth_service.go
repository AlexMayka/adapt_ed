package auth

import (
	"backend/internal/dto"
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
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
		s.log.Info("пользователь не найден", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}

	if err != nil {
		s.log.Error("ошибка поиска пользователя по email", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if !user.IsActive {
		s.log.Info("пользователь деактивирован", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusForbidden, appErr.ErrCodeForbidden, "пользователь деактивирован")
	}

	if user.PasswordHash == nil || !utils.CheckValuesHash(password, *user.PasswordHash) {
		s.log.Info("неверные учётные данные", "email", email)
		return nil, nil, appErr.NewAppError(http.StatusUnauthorized, appErr.ErrInvalidCredentials, "неверный email или пароль")
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.SchoolID, user.SessionVersion, user.Role, userAgent, ip)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Registration создаёт нового пользователя с ролью student и возвращает пару токенов.
func (s *AuthService) Registration(ctx context.Context, user *dto.User, password string, userAgent, ip string) (*dto.User, *dto.TokenPair, error) {
	passwordHash, err := utils.HashValue(password)
	if err != nil {
		s.log.Error("ошибка хэширования пароля", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка при хэшировании пароля")
	}

	now := time.Now()

	user = &dto.User{
		ID:             utils.GetUniqUUID(),
		Role:           dto.RoleStudent,
		Email:          user.Email,
		PasswordHash:   &passwordHash,
		LastName:       user.LastName,
		FirstName:      user.FirstName,
		MiddleName:     user.MiddleName,
		SessionVersion: 1,
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	userDB, err := s.userRep.CreateUser(ctx, user)

	if errors.Is(err, appErr.ErrEmailAlreadyExists) {
		s.log.Info("email уже зарегистрирован", "email", user.Email)
		return nil, nil, appErr.NewAppError(http.StatusConflict, appErr.ErrEmailAlreadyExists, "email уже зарегистрирован")
	}

	if err != nil {
		s.log.Error("ошибка создания пользователя", "err", err)
		return nil, nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка при создании пользователя")
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.SchoolID, user.SessionVersion, user.Role, userAgent, ip)
	if err != nil {
		return nil, nil, err
	}

	return userDB, tokens, nil
}

// issueTokens генерирует пару токенов, сохраняет refresh в БД и прогревает кэш Redis.
func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID, schoolID *uuid.UUID, sessionVersion int, role dto.UserRole, userAgent, ip string) (*dto.TokenPair, error) {
	accessToken, err := s.authManager.GenerateAccessToken(userID, schoolID, sessionVersion, role)
	if err != nil {
		s.log.Error("ошибка генерации access-токена", "err", err)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации токена")
	}

	refreshToken, exp := s.authManager.GenerateRefreshToken()

	refreshTokenHash := utils.HashSHA256(refreshToken)

	deviceInfo := fmt.Sprintf("%s %s", userAgent, ip)

	if err := s.tokenRep.SetTokenByUser(ctx, userID, refreshTokenHash, deviceInfo, exp); err != nil {
		s.log.Error("ошибка сохранения refresh-токена", "err", err)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка сохранения сессии")
	}

	if err := s.sessionCache.SetSessionVersion(ctx, userID, sessionVersion, s.authManager.AccessTTL()); err != nil {
		s.log.Warn("ошибка кэширования версии сессии", "err", err, "user_id", userID)
	}
	if err := s.sessionCache.SetRefreshTokenHash(ctx, userID, refreshTokenHash, s.authManager.RefreshTTL()); err != nil {
		s.log.Warn("ошибка кэширования refresh-токена", "err", err, "user_id", userID)
	}

	return &dto.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetMe возвращает данные пользователя по ID.
func (s *AuthService) GetMe(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	user, err := s.userRep.GetUserByID(ctx, id)

	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден", "user_id", id)
		return nil, appErr.NewAppError(http.StatusNotFound, appErr.ErrCodeNotFound, "пользователь не найден")
	}

	if err != nil {
		s.log.Error("ошибка поиска пользователя по id", "user_id", id)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if !user.IsActive {
		s.log.Info("пользователь деактивирован", "user_id", id)
		return nil, appErr.NewAppError(http.StatusForbidden, appErr.ErrCodeForbidden, "пользователь деактивирован")
	}

	return user, nil
}

// Refresh обновляет пару токенов по refresh token.
func (s *AuthService) Refresh(ctx context.Context, userID uuid.UUID, refreshToken, userAgent, ip string) (*dto.TokenPair, error) {
	user, err := s.userRep.GetUserByID(ctx, userID)
	if errors.Is(err, appErr.ErrUserNotFound) {
		s.log.Info("пользователь не найден при обновлении токена", "user_id", userID)
		return nil, appErr.NewAppError(http.StatusUnauthorized, appErr.ErrCodeUnauthenticated, "пользователь не найден")
	}
	if err != nil {
		s.log.Error("ошибка поиска пользователя", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "внутренняя ошибка сервера")
	}

	if !user.IsActive {
		s.log.Info("пользователь деактивирован", "user_id", userID)
		return nil, appErr.NewAppError(http.StatusForbidden, appErr.ErrCodeForbidden, "пользователь деактивирован")
	}

	tokenHash := utils.HashSHA256(refreshToken)

	revoked, err := s.tokenRep.RevokeTokenByUser(ctx, userID, tokenHash)
	if err != nil {
		s.log.Error("ошибка отзыва refresh-токена", "err", err, "user_id", userID)
		return nil, appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка отзыва токена")
	}

	if !revoked {
		s.log.Info("refresh-токен не найден или невалиден", "user_id", userID)
		return nil, appErr.NewAppError(http.StatusUnauthorized, appErr.ErrCodeUnauthenticated, "невалидный refresh token")
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.SchoolID, user.SessionVersion, user.Role, userAgent, ip)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Logout отзывает один refresh token текущего пользователя.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	tokenHash := utils.HashSHA256(refreshToken)

	revoked, err := s.tokenRep.RevokeTokenByUser(ctx, userID, tokenHash)
	if err != nil {
		s.log.Error("ошибка отзыва refresh-токена", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка отзыва токена")
	}

	if !revoked {
		s.log.Info("refresh-токен не найден", "user_id", userID)
		return appErr.NewAppError(http.StatusUnauthorized, appErr.ErrCodeUnauthenticated, "невалидный refresh token")
	}

	s.log.Info("refresh-токен отозван", "user_id", userID)
	return nil
}

// LogoutAll отзывает все refresh token и инкрементит версию сессии (инвалидирует все access token).
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.tokenRep.RevokeAllByUser(ctx, userID); err != nil {
		s.log.Error("ошибка отзыва всех refresh-токенов", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка отзыва токенов")
	}

	newVersion, err := s.userRep.IncrementSessionVersion(ctx, userID)
	if err != nil {
		s.log.Error("ошибка инкремента версии сессии", "err", err, "user_id", userID)
		return appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка обновления сессии")
	}

	if err := s.sessionCache.SetSessionVersion(ctx, userID, newVersion, s.authManager.AccessTTL()); err != nil {
		s.log.Warn("ошибка обновления кэша версии сессии", "err", err, "user_id", userID)
	}

	if err := s.sessionCache.DelRefreshTokenHash(ctx, userID); err != nil {
		s.log.Warn("ошибка очистки кэша refresh-токена", "err", err, "user_id", userID)
	}

	s.log.Info("все сессии пользователя отозваны", "user_id", userID, "new_version", newVersion)
	return nil
}

// RegistrationByAdmin создаёт пользователя с указанной ролью и генерирует временный пароль.
func (s *AuthService) RegistrationByAdmin(ctx context.Context, user *dto.User) (*dto.User, string, error) {
	if (user.Role == dto.RoleTeacher || user.Role == dto.RoleSchoolAdmin) && user.SchoolID == nil {
		return nil, "", appErr.NewAppError(http.StatusBadRequest, appErr.ErrCodeBadRequest, "для роли teacher и school_admin обязательна привязка к школе")
	}

	password, err := utils.GeneratePassword(16)
	if err != nil {
		s.log.Error("ошибка генерации пароля", "err", err)
		return nil, "", appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка генерации пароля")
	}

	passwordHash, err := utils.HashValue(password)
	if err != nil {
		s.log.Error("ошибка хэширования пароля", "err", err)
		return nil, "", appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка хэширования пароля")
	}

	now := time.Now()

	user = &dto.User{
		ID:             utils.GetUniqUUID(),
		Role:           user.Role,
		Email:          user.Email,
		PasswordHash:   &passwordHash,
		LastName:       user.LastName,
		FirstName:      user.FirstName,
		MiddleName:     user.MiddleName,
		ClassID:        user.ClassID,
		SchoolID:       user.SchoolID,
		SessionVersion: 1,
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	userDB, err := s.userRep.CreateUser(ctx, user)
	if errors.Is(err, appErr.ErrEmailAlreadyExists) {
		s.log.Info("email уже зарегистрирован", "email", user.Email)
		return nil, "", appErr.NewAppError(http.StatusConflict, appErr.ErrEmailAlreadyExists, "email уже зарегистрирован")
	}
	if err != nil {
		s.log.Error("ошибка создания пользователя", "err", err)
		return nil, "", appErr.NewAppError(http.StatusInternalServerError, appErr.ErrCodeInternalServer, "ошибка создания пользователя")
	}

	return userDB, password, nil
}
