package dto

import "time"

type CreateWishRequest struct {
    Text     string `json:"text" validate:"required,min=3,max=500"`
    Author   string `json:"author" validate:"omitempty,min=1,max=100"`
    Priority int    `json:"priority" validate:"omitempty,min=1,max=5"`
}

type UpdateWishRequest struct {
    Text     string `json:"text" validate:"omitempty,min=3,max=500"`
    Author   string `json:"author" validate:"omitempty,min=1,max=100"`
    Priority int    `json:"priority" validate:"omitempty,min=1,max=5"`
}

type WishResponse struct {
    ID        uint      `json:"id"`
    Text      string    `json:"text"`
    Author    string    `json:"author"`
    Priority  int       `json:"priority"`
    CreatedAt time.Time `json:"created_at"`
}

type PaginatedResponse struct {
    Data []WishResponse `json:"data"`
    Meta Meta           `json:"meta"`
}

type Meta struct {
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    TotalPages int64 `json:"total_pages"`
}