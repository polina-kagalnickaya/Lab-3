package service

import (
    "errors"
    "fmt"
    
    "newyear-app/internal/dto"
    "newyear-app/internal/models"
    "newyear-app/internal/repository"
    
    "gorm.io/gorm"
)

type WishService struct {
    repo *repository.WishRepository
}

func NewWishService(repo *repository.WishRepository) *WishService {
    return &WishService{repo: repo}
}

func (s *WishService) GetDB() *gorm.DB {
    return s.repo.GetDB()
}

func (s *WishService) Create(req dto.CreateWishRequest, userID uint) (*dto.WishResponse, error) {
    wish := &models.Wish{
        UserID:   userID,
        Text:     req.Text,
        Priority: req.Priority,
    }

    if req.Author != "" {
        wish.Author = req.Author
    } else {
        wish.Author = "Anonymous"
    }

    if wish.Priority == 0 {
        wish.Priority = 1
    }

    if err := s.repo.Create(wish); err != nil {
        return nil, err
    }

    return s.toResponse(wish), nil
}

func (s *WishService) GetByID(id uint, userID uint) (*dto.WishResponse, error) {
    wish, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("wish not found")
        }
        return nil, err
    }

    if wish.UserID != userID {
        return nil, fmt.Errorf("permission denied")
    }

    return s.toResponse(wish), nil
}

func (s *WishService) GetAll(page, limit int, userID uint) (*dto.PaginatedResponse, error) {
    wishes, total, err := s.repo.FindAll(page, limit, userID)
    if err != nil {
        return nil, err
    }

    responses := make([]dto.WishResponse, len(wishes))
    for i, wish := range wishes {
        responses[i] = *s.toResponse(&wish)
    }

    totalPages := (total + int64(limit) - 1) / int64(limit)

    return &dto.PaginatedResponse{
        Data: responses,
        Meta: dto.Meta{
            Total:      total,
            Page:       page,
            Limit:      limit,
            TotalPages: totalPages,
        },
    }, nil
}

func (s *WishService) Update(id uint, req dto.UpdateWishRequest, userID uint) (*dto.WishResponse, error) {
    wish, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("wish not found")
        }
        return nil, err
    }

    if wish.UserID != userID {
        return nil, fmt.Errorf("permission denied")
    }

    if req.Text != "" {
        wish.Text = req.Text
    }
    if req.Author != "" {
        wish.Author = req.Author
    }
    if req.Priority != 0 {
        wish.Priority = req.Priority
    }

    if err := s.repo.Update(wish); err != nil {
        return nil, err
    }

    return s.toResponse(wish), nil
}

func (s *WishService) Delete(id uint, userID uint) error {
    wish, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return fmt.Errorf("wish not found")
        }
        return err
    }

    if wish.UserID != userID {
        return fmt.Errorf("permission denied")
    }

    return s.repo.Delete(id)
}

func (s *WishService) toResponse(wish *models.Wish) *dto.WishResponse {
    return &dto.WishResponse{
        ID:        wish.ID,
        Text:      wish.Text,
        Author:    wish.Author,
        Priority:  wish.Priority,
        CreatedAt: wish.CreatedAt,
    }
}