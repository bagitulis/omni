package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Token TTL constants
const (
	// AccessTokenTTL is the default access token lifetime (short-lived for security)
	AccessTokenTTL = 15 * time.Minute

	// RefreshTokenTTL is the default refresh token lifetime
	RefreshTokenTTL = 7 * 24 * time.Hour // 7 days

	// TokenIssuer identifies the token issuer
	TokenIssuer = "omni-backend"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string) *JWTService {
	return &JWTService{secretKey: []byte(secret)}
}

// GenerateToken creates a new JWT token
func (s *JWTService) GenerateToken(userID, tenantID, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates JWT token and returns claims
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateAccessToken creates a new short-lived JWT access token
func (s *JWTService) GenerateAccessToken(userID, tenantID, role string) (string, error) {
	return s.GenerateAccessTokenWithTTL(userID, tenantID, role, AccessTokenTTL)
}

// GenerateAccessTokenWithTTL creates access token with custom TTL
func (s *JWTService) GenerateAccessTokenWithTTL(userID, tenantID, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.New().String(), // jti for potential blacklisting
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken creates a cryptographically secure random refresh token (not JWT)
func (s *JWTService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken validates JWT access token with strict checks (includes issuer validation)
func (s *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(TokenIssuer),
		jwt.WithLeeway(30*time.Second),
	)

	token, err := parser.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
