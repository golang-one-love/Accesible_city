package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

type JWTTokenGenerator struct {
	privateKeyPath string
	publicKeyPath  string
	accessTTL      time.Duration
	refreshTTL     time.Duration
	privateKey     interface{}
	publicKey      interface{}
}

func NewJWTTokenGenerator(privateKeyPath, publicKeyPath string, accessTTL, refreshTTL time.Duration) (*JWTTokenGenerator, error) {
	g := &JWTTokenGenerator{
		privateKeyPath: privateKeyPath,
		publicKeyPath:  publicKeyPath,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
	}
	if err := g.loadKeys(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *JWTTokenGenerator) loadKeys() error {
	privateKeyData, err := os.ReadFile(g.privateKeyPath)
	if err != nil {
		return fmt.Errorf("read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}
	g.privateKey = privateKey

	publicKeyData, err := os.ReadFile(g.publicKeyPath)
	if err != nil {
		return fmt.Errorf("read public key: %w", err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	g.publicKey = publicKey
	return nil
}

func (g *JWTTokenGenerator) GenerateAccessToken(user *entity.User) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email.String(),
		"role":  string(user.Role),
		"type":  "access",
		"iat":   now.Unix(),
		"exp":   now.Add(g.accessTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.privateKey)
}

func (g *JWTTokenGenerator) GenerateRefreshToken(user *entity.User) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  now.Add(g.refreshTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.privateKey)
}

func (g *JWTTokenGenerator) ValidateAccessToken(tokenString string) (*service.TokenClaims, error) {
	return g.validateToken(tokenString, "access")
}

func (g *JWTTokenGenerator) ValidateRefreshToken(tokenString string) (*service.TokenClaims, error) {
	return g.validateToken(tokenString, "refresh")
}

func (g *JWTTokenGenerator) validateToken(tokenString, expectedType string) (*service.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return g.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, service.ErrTokenExpired
		}
		return nil, service.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, service.ErrTokenInvalid
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != expectedType {
		return nil, service.ErrTokenInvalid
	}

	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	roleStr, _ := claims["role"].(string)
	exp, _ := claims["exp"].(float64)

	return &service.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   entity.Role(roleStr),
		Type:   tokenType,
		Exp:    int64(exp),
	}, nil
}