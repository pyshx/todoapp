package user_test

import (
	"testing"

	"github.com/pyshx/todoapp/pkg/user"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		role user.Role
		want bool
	}{
		{user.RoleEditor, true},
		{user.RoleViewer, true},
		{user.Role("admin"), false},
		{user.Role(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_CanEdit(t *testing.T) {
	tests := []struct {
		role user.Role
		want bool
	}{
		{user.RoleEditor, true},
		{user.RoleViewer, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.CanEdit(); got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		input    string
		wantRole user.Role
		wantOK   bool
	}{
		{"editor", user.RoleEditor, true},
		{"viewer", user.RoleViewer, true},
		{"admin", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			role, ok := user.ParseRole(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseRole() ok = %v, want %v", ok, tt.wantOK)
			}
			if role != tt.wantRole {
				t.Errorf("ParseRole() role = %v, want %v", role, tt.wantRole)
			}
		})
	}
}
