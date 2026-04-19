package service

import (
    "errors"
    "time"
    
    "newyear-app/internal/dto"
    "newyear-app/internal/models"
    "newyear-app/internal/repository"
    "newyear-app/internal/utils"
    
    "gorm.io/gorm"
)

type AuthService struct {
    userRepo   *repository.UserRepository
    tokenRepo  *repository.TokenRepository
    jwtSecret  string
    jwtExp     time.Duration
    refreshExp time.Duration
}

func NewAuthService(
    userRepo *repository.UserRepository,
    tokenRepo *repository.TokenRepository,
    jwtSecret string,
    jwtExp time.Duration,
    refreshExp time.Duration,
) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        tokenRepo:  tokenRepo,
        jwtSecret:  jwtSecret,
        jwtExp:     jwtExp,
        refreshExp: refreshExp,
    }
}

// Геттеры для доступа из OAuthService
func (s *AuthService) GetJWTSecret() string {
    return s.jwtSecret
}

func (s *AuthService) GetJWTExp() time.Duration {
    return s.jwtExp
}

func (s *AuthService) GetRefreshExp() time.Duration {
    return s.refreshExp
}

func (s *AuthService) GetUserRepo() *repository.UserRepository {
    return s.userRepo
}

func (s *AuthService) GetTokenRepo() *repository.TokenRepository {
    return s.tokenRepo
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
    // Проверка существования пользователя
    _, err := s.userRepo.FindByEmail(req.Email)
    if err == nil {
        return nil, errors.New("user already exists")
    }
    
    // Генерация соли и хеша пароля
    salt, err := utils.GenerateSalt()
    if err != nil {
        return nil, err
    }
    
    hash, err := utils.HashPassword(req.Password, salt)
    if err != nil {
        return nil, err
    }
    
    user := &models.User{
        Email:    req.Email,
        Password: hash,
        Salt:     salt,
        FullName: req.FullName,
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    return &dto.UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        FullName:  user.FullName,
        CreatedAt: user.CreatedAt,
    }, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (accessToken, refreshToken string, user *dto.UserResponse, err error) {
    // Поиск пользователя
    userModel, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "", "", nil, errors.New("invalid credentials")
        }
        return "", "", nil, err
    }
    
    // Проверка пароля
    if !utils.CheckPassword(req.Password, userModel.Salt, userModel.Password) {
        return "", "", nil, errors.New("invalid credentials")
    }
    
    // Генерация токенов
    accessToken, err = utils.GenerateAccessToken(userModel.ID, s.jwtSecret, s.jwtExp)
    if err != nil {
        return "", "", nil, err
    }
    
    refreshToken, err = utils.GenerateRefreshToken()
    if err != nil {
        return "", "", nil, err
    }
    
    // Сохранение refresh token
    tokenHash := utils.HashToken(refreshToken)
    token := &models.Token{
        UserID:    userModel.ID,
        Hash:      tokenHash,
        ExpiresAt: time.Now().Add(s.refreshExp),
    }
    
    if err := s.tokenRepo.Create(token); err != nil {
        return "", "", nil, err
    }
    
    user = &dto.UserResponse{
        ID:        userModel.ID,
        Email:     userModel.Email,
        FullName:  userModel.FullName,
        CreatedAt: userModel.CreatedAt,
    }
    
    return accessToken, refreshToken, user, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
    // Хеширование полученного токена
    tokenHash := utils.HashToken(refreshToken)
    
    // Поиск токена в БД
    token, err := s.tokenRepo.FindByHash(tokenHash)
    if err != nil {
        return "", "", errors.New("invalid refresh token")
    }
    
    // Проверка срока действия
    if time.Now().After(token.ExpiresAt) {
        return "", "", errors.New("refresh token expired")
    }
    
    // Генерация новых токенов
    newAccessToken, err = utils.GenerateAccessToken(token.UserID, s.jwtSecret, s.jwtExp)
    if err != nil {
        return "", "", err
    }
    
    newRefreshToken, err = utils.GenerateRefreshToken()
    if err != nil {
        return "", "", err
    }
    
    // Отзыв старого токена
    s.tokenRepo.RevokeToken(token.ID)
    
    // Сохранение нового refresh token
    newTokenHash := utils.HashToken(newRefreshToken)
    newToken := &models.Token{
        UserID:    token.UserID,
        Hash:      newTokenHash,
        ExpiresAt: time.Now().Add(s.refreshExp),
    }
    
    if err := s.tokenRepo.Create(newToken); err != nil {
        return "", "", err
    }
    
    return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(refreshToken string) error {
    tokenHash := utils.HashToken(refreshToken)
    token, err := s.tokenRepo.FindByHash(tokenHash)
    if err != nil {
        return err
    }
    
    return s.tokenRepo.RevokeToken(token.ID)
}

func (s *AuthService) LogoutAll(userID uint) error {
    return s.tokenRepo.RevokeAllUserTokens(userID)
}

func (s *AuthService) GetUserByID(userID uint) (*dto.UserResponse, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, err
    }
    
    return &dto.UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        FullName:  user.FullName,
        CreatedAt: user.CreatedAt,
    }, nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (uint, error) {
    claims, err := utils.ValidateAccessToken(tokenString, s.jwtSecret)
    if err != nil {
        return 0, err
    }
    return claims.UserID, nil
}

// ForgotPassword - генерация токена сброса пароля (упрощенная версия)
func (s *AuthService) ForgotPassword(email string) error {
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        // Не раскрываем информацию о существовании пользователя
        return nil
    }
    
    // Генерация токена сброса (в реальном проекте отправляем на email)
    resetToken, err := utils.GenerateRefreshToken()
    if err != nil {
        return err
    }
    
    // Сохраняем токен сброса в специальную таблицу или поле
    // Для простоты - сохраняем как обычный токен с очень коротким сроком
    tokenHash := utils.HashToken(resetToken)
    token := &models.Token{
        UserID:    user.ID,
        Hash:      tokenHash,
        ExpiresAt: time.Now().Add(15 * time.Minute),
    }
    
    return s.tokenRepo.Create(token)
}

// ResetPassword - сброс пароля
func (s *AuthService) ResetPassword(token, newPassword string) error {
    tokenHash := utils.HashToken(token)
    storedToken, err := s.tokenRepo.FindByHash(tokenHash)
    if err != nil {
        return errors.New("invalid reset token")
    }
    
    if time.Now().After(storedToken.ExpiresAt) {
        return errors.New("reset token expired")
    }
    
    // Генерация новой соли и хеша
    salt, err := utils.GenerateSalt()
    if err != nil {
        return err
    }
    
    hash, err := utils.HashPassword(newPassword, salt)
    if err != nil {
        return err
    }
    
    user, err := s.userRepo.FindByID(storedToken.UserID)
    if err != nil {
        return err
    }
    
    user.Password = hash
    user.Salt = salt
    
    if err := s.userRepo.Update(user); err != nil {
        return err
    }
    
    // Отзыв всех токенов после сброса пароля
    return s.tokenRepo.RevokeAllUserTokens(user.ID)
}