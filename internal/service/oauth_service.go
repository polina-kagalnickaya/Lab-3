package service

import (
    "crypto/rand"
    "encoding/base64"
    "errors"
    "time"
    "log"
    
    "newyear-app/internal/models"
    "newyear-app/internal/oauth"
    "newyear-app/internal/repository"
    "newyear-app/internal/utils"
    "newyear-app/internal/dto"
)

type OAuthService struct {
    userRepo       *repository.UserRepository
    tokenRepo      *repository.TokenRepository
    authService    *AuthService
    yandexProvider oauth.Provider
    vkProvider     oauth.Provider
}

func NewOAuthService(
    userRepo *repository.UserRepository,
    tokenRepo *repository.TokenRepository,
    authService *AuthService,
    yandexProvider oauth.Provider,
    vkProvider oauth.Provider,
) *OAuthService {
    return &OAuthService{
        userRepo:       userRepo,
        tokenRepo:      tokenRepo,
        authService:    authService,
        yandexProvider: yandexProvider,
        vkProvider:     vkProvider,
    }
}

func (s *OAuthService) GenerateState() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *OAuthService) GetYandexAuthURL(state string) string {
    url := s.yandexProvider.GetAuthURL(state)
    log.Printf("Yandex auth URL: %s", url)
    return url
}

func (s *OAuthService) GetVkAuthURL(state string) string {
    return s.vkProvider.GetAuthURL(state)
}

func (s *OAuthService) HandleYandexCallback(code, state, storedState string) (accessToken, refreshToken string, user *dto.UserResponse, err error) {
    if state != storedState {
        return "", "", nil, errors.New("invalid state")
    }
    
    // Обмен кода на access token
    yandexToken, err := s.yandexProvider.ExchangeCode(code)
    if err != nil {
        return "", "", nil, err
    }
    
    // Получение информации о пользователе
    userInfo, err := s.yandexProvider.GetUserInfo(yandexToken)
    if err != nil {
        return "", "", nil, err
    }
    
    return s.handleOAuthUser(userInfo)
}

func (s *OAuthService) HandleVkCallback(code, state, storedState string) (accessToken, refreshToken string, user *dto.UserResponse, err error) {
    if state != storedState {
        return "", "", nil, errors.New("invalid state")
    }
    
    // Обмен кода на access token
    vkToken, err := s.vkProvider.ExchangeCode(code)
    if err != nil {
        return "", "", nil, err
    }
    
    // Получение информации о пользователе
    userInfo, err := s.vkProvider.GetUserInfo(vkToken)
    if err != nil {
        return "", "", nil, err
    }
    
    return s.handleOAuthUser(userInfo)
}

func (s *OAuthService) handleOAuthUser(userInfo oauth.UserInfo) (accessToken, refreshToken string, user *dto.UserResponse, err error) {
    var userModel *models.User
    
    // Поиск пользователя по ID провайдера
    switch userInfo.Provider {
    case "yandex":
        userModel, err = s.userRepo.FindByYandexID(userInfo.ID)
    case "vk":
        userModel, err = s.userRepo.FindByVkID(userInfo.ID)
    default:
        return "", "", nil, errors.New("unknown provider")
    }
    
    if err != nil {
        // Создание нового пользователя
        userModel = &models.User{
            Email:    userInfo.Email,
            FullName: userInfo.FullName,
        }
        
        switch userInfo.Provider {
        case "yandex":
            yandexID := userInfo.ID
            userModel.YandexID = &yandexID
        case "vk":
            vkID := userInfo.ID
            userModel.VkID = &vkID
        }
        
        if err := s.userRepo.Create(userModel); err != nil {
            return "", "", nil, err
        }
    }
    
    // Генерация токенов
    accessToken, err = utils.GenerateAccessToken(userModel.ID, s.authService.GetJWTSecret(), s.authService.GetJWTExp())
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
        ExpiresAt: time.Now().Add(s.authService.GetRefreshExp()),
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