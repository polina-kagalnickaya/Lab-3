package middleware

import (
    "context"
    "net/http"
    "fmt"
    
    "newyear-app/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
    authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
    return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("access_token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        userID, err := m.authService.ValidateAccessToken(cookie.Value)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), UserIDKey, userID)
        next(w, r.WithContext(ctx))
    }
}

func GetUserIDFromContext(ctx context.Context) (uint, bool) {
    userID, ok := ctx.Value(UserIDKey).(uint)
    return userID, ok
}

// Добавьте этот метод в конец файла auth.go
func (m *AuthMiddleware) GetUserIDFromToken(tokenString string) (uint, error) {
    if m.authService == nil {
        return 0, fmt.Errorf("auth service not initialized")
    }
    return m.authService.ValidateAccessToken(tokenString)
}