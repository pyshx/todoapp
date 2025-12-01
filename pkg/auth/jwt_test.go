package auth

import (
	"testing"
	"time"

	"github.com/pyshx/todoapp/pkg/id"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-key-12345", 1*time.Hour)

	userID := id.NewUserID()
	companyID := id.NewCompanyID()
	role := "editor"

	token, err := svc.GenerateToken(userID, companyID, role)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if !claims.UserID.Equal(userID) {
		t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
	}
	if !claims.CompanyID.Equal(companyID) {
		t.Errorf("expected company ID %s, got %s", companyID, claims.CompanyID)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret-key-12345", -1*time.Hour)

	userID := id.NewUserID()
	companyID := id.NewCompanyID()

	token, err := svc.GenerateToken(userID, companyID, "editor")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestJWTService_InvalidSignature(t *testing.T) {
	svc1 := NewJWTService("secret-key-1", 1*time.Hour)
	svc2 := NewJWTService("secret-key-2", 1*time.Hour)

	userID := id.NewUserID()
	companyID := id.NewCompanyID()

	token, err := svc1.GenerateToken(userID, companyID, "editor")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour)

	testCases := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"no dots", "notavalidtoken"},
		{"one dot", "part1.part2"},
		{"invalid base64", "abc.def.ghi"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tc.token)
			if err == nil {
				t.Error("expected error for invalid token")
			}
		})
	}
}
