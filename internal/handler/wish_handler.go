package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    
    "newyear-app/internal/dto"
    "newyear-app/internal/middleware"
    "newyear-app/internal/service"
    
    "github.com/go-playground/validator/v10"
)

type WishHandler struct {
    service   *service.WishService
    validator *validator.Validate
}

func NewWishHandler(service *service.WishService) *WishHandler {
    return &WishHandler{
        service:   service,
        validator: validator.New(),
    }
}

func (h *WishHandler) InfoHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    now := time.Now()
    nextYear := now.Year() + 1
    newYear := time.Date(nextYear, time.January, 1, 0, 0, 0, 0, now.Location())
    daysLeft := int(newYear.Sub(now).Hours() / 24)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{
        "days_before_new_year": daysLeft,
    })
}

func (h *WishHandler) CreateWish(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Получаем user ID из контекста (установлен middleware)
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req dto.CreateWishRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.validator.Struct(req); err != nil {
        http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
        return
    }

    wish, err := h.service.Create(req, userID)
    if err != nil {
        http.Error(w, "Failed to create wish", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(wish)
}

func (h *WishHandler) GetAllWishes(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    response, err := h.service.GetAll(page, limit, userID)
    if err != nil {
        http.Error(w, "Failed to fetch wishes", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *WishHandler) GetWishByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := r.URL.Query().Get("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    wish, err := h.service.GetByID(uint(id), userID)
    if err != nil {
        if err.Error() == "wish not found" {
            http.Error(w, err.Error(), http.StatusNotFound)
        } else if err.Error() == "permission denied" {
            http.Error(w, err.Error(), http.StatusForbidden)
        } else {
            http.Error(w, "Failed to fetch wish", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(wish)
}

func (h *WishHandler) UpdateWish(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut && r.Method != http.MethodPatch {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := r.URL.Query().Get("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var req dto.UpdateWishRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.validator.Struct(req); err != nil {
        http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
        return
    }

    wish, err := h.service.Update(uint(id), req, userID)
    if err != nil {
        if err.Error() == "wish not found" {
            http.Error(w, err.Error(), http.StatusNotFound)
        } else if err.Error() == "permission denied" {
            http.Error(w, err.Error(), http.StatusForbidden)
        } else {
            http.Error(w, "Failed to update wish", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(wish)
}

func (h *WishHandler) DeleteWish(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := r.URL.Query().Get("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    if err := h.service.Delete(uint(id), userID); err != nil {
        if err.Error() == "wish not found" {
            http.Error(w, err.Error(), http.StatusNotFound)
        } else if err.Error() == "permission denied" {
            http.Error(w, err.Error(), http.StatusForbidden)
        } else {
            http.Error(w, "Failed to delete wish", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h *WishHandler) WishesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        if r.URL.Query().Has("id") {
            h.GetWishByID(w, r)
        } else {
            h.GetAllWishes(w, r)
        }
    case http.MethodPost:
        h.CreateWish(w, r)
    case http.MethodPut, http.MethodPatch:
        h.UpdateWish(w, r)
    case http.MethodDelete:
        h.DeleteWish(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}