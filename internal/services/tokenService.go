package services

import (
	"os"
	"time"
	"errors"
	"crypto/rand"
	"encoding/base64"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"crypto/sha256"
    "encoding/hex"
)

var jwtSecret []byte

func init() {
	_ = godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	jwtSecret = []byte(secret)
}

func GenerateJWT(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}
func GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}
func GetUserIDFromJWT(token *jwt.Token) (uint, error) {
    claims, err := GetClaims(token)
    if err != nil {
        return 0, err
    }
    uidFloat, ok := claims["user_id"].(float64)
    if !ok {
        return 0, errors.New("user_id not found in token")
    }

    return uint(uidFloat), nil
}
func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
}

// api
func GenerateToken() (string, error) {
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        log.Printf("CRITICAL: Entropy failure: %v", err)
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func GenerateHashedToken() (string, string, error) {
    plainToken, err := GenerateToken()
    if err != nil {
        return "", "", err
    }
    hash := sha256.Sum256([]byte(plainToken))
    hashedToken := hex.EncodeToString(hash[:])

    return plainToken, hashedToken, nil
}