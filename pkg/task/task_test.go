package task_test

import (
	"testing"
	"time"

	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

func TestTask_CanBeViewedBy(t *testing.T) {
	companyID := id.NewCompanyID()
	otherCompanyID := id.NewCompanyID()
	creatorID := id.NewUserID()
	assigneeID := id.NewUserID()
	otherUserID := id.NewUserID()

	now := time.Now()

	// Create test users
	creator := user.NewBuilder().
		ID(creatorID).
		CompanyID(companyID).
		Email("creator@test.com").
		Role(user.RoleEditor).
		MustBuild()

	assignee := user.NewBuilder().
		ID(assigneeID).
		CompanyID(companyID).
		Email("assignee@test.com").
		Role(user.RoleViewer).
		MustBuild()

	otherUser := user.NewBuilder().
		ID(otherUserID).
		CompanyID(companyID).
		Email("other@test.com").
		Role(user.RoleViewer).
		MustBuild()

	otherCompanyUser := user.NewBuilder().
		ID(id.NewUserID()).
		CompanyID(otherCompanyID).
		Email("external@test.com").
		Role(user.RoleEditor).
		MustBuild()

	tests := []struct {
		name       string
		task       *task.Task
		user       *user.User
		wantAccess bool
	}{
		{
			name: "company-wide task visible to all company users",
			task: task.NewBuilder().
				ID(id.NewTaskID()).
				CompanyID(companyID).
				CreatorID(creatorID).
				Title("Test").
				Visibility(task.VisibilityCompanyWide).
				CreatedAt(now).
				UpdatedAt(now).
				MustBuild(),
			user:       otherUser,
			wantAccess: true,
		},
		{
			name: "only-me task visible to creator",
			task: task.NewBuilder().
				ID(id.NewTaskID()).
				CompanyID(companyID).
				CreatorID(creatorID).
				Title("Test").
				Visibility(task.VisibilityOnlyMe).
				CreatedAt(now).
				UpdatedAt(now).
				MustBuild(),
			user:       creator,
			wantAccess: true,
		},
		{
			name: "only-me task visible to assignee",
			task: task.NewBuilder().
				ID(id.NewTaskID()).
				CompanyID(companyID).
				CreatorID(creatorID).
				AssigneeID(&assigneeID).
				Title("Test").
				Visibility(task.VisibilityOnlyMe).
				CreatedAt(now).
				UpdatedAt(now).
				MustBuild(),
			user:       assignee,
			wantAccess: true,
		},
		{
			name: "only-me task not visible to other users",
			task: task.NewBuilder().
				ID(id.NewTaskID()).
				CompanyID(companyID).
				CreatorID(creatorID).
				Title("Test").
				Visibility(task.VisibilityOnlyMe).
				CreatedAt(now).
				UpdatedAt(now).
				MustBuild(),
			user:       otherUser,
			wantAccess: false,
		},
		{
			name: "task not visible to users from other companies",
			task: task.NewBuilder().
				ID(id.NewTaskID()).
				CompanyID(companyID).
				CreatorID(creatorID).
				Title("Test").
				Visibility(task.VisibilityCompanyWide).
				CreatedAt(now).
				UpdatedAt(now).
				MustBuild(),
			user:       otherCompanyUser,
			wantAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.CanBeViewedBy(tt.user)
			if got != tt.wantAccess {
				t.Errorf("CanBeViewedBy() = %v, want %v", got, tt.wantAccess)
			}
		})
	}
}

func TestTask_ApplyUpdate(t *testing.T) {
	now := time.Now()
	taskID := id.NewTaskID()
	companyID := id.NewCompanyID()
	creatorID := id.NewUserID()

	original := task.NewBuilder().
		ID(taskID).
		CompanyID(companyID).
		CreatorID(creatorID).
		Title("Original Title").
		Visibility(task.VisibilityOnlyMe).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		MustBuild()

	newTitle := "Updated Title"
	newStatus := task.StatusInProgress
	newNow := now.Add(time.Hour)

	update := task.Update{
		Title:  &newTitle,
		Status: &newStatus,
	}

	updated := original.ApplyUpdate(update, newNow)

	// Check version incremented
	if updated.Version() != 2 {
		t.Errorf("Version = %d, want 2", updated.Version())
	}

	// Check title updated
	if updated.Title() != newTitle {
		t.Errorf("Title = %s, want %s", updated.Title(), newTitle)
	}

	// Check status updated
	if updated.Status() != newStatus {
		t.Errorf("Status = %s, want %s", updated.Status(), newStatus)
	}

	// Check unchanged fields
	if updated.Visibility() != original.Visibility() {
		t.Errorf("Visibility changed unexpectedly")
	}

	// Check updated timestamp
	if !updated.UpdatedAt().Equal(newNow) {
		t.Errorf("UpdatedAt = %v, want %v", updated.UpdatedAt(), newNow)
	}

	// Original should be unchanged
	if original.Title() != "Original Title" {
		t.Errorf("Original title was modified")
	}
	if original.Version() != 1 {
		t.Errorf("Original version was modified")
	}
}

func TestVisibility_IsValid(t *testing.T) {
	tests := []struct {
		v    task.Visibility
		want bool
	}{
		{task.VisibilityOnlyMe, true},
		{task.VisibilityCompanyWide, true},
		{task.Visibility("invalid"), false},
		{task.Visibility(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.v), func(t *testing.T) {
			if got := tt.v.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		s    task.Status
		want bool
	}{
		{task.StatusTodo, true},
		{task.StatusInProgress, true},
		{task.StatusDone, true},
		{task.Status("invalid"), false},
		{task.Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.s), func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
