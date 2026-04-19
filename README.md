# Новогодние желания (New Year Wishes App)

## Описание проекта

Веб-приложение для управления списком новогодних желаний. Пользователи могут создавать, просматривать, редактировать и удалять свои желания. Приложение поддерживает аутентификацию через JWT токены и OAuth провайдеров (Яндекс, ВКонтакте).

## Основные возможности

- Регистрация и аутентификация пользователей
- OAuth авторизация через Яндекс и ВКонтакте
- JWT токены с refresh механизмом
- CRUD операции для управления желаниями
- Мягкое удаление записей (soft delete)
- Пагинация списка желаний
- Валидация входных данных
- Автоматический подсчет дней до Нового года
- Docker контейнеризация

## Технологии

- Go 1.21
- PostgreSQL 16
- GORM (ORM)
- JWT для аутентификации
- Docker и Docker Compose
- Goose для миграций

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)

### Настройка переменных окружения

Создайте файл `.env` в корне проекта на основе примера:

```env
# База данных
DB_HOST=
DB_USER=
DB_PASSWORD=
DB_NAME=newyear_app
DB_PORT=5432

# JWT настройки
JWT_ACCESS_SECRET=your-super-secret-jwt-key-change-this
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h

# OAuth Яндекс
YANDEX_CLIENT_ID=your_yandex_client_id
YANDEX_CLIENT_SECRET=your_yandex_client_secret
YANDEX_CALLBACK_URL=http://localhost:4200/auth/oauth/yandex/callback
