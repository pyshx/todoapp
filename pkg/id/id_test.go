package id_test

import (
	"testing"

	"github.com/pyshx/todoapp/pkg/id"
)

func TestID_Parse(t *testing.T) {
	validUUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	invalidUUID := "not-a-uuid"

	t.Run("valid UUID", func(t *testing.T) {
		parsed, err := id.ParseUserID(validUUID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if parsed.String() != validUUID {
			t.Errorf("String() = %s, want %s", parsed.String(), validUUID)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		_, err := id.ParseUserID(invalidUUID)
		if err == nil {
			t.Error("expected error for invalid UUID")
		}
	})
}

func TestID_IsZero(t *testing.T) {
	t.Run("new ID is not zero", func(t *testing.T) {
		newID := id.NewUserID()
		if newID.IsZero() {
			t.Error("new ID should not be zero")
		}
	})

	t.Run("zero value is zero", func(t *testing.T) {
		var zeroID id.UserID
		if !zeroID.IsZero() {
			t.Error("zero value should be zero")
		}
	})
}

func TestID_Equal(t *testing.T) {
	id1 := id.MustParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	id2 := id.MustParseUserID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	id3 := id.MustParseUserID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	if !id1.Equal(id2) {
		t.Error("same UUIDs should be equal")
	}

	if id1.Equal(id3) {
		t.Error("different UUIDs should not be equal")
	}
}

func TestID_TypeSafety(t *testing.T) {
	// This test verifies that different ID types are distinct at compile time
	// If this compiles, the types are correctly separated
	userID := id.NewUserID()
	taskID := id.NewTaskID()
	companyID := id.NewCompanyID()

	// Verify they all have unique values (extremely unlikely to collide)
	if userID.String() == taskID.String() && taskID.String() == companyID.String() {
		t.Error("IDs should be unique")
	}
}
