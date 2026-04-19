package utils

import (
    "crypto/rand"
    "encoding/base64"
    "golang.org/x/crypto/bcrypt"
)

// Генерация уникальной соли
func GenerateSalt() (string, error) {
    salt := make([]byte, 16)  // 16 байт = 128 бит, в base64 будет ~24 символа
    _, err := rand.Read(salt)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(salt), nil
}

// Хеширование пароля с солью
func HashPassword(password, salt string) (string, error) {
    saltedPassword := password + salt
    hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// Проверка пароля
func CheckPassword(password, salt, hash string) bool {
    saltedPassword := password + salt
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
    return err == nil
}
