package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtHelperService struct {
	secretKey string
}

func NewJwtHelperService(secretKey string) *JwtHelperService {
	return &JwtHelperService{secretKey: secretKey}
}

// SignToken generates a JWT token with the given payload and options
func (j *JwtHelperService) SignToken(payload map[string]interface{}, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(expiration).Unix(),
	}

	// Adding the payload to claims
	for key, value := range payload {
		claims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken verifies the JWT token and returns the claims if valid
func (j *JwtHelperService) VerifyToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the token signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
