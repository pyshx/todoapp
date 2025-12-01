package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/pyshx/todoapp/pkg/id"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

// Claims represents the JWT claims
type Claims struct {
	UserID    id.UserID    `json:"sub"`
	CompanyID id.CompanyID `json:"company_id"`
	Role      string       `json:"role"`
	IssuedAt  int64        `json:"iat"`
	ExpiresAt int64        `json:"exp"`
}

// Header represents the JWT header
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, tokenDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token for the given user
func (s *JWTService) GenerateToken(userID id.UserID, companyID id.CompanyID, role string) (string, error) {
	now := time.Now()

	header := Header{
		Algorithm: "HS256",
		Type:      "JWT",
	}

	claims := Claims{
		UserID:    userID,
		CompanyID: companyID,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.tokenDuration).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signatureInput := headerEncoded + "." + claimsEncoded
	signature := s.sign(signatureInput)

	return signatureInput + "." + signature, nil
}

// ValidateToken parses and validates a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	signatureInput := headerEncoded + "." + claimsEncoded
	expectedSignature := s.sign(signatureInput)
	if !hmac.Equal([]byte(signatureEncoded), []byte(expectedSignature)) {
		return nil, ErrInvalidSignature
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

func (s *JWTService) sign(input string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
