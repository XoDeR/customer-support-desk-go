package jwt

import (
	"errors"
	"time"

	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
	Role   string      `json:"role"`
	jwtlib.RegisteredClaims
}

type Manager struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

func NewManager(secretKey string, accessTTL, refreshTTL time.Duration, issuer string) *Manager {
	return &Manager{
		secretKey:       secretKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		issuer:          issuer,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTokenTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTokenTTL }

func (m *Manager) GenerateAccessToken(userID uuidv7.UUID, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.accessTokenTTL)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			NotBefore: jwtlib.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); ok {
			return []byte(m.secretKey), nil
		}
		return nil, ErrInvalidToken
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}
	return claims, nil
}
