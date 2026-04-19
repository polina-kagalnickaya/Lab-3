package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "newyear-app/internal/handler"
    "newyear-app/internal/middleware"
    "newyear-app/internal/models"
    "newyear-app/internal/oauth"
    "newyear-app/internal/repository"
    "newyear-app/internal/service"
)

func initDB() (*gorm.DB, error) {
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not loaded: %v", err)
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
        getEnv("DB_HOST", "localhost"),
        getEnv("DB_USER", "postgres"),
        getEnv("DB_PASSWORD", "postgres"),
        getEnv("DB_NAME", "newyear_app"),
        getEnv("DB_PORT", "5432"),
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Автомиграция моделей
    if err := db.AutoMigrate(
        &models.User{},
        &models.Token{},
        &models.Wish{},
    ); err != nil {
        return nil, fmt.Errorf("failed to auto migrate: %v", err)
    }

    return db, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func escapeHtml(text string) string {
    result := ""
    for _, c := range text {
        switch c {
        case '<':
            result += "&lt;"
        case '>':
            result += "&gt;"
        case '&':
            result += "&amp;"
        case '"':
            result += "&quot;"
        default:
            result += string(c)
        }
    }
    return result
}

func homePage(w http.ResponseWriter, r *http.Request, wishService *service.WishService, authService *service.AuthService) {
    // Проверяем авторизацию
    var userID uint
    var isAuthenticated bool
    
    cookie, err := r.Cookie("access_token")
    if err == nil {
        userID, err = authService.ValidateAccessToken(cookie.Value)
        if err == nil {
            isAuthenticated = true
        }
    }

    // Получаем дни до Нового года
    now := time.Now()
    nextYear := now.Year() + 1
    newYear := time.Date(nextYear, time.January, 1, 0, 0, 0, 0, now.Location())
    daysLeft := int(newYear.Sub(now).Hours() / 24)

    // Получаем список желаний
    var wishes []models.Wish
    repo := repository.NewWishRepository(wishService.GetDB())
    
    if isAuthenticated {
        wishes, _, _ = repo.FindAll(1, 100, userID)
    } else {
        wishes, _, _ = repo.FindAllPublic(1, 100)
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    // Генерируем HTML с желаниями
    wishesHTML := ""
    if len(wishes) == 0 {
        wishesHTML = `<div class="empty">✨ Пока нет желаний. Будьте первым! ✨</div>`
    } else {
        for _, wish := range wishes {
            priorityClass := "priority-" + strconv.Itoa(wish.Priority)
            priorityText := ""
            switch wish.Priority {
            case 1:
                priorityText = "⭐ Низкий"
            case 2:
                priorityText = "⭐⭐ Средний"
            case 3:
                priorityText = "⭐⭐⭐ Хороший"
            case 4:
                priorityText = "⭐⭐⭐⭐ Высокий"
            case 5:
                priorityText = "⭐⭐⭐⭐⭐ Очень важно!"
            }
            date := wish.CreatedAt.Format("02.01.2006")
            author := wish.Author
            if author == "" {
                author = "Anonymous"
            }

            wishesHTML += fmt.Sprintf(`
                <div class="wish-card" data-id="%d">
                    <div class="wish-text">%s</div>
                    <div class="wish-meta">
                        <div class="wish-author">👤 %s</div>
                        <div class="wish-priority %s">%s</div>
                        <div class="wish-date">📅 %s</div>
                    </div>
                </div>`,
                wish.ID, escapeHtml(wish.Text), escapeHtml(author), priorityClass, priorityText, date)
        }
    }

    // Форма добавления желания
    addWishForm := ""
    if isAuthenticated {
        addWishForm = `
        <div class="add-wish">
            <h3>✨ Добавить новое желание</h3>
            <form id="wishForm">
                <div class="form-group">
                    <label>Текст желания *</label>
                    <textarea name="text" required minlength="3" maxlength="500" placeholder="Напишите ваше желание..."></textarea>
                </div>
                <div class="form-group">
                    <label>Автор (опционально)</label>
                    <input type="text" name="author" maxlength="100" placeholder="Кто загадал желание?">
                </div>
                <div class="form-group">
                    <label>Приоритет</label>
                    <select name="priority">
                        <option value="1">⭐ Низкий</option>
                        <option value="2">⭐⭐ Средний</option>
                        <option value="3" selected>⭐⭐⭐ Хороший</option>
                        <option value="4">⭐⭐⭐⭐ Высокий</option>
                        <option value="5">⭐⭐⭐⭐⭐ Очень важно!</option>
                    </select>
                </div>
                <button type="submit" class="submit-btn">➕ Добавить желание</button>
            </form>
        </div>
        `
    }

    html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Новогодние желания</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: rgba(255,255,255,0.95);
            border-radius: 15px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.2);
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
        }
        .title {
            color: #764ba2;
        }
        .title h1 {
            font-size: 28px;
            margin-bottom: 5px;
        }
        .days {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 10px 20px;
            border-radius: 50px;
            font-size: 20px;
            font-weight: bold;
        }
        .logout-btn {
            background: #ff4757;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
        }
        .logout-btn:hover {
            background: #ff3838;
        }
        .add-wish {
            background: white;
            border-radius: 15px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .add-wish h3 {
            margin-bottom: 15px;
            color: #764ba2;
        }
        .form-group {
            margin-bottom: 15px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            color: #555;
            font-weight: bold;
        }
        .form-group input, .form-group textarea, .form-group select {
            width: 100%%;
            padding: 10px;
            border: 2px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
        }
        .form-group textarea {
            resize: vertical;
            min-height: 80px;
        }
        .submit-btn {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        .submit-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 10px rgba(0,0,0,0.2);
        }
        .wishes-list {
            background: white;
            border-radius: 15px;
            padding: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .wish-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 15px;
            margin-bottom: 15px;
            border-left: 5px solid #667eea;
            transition: transform 0.2s;
        }
        .wish-card:hover {
            transform: translateX(5px);
            box-shadow: 0 3px 10px rgba(0,0,0,0.1);
        }
        .wish-text {
            font-size: 18px;
            color: #333;
            margin-bottom: 10px;
        }
        .wish-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 10px;
        }
        .wish-author {
            color: #667eea;
            font-weight: bold;
        }
        .wish-priority {
            padding: 3px 10px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
        }
        .priority-5 { background: #ff4757; color: white; }
        .priority-4 { background: #ffa502; color: white; }
        .priority-3 { background: #ffd32a; color: #333; }
        .priority-2 { background: #7bed9f; color: #333; }
        .priority-1 { background: #70a1ff; color: white; }
        .wish-date {
            color: #999;
            font-size: 12px;
        }
        .empty {
            text-align: center;
            padding: 50px;
            color: #999;
            font-size: 18px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="title">
                <h1>🎄 Новогодние желания</h1>
                <p>Список желаний на Новый год</p>
            </div>
            <div class="days">
                🎅 До Нового года: %d дней
            </div>
            %s
        </div>

        %s

        <div class="wishes-list">
            <h3>📋 Список желаний</h3>
            <div id="wishesContainer">
                %s
            </div>
        </div>
    </div>

    <script>
        function logout() {
            fetch('/auth/logout', { method: 'POST' })
                .then(() => {
                    window.location.href = '/';
                });
        }

        %s
    </script>
</body>
</html>`,
        daysLeft,
        func() string {
            if isAuthenticated {
                return `<button class="logout-btn" onclick="logout()">🚪 Выйти</button>`
            }
            return `<a href="/auth/oauth/yandex" style="background: #ffd700; color: #333; text-decoration: none; padding: 10px 20px; border-radius: 5px;">🔑 Войти через Яндекс</a>`
        }(),
        addWishForm,
        wishesHTML,
        func() string {
            if isAuthenticated {
                return `
        document.getElementById('wishForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const wish = {
                text: formData.get('text'),
                author: formData.get('author') || '',
                priority: parseInt(formData.get('priority'))
            };
            
            const res = await fetch('/wishes', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(wish)
            });
            
            if (res.ok) {
                e.target.reset();
                location.reload();
            } else {
                alert('Ошибка при создании желания');
            }
        });`
            }
            return ""
        }())

    fmt.Fprint(w, html)
}

func main() {
    db, err := initDB()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Инициализация репозиториев
    userRepo := repository.NewUserRepository(db)
    tokenRepo := repository.NewTokenRepository(db)
    wishRepo := repository.NewWishRepository(db)

    // Получение конфигурации JWT
    jwtSecret := getEnv("JWT_ACCESS_SECRET", "default-secret-change-me")
    
    jwtExp, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRATION", "15m"))
    if err != nil {
        jwtExp = 15 * time.Minute
    }

    refreshExp, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h"))
    if err != nil {
        refreshExp = 168 * time.Hour
    }

    // Инициализация сервисов
    authService := service.NewAuthService(userRepo, tokenRepo, jwtSecret, jwtExp, refreshExp)

    // Настройка OAuth провайдеров
    yandexProvider := oauth.NewYandexProvider(
        getEnv("YANDEX_CLIENT_ID", ""),
        getEnv("YANDEX_CLIENT_SECRET", ""),
        getEnv("YANDEX_CALLBACK_URL", "http://localhost:4200/auth/oauth/yandex/callback"),
    )

    vkProvider := oauth.NewVkProvider(
        getEnv("VK_CLIENT_ID", ""),
        getEnv("VK_CLIENT_SECRET", ""),
        getEnv("VK_CALLBACK_URL", "http://localhost:4200/auth/oauth/vk/callback"),
    )

    oauthService := service.NewOAuthService(userRepo, tokenRepo, authService, yandexProvider, vkProvider)
    wishService := service.NewWishService(wishRepo)

    // Инициализация хендлеров
    authHandler := handler.NewAuthHandler(authService, oauthService)
    wishHandler := handler.NewWishHandler(wishService)
    authMiddleware := middleware.NewAuthMiddleware(authService)

    // ГЛАВНАЯ СТРАНИЦА
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        homePage(w, r, wishService, authService)
    })

    // ПУБЛИЧНЫЕ МАРШРУТЫ
    http.HandleFunc("/info", handler.InfoHandler)
    http.HandleFunc("/auth/register", authHandler.Register)
    http.HandleFunc("/auth/login", authHandler.Login)
    http.HandleFunc("/auth/refresh", authHandler.Refresh)
    http.HandleFunc("/auth/forgot-password", authHandler.ForgotPassword)
    http.HandleFunc("/auth/reset-password", authHandler.ResetPassword)
    http.HandleFunc("/auth/oauth/yandex", authHandler.YandexAuth)
    http.HandleFunc("/auth/oauth/yandex/callback", authHandler.YandexCallback)
    http.HandleFunc("/auth/oauth/vk", authHandler.VkAuth)
    http.HandleFunc("/auth/oauth/vk/callback", authHandler.VkCallback)

    // ЗАЩИЩЕННЫЕ МАРШРУТЫ
    http.HandleFunc("/auth/logout", authMiddleware.Authenticate(authHandler.Logout))
    http.HandleFunc("/auth/logout-all", authMiddleware.Authenticate(authHandler.LogoutAll))
    http.HandleFunc("/auth/whoami", authMiddleware.Authenticate(authHandler.Whoami))
    http.HandleFunc("/wishes", authMiddleware.Authenticate(wishHandler.WishesHandler))

    fmt.Println("Server starting on :4200...")
    fmt.Println("Open http://localhost:4200 in your browser")

    log.Fatal(http.ListenAndServe(":4200", nil))
}