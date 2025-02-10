package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
)

var (
	ErrAuthHeaderMiss          = errors.New("authorization header missin")
	ErrInvalidAuthHeaderFormat = errors.New("invalid Authorization header format")
	ErrUnexpectSignMethod      = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrUserNotFoundToken       = errors.New("user data not found in token")
)

// UserClaims - custom token data
type UserClaims struct {
	User struct {
		Username string `json:"username"`
		ID       string `json:"id"`
	} `json:"user"`
	jwt.StandardClaims
}

// Generate Token - a function for generating a JWT token based on username, id, secret, where secret is the key for signing the JWT
func GenerateToken(username, id string, secret []byte) (string, error) {
	// Создаем данные токена
	claims := UserClaims{
		User: struct {
			Username string `json:"username"`
			ID       string `json:"id"`
		}{
			Username: username,
			ID:       id,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // The token expires in 7 days
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Creating a token with the HS256 algorithm and passing claims to it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signing the token with a secret key
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateID creates a 24-character token in hexadecimal format
func GenerateID() (string, error) {
	// Creating a 12-byte slice, since 12 bytes = 24 characters in hex format
	bytes := make([]byte, 12)

	// Filling bytes with random data
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert bytes to a hexadecimal string
	return hex.EncodeToString(bytes), nil
}

func getJWTByRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrAuthHeaderMiss
	}

	// Check that the title starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrInvalidAuthHeaderFormat
	}

	// Extracting the token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func getUserByJWT(tokenString string, secret []byte) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Checking the signature algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectSignMethod
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extracting user data from claims["user"]
		if userMap, ok := claims["user"].(map[string]interface{}); ok {
			user := &models.User{
				Username: fmt.Sprintf("%v", userMap["username"]),
				ID:       fmt.Sprintf("%v", userMap["id"]),
			}
			return user, nil
		}
		return nil, ErrUserNotFoundToken
	}
	return nil, ErrInvalidToken
}
