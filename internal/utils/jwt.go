package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// TokenTypeAccess marks access tokens.
	TokenTypeAccess = "access"
	// TokenTypeRefresh marks refresh tokens.
	TokenTypeRefresh = "refresh"
)

// JWTClaims represents standardized auth claims for this service.
type JWTClaims struct {
	UserID    string `json:"uid"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// GenerateToken builds and signs a JWT for the given subject and token type.
func GenerateToken(secret string, userID uuid.UUID, role, tokenType string, ttl time.Duration) (string, string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()
	claims := JWTClaims{
		UserID:    userID.String(),
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, jti, expiresAt, nil
}

// ParseAndValidateToken parses and validates jwt signature, exp and token type.
func ParseAndValidateToken(secret, tokenString, expectedType string) (JWTClaims, error) {
	claims := JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil {
		return JWTClaims{}, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return JWTClaims{}, fmt.Errorf("token is invalid")
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return JWTClaims{}, fmt.Errorf("invalid token type")
	}
	if claims.UserID == "" {
		return JWTClaims{}, fmt.Errorf("missing user id")
	}
	return claims, nil
}
