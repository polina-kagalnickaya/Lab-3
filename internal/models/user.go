package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
    Password  string         `gorm:"size:255" json:"-"`
    Salt      string         `gorm:"size:255" json:"-"`
    FullName  string         `gorm:"size:255" json:"full_name"`
    YandexID  *string        `gorm:"uniqueIndex;size:255" json:"-"`
    VkID      *string        `gorm:"uniqueIndex;size:255" json:"-"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Tokens    []Token        `gorm:"foreignKey:UserID" json:"-"`
    Wishes    []Wish         `gorm:"foreignKey:UserID" json:"-"`
}