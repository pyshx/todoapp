package taskuc

import (
	"context"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type DeleteTask struct {
	TaskRepo task.Repo
}

func NewDeleteTask(taskRepo task.Repo) *DeleteTask {
	return &DeleteTask{TaskRepo: taskRepo}
}

func (uc *DeleteTask) Execute(ctx context.Context, actor *user.User, taskID id.TaskID) error {
	if !actor.CanEdit() {
		return apperr.NewErrPermissionDenied("delete", "task", "viewer role cannot delete tasks")
	}

	return uc.TaskRepo.Delete(ctx, taskID, actor.CompanyID())
}
