package taskuc

import (
	"context"
	"time"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type UpdateTaskInput struct {
	TaskID      id.TaskID
	Version     int
	Title       *string
	Description **string
	AssigneeID  **id.UserID
	DueDate     **time.Time
	Visibility  *task.Visibility
	Status      *task.Status
}

type UpdateTask struct {
	TaskRepo task.Repo
	UserRepo user.Repo
}

func NewUpdateTask(taskRepo task.Repo, userRepo user.Repo) *UpdateTask {
	return &UpdateTask{
		TaskRepo: taskRepo,
		UserRepo: userRepo,
	}
}

func (uc *UpdateTask) Execute(ctx context.Context, actor *user.User, input UpdateTaskInput) (*task.Task, error) {
	if !actor.CanEdit() {
		return nil, apperr.NewErrPermissionDenied("update", "task", "viewer role cannot update tasks")
	}

	existingTask, err := uc.TaskRepo.FindByIDForCompany(ctx, input.TaskID, actor.CompanyID())
	if err != nil {
		return nil, err
	}

	if input.Title != nil && *input.Title == "" {
		return nil, apperr.NewErrInvalidInput("title", "cannot be empty")
	}

	if input.Visibility != nil && !input.Visibility.IsValid() {
		return nil, apperr.NewErrInvalidInput("visibility", "must be only_me or company_wide")
	}

	if input.Status != nil && !input.Status.IsValid() {
		return nil, apperr.NewErrInvalidInput("status", "must be todo, in_progress, or done")
	}

	if input.AssigneeID != nil && *input.AssigneeID != nil {
		assignee, err := uc.UserRepo.FindByID(ctx, **input.AssigneeID)
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

	update := task.Update{
		Title:       input.Title,
		Description: input.Description,
		AssigneeID:  input.AssigneeID,
		DueDate:     input.DueDate,
		Visibility:  input.Visibility,
		Status:      input.Status,
	}

	updatedTask := existingTask.ApplyUpdate(update, time.Now())

	if err := uc.TaskRepo.Update(ctx, updatedTask, input.Version); err != nil {
		return nil, err
	}

	return updatedTask, nil
}
