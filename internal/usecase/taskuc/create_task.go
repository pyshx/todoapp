package taskuc

import (
	"context"
	"time"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type CreateTaskInput struct {
	Title       string
	Description *string
	AssigneeID  *id.UserID
	DueDate     *time.Time
	Visibility  task.Visibility
}

type CreateTask struct {
	TaskRepo task.Repo
	UserRepo user.Repo
}

func NewCreateTask(taskRepo task.Repo, userRepo user.Repo) *CreateTask {
	return &CreateTask{
		TaskRepo: taskRepo,
		UserRepo: userRepo,
	}
}

func (uc *CreateTask) Execute(ctx context.Context, actor *user.User, input CreateTaskInput) (*task.Task, error) {
	if !actor.CanEdit() {
		return nil, apperr.NewErrPermissionDenied("create", "task", "viewer role cannot create tasks")
	}

	if input.Title == "" {
		return nil, apperr.NewErrInvalidInput("title", "cannot be empty")
	}

	if !input.Visibility.IsValid() {
		return nil, apperr.NewErrInvalidInput("visibility", "must be only_me or company_wide")
	}

	if input.AssigneeID != nil {
		assignee, err := uc.UserRepo.FindByID(ctx, *input.AssigneeID)
		if err != nil {
			if apperr.IsNotFound(err) {
				return nil, apperr.NewErrInvalidInput("assignee_id", "user not found")
			}
			return nil, err
		}
		if !assignee.CompanyID().Equal(actor.CompanyID()) {
			return nil, apperr.NewErrInvalidInput("assignee_id", "assignee must be in the same company")
		}
	}

	now := time.Now()
	t, err := task.NewBuilder().
		ID(id.NewTaskID()).
		CompanyID(actor.CompanyID()).
		CreatorID(actor.ID()).
		AssigneeID(input.AssigneeID).
		Title(input.Title).
		Description(input.Description).
		DueDate(input.DueDate).
		Visibility(input.Visibility).
		Status(task.StatusTodo).
		Version(1).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := uc.TaskRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}
