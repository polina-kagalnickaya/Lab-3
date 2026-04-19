package models

import (
    "time"
)

type Token struct {
    ID        uint       `gorm:"primarykey"`
    UserID    uint       `gorm:"not null;index"`
    Hash      string     `gorm:"size:255;not null"`
    ExpiresAt time.Time  `gorm:"not null"`
    Revoked   bool       `gorm:"default:false"`
    RevokedAt *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
    
    User      User       `gorm:"foreignKey:UserID"`
}