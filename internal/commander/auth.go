package commander

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var hmacSecret = []byte("SuperSecretSharedKeyForHMAC")

type Claims struct {
	SoldierID string `json:"soldier_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a token valid for 30 seconds
func GenerateToken(soldierID string) (string, error) {
	claims := Claims{
		SoldierID: soldierID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   soldierID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(hmacSecret)
}

// VerifyToken parses and validates token, returns claims if valid
func VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return hmacSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}
