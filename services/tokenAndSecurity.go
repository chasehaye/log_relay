package services

import (
    "crypto/rand"
    "encoding/base64"
    "log"

    "golang.org/x/crypto/bcrypt"
)

func GenerateToken() (string, error) {
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        log.Printf("CRITICAL: Entropy failure: %v", err)
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func HashPassword(password string) (string, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    if err != nil {
        return "", err
    }
    return string(hashed), nil
}

func ComparePassword(hashedPassword, plainPassword string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}