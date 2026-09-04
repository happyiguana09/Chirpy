package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	secretKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString(secretKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	validToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := validToken.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(userID)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("no auth header included in request")
	}

	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	token_string := splitAuth[1]

	return token_string, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("no auth header included in request")
	}

	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	apiKey := splitAuth[1]

	return apiKey, nil
}
