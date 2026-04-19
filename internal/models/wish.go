package models

import (
    "time"
    "gorm.io/gorm"
)

type Wish struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    UserID    uint           `gorm:"not null;index" json:"user_id"`
    Text      string         `gorm:"type:text;not null" json:"text"`
    Author    string         `gorm:"size:100;default:'Anonymous'" json:"author"`
    Priority  int            `gorm:"default:1;check:priority >= 1 AND priority <= 5" json:"priority"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    User      User           `gorm:"foreignKey:UserID" json:"-"`
}