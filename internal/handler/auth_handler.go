package handler

import (
    "encoding/json"
    "net/http"
    
    "newyear-app/internal/dto"
    "newyear-app/internal/middleware"
    "newyear-app/internal/service"
    
    "github.com/go-playground/validator/v10"
)

type AuthHandler struct {
    authService  *service.AuthService
    oauthService *service.OAuthService
    validator    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService, oauthService *service.OAuthService) *AuthHandler {
    return &AuthHandler{
        authService:  authService,
        oauthService: oauthService,
        validator:    validator.New(),
    }
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req dto.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    user, err := h.authService.Register(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusConflict)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req dto.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    accessToken, refreshToken, user, err := h.authService.Login(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   900,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   604800,
    })
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("refresh_token")
    if err != nil {
        http.Error(w, "Refresh token required", http.StatusUnauthorized)
        return
    }
    
    newAccessToken, newRefreshToken, err := h.authService.RefreshToken(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    newAccessToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   900,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    newRefreshToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   604800,
    })
    
    w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("refresh_token")
    if err == nil {
        h.authService.Logout(cookie.Value)
    }
    
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   -1,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   -1,
    })
    
    w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    if err := h.authService.LogoutAll(userID); err != nil {
        http.Error(w, "Failed to logout all sessions", http.StatusInternalServerError)
        return
    }
    
    // Также очищаем текущие cookies
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   -1,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   -1,
    })
    
    w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Whoami(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    user, err := h.authService.GetUserByID(userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
    var req dto.ForgotPasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := h.authService.ForgotPassword(req.Email); err != nil {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "If email exists, reset token will be sent"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    var req dto.ResetPasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := h.authService.ResetPassword(req.Token, req.Password); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func (h *AuthHandler) YandexAuth(w http.ResponseWriter, r *http.Request) {
    state, err := h.oauthService.GenerateState()
    if err != nil {
        http.Error(w, "Failed to generate state", http.StatusInternalServerError)
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state_yandex",
        Value:    state,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/auth/oauth/yandex/callback",
        MaxAge:   300,
    })
    
    authURL := h.oauthService.GetYandexAuthURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) YandexCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    cookie, err := r.Cookie("oauth_state_yandex")
    if err != nil {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    accessToken, refreshToken, user, err := h.oauthService.HandleYandexCallback(code, state, cookie.Value)
    if err != nil {
        http.Error(w, "OAuth failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    _ = user
    
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   900,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   604800,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state_yandex",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/auth/oauth/yandex/callback",
        MaxAge:   -1,
    })
    
    http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) VkAuth(w http.ResponseWriter, r *http.Request) {
    state, err := h.oauthService.GenerateState()
    if err != nil {
        http.Error(w, "Failed to generate state", http.StatusInternalServerError)
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state_vk",
        Value:    state,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/auth/oauth/vk/callback",
        MaxAge:   300,
    })
    
    authURL := h.oauthService.GetVkAuthURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) VkCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    cookie, err := r.Cookie("oauth_state_vk")
    if err != nil {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    accessToken, refreshToken, user, err := h.oauthService.HandleVkCallback(code, state, cookie.Value)
    if err != nil {
        http.Error(w, "OAuth failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    _ = user
    
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
        MaxAge:   900,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/auth/refresh",
        MaxAge:   604800,
    })
    
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state_vk",
        Value:    "",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/auth/oauth/vk/callback",
        MaxAge:   -1,
    })
    
    http.Redirect(w, r, "/", http.StatusFound)
}