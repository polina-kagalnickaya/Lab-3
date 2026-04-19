package repository

import (
    "newyear-app/internal/models"
    
    "gorm.io/gorm"
)

type WishRepository struct {
    db *gorm.DB
}

func NewWishRepository(db *gorm.DB) *WishRepository {
    return &WishRepository{db: db}
}

// GetDB возвращает экземпляр БД
func (r *WishRepository) GetDB() *gorm.DB {
    return r.db
}

// остальные методы остаются без изменений...
func (r *WishRepository) Create(wish *models.Wish) error {
    return r.db.Create(wish).Error
}

func (r *WishRepository) FindByID(id uint) (*models.Wish, error) {
    var wish models.Wish
    err := r.db.First(&wish, id).Error
    return &wish, err
}

func (r *WishRepository) FindAll(page, limit int, userID uint) ([]models.Wish, int64, error) {
    var wishes []models.Wish
    var total int64

    // Подсчет общего количества записей пользователя
    r.db.Model(&models.Wish{}).Where("user_id = ?", userID).Count(&total)

    offset := (page - 1) * limit
    err := r.db.Where("user_id = ?", userID).
        Offset(offset).
        Limit(limit).
        Order("priority DESC, created_at DESC").
        Find(&wishes).Error

    return wishes, total, err
}

func (r *WishRepository) FindAllPublic(page, limit int) ([]models.Wish, int64, error) {
    var wishes []models.Wish
    var total int64

    // Подсчет общего количества записей
    r.db.Model(&models.Wish{}).Count(&total)

    offset := (page - 1) * limit
    err := r.db.Offset(offset).
        Limit(limit).
        Order("priority DESC, created_at DESC").
        Find(&wishes).Error

    return wishes, total, err
}

func (r *WishRepository) Update(wish *models.Wish) error {
    return r.db.Save(wish).Error
}

func (r *WishRepository) Delete(id uint) error {
    return r.db.Delete(&models.Wish{}, id).Error
}