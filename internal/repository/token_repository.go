package repository

import (
    "time"
    "newyear-app/internal/models"
    "gorm.io/gorm"
)

type TokenRepository struct {
    db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
    return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(token *models.Token) error {
    return r.db.Create(token).Error
}

func (r *TokenRepository) FindByHash(hash string) (*models.Token, error) {
    var token models.Token
    err := r.db.Where("hash = ? AND revoked = ?", hash, false).First(&token).Error
    return &token, err
}

func (r *TokenRepository) RevokeAllUserTokens(userID uint) error {
    now := time.Now()
    return r.db.Model(&models.Token{}).
        Where("user_id = ? AND revoked = ?", userID, false).
        Updates(map[string]interface{}{
            "revoked": true,
            "revoked_at": now,
        }).Error
}

func (r *TokenRepository) RevokeToken(tokenID uint) error {
    now := time.Now()
    return r.db.Model(&models.Token{}).
        Where("id = ?", tokenID).
        Updates(map[string]interface{}{
            "revoked": true,
            "revoked_at": now,
        }).Error
}

func (r *TokenRepository) CleanExpiredTokens() error {
    return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Token{}).Error
}