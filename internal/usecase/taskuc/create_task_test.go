package taskuc_test

import (
	"context"
	"testing"

	"github.com/pyshx/todoapp/internal/usecase/taskuc"
	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

// mockTaskRepo is a simple mock for task.Repo
type mockTaskRepo struct {
	tasks   map[string]*task.Task
	created *task.Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*task.Task)}
}

func (m *mockTaskRepo) Create(ctx context.Context, t *task.Task) error {
	m.created = t
	m.tasks[t.ID().String()] = t
	return nil
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id id.TaskID) (*task.Task, error) {
	if t, ok := m.tasks[id.String()]; ok {
		return t, nil
	}
	return nil, apperr.NewErrNotFound("task", id.String())
}

func (m *mockTaskRepo) FindByIDForCompany(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) (*task.Task, error) {
	if t, ok := m.tasks[taskID.String()]; ok && t.CompanyID().Equal(companyID) {
		return t, nil
	}
	return nil, apperr.NewErrNotFound("task", taskID.String())
}

func (m *mockTaskRepo) ListByCompany(ctx context.Context, companyID id.CompanyID, opts task.ListOptions) (*task.ListResult, error) {
	return &task.ListResult{Tasks: nil}, nil
}

func (m *mockTaskRepo) ListByAssignee(ctx context.Context, companyID id.CompanyID, assigneeID id.UserID, opts task.ListOptions) (*task.ListResult, error) {
	return &task.ListResult{Tasks: nil}, nil
}

func (m *mockTaskRepo) Update(ctx context.Context, t *task.Task, expectedVersion int) error {
	return nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) error {
	return nil
}

// mockUserRepo is a simple mock for user.Repo
type mockUserRepo struct {
	users map[string]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*user.User)}
}

func (m *mockUserRepo) AddUser(u *user.User) {
	m.users[u.ID().String()] = u
}

func (m *mockUserRepo) FindByID(ctx context.Context, id id.UserID) (*user.User, error) {
	if u, ok := m.users[id.String()]; ok {
		return u, nil
	}
	return nil, apperr.NewErrNotFound("user", id.String())
}

func TestCreateTask_Execute(t *testing.T) {
	companyID := id.NewCompanyID()
	editorID := id.NewUserID()
	viewerID := id.NewUserID()

	editor := user.NewBuilder().
		ID(editorID).
		CompanyID(companyID).
		Email("editor@test.com").
		Role(user.RoleEditor).
		MustBuild()

	viewer := user.NewBuilder().
		ID(viewerID).
		CompanyID(companyID).
		Email("viewer@test.com").
		Role(user.RoleViewer).
		MustBuild()

	tests := []struct {
		name      string
		actor     *user.User
		input     taskuc.CreateTaskInput
		wantErr   bool
		errType   error
	}{
		{
			name:  "editor can create task",
			actor: editor,
			input: taskuc.CreateTaskInput{
				Title:      "Test Task",
				Visibility: task.VisibilityCompanyWide,
			},
			wantErr: false,
		},
		{
			name:  "viewer cannot create task",
			actor: viewer,
			input: taskuc.CreateTaskInput{
				Title:      "Test Task",
				Visibility: task.VisibilityCompanyWide,
			},
			wantErr: true,
		},
		{
			name:  "empty title fails",
			actor: editor,
			input: taskuc.CreateTaskInput{
				Title:      "",
				Visibility: task.VisibilityCompanyWide,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskRepo := newMockTaskRepo()
			userRepo := newMockUserRepo()
			userRepo.AddUser(editor)
			userRepo.AddUser(viewer)

			uc := taskuc.NewCreateTask(taskRepo, userRepo)
			result, err := uc.Execute(context.Background(), tt.actor, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.Title() != tt.input.Title {
				t.Errorf("Title = %s, want %s", result.Title(), tt.input.Title)
			}

			if !result.CompanyID().Equal(tt.actor.CompanyID()) {
				t.Error("company ID should match actor's company")
			}

			if !result.CreatorID().Equal(tt.actor.ID()) {
				t.Error("creator ID should match actor")
			}
		})
	}
}

func TestCreateTask_AssigneeValidation(t *testing.T) {
	companyID := id.NewCompanyID()
	otherCompanyID := id.NewCompanyID()
	editorID := id.NewUserID()
	validAssigneeID := id.NewUserID()
	invalidAssigneeID := id.NewUserID()

	editor := user.NewBuilder().
		ID(editorID).
		CompanyID(companyID).
		Email("editor@test.com").
		Role(user.RoleEditor).
		MustBuild()

	validAssignee := user.NewBuilder().
		ID(validAssigneeID).
		CompanyID(companyID).
		Email("assignee@test.com").
		Role(user.RoleViewer).
		MustBuild()

	invalidAssignee := user.NewBuilder().
		ID(invalidAssigneeID).
		CompanyID(otherCompanyID).
		Email("external@test.com").
		Role(user.RoleViewer).
		MustBuild()

	taskRepo := newMockTaskRepo()
	userRepo := newMockUserRepo()
	userRepo.AddUser(editor)
	userRepo.AddUser(validAssignee)
	userRepo.AddUser(invalidAssignee)

	uc := taskuc.NewCreateTask(taskRepo, userRepo)

	t.Run("valid assignee", func(t *testing.T) {
		input := taskuc.CreateTaskInput{
			Title:      "Test",
			AssigneeID: &validAssigneeID,
			Visibility: task.VisibilityCompanyWide,
		}

		_, err := uc.Execute(context.Background(), editor, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("assignee from different company", func(t *testing.T) {
		input := taskuc.CreateTaskInput{
			Title:      "Test",
			AssigneeID: &invalidAssigneeID,
			Visibility: task.VisibilityCompanyWide,
		}

		_, err := uc.Execute(context.Background(), editor, input)
		if err == nil {
			t.Error("expected error for cross-company assignee")
		}
	})

	t.Run("non-existent assignee", func(t *testing.T) {
		nonExistentID := id.NewUserID()
		input := taskuc.CreateTaskInput{
			Title:      "Test",
			AssigneeID: &nonExistentID,
			Visibility: task.VisibilityCompanyWide,
		}

		_, err := uc.Execute(context.Background(), editor, input)
		if err == nil {
			t.Error("expected error for non-existent assignee")
		}
	})
}
