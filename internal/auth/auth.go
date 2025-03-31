package auth

import "github.com/golang-jwt/jwt/v5"

type authenticator interface {
	GenerateToken(claim jwt.Claims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}
