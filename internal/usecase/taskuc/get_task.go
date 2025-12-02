package taskuc

import (
	"context"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type GetTask struct {
	TaskRepo task.Repo
}

func NewGetTask(taskRepo task.Repo) *GetTask {
	return &GetTask{TaskRepo: taskRepo}
}

func (uc *GetTask) Execute(ctx context.Context, actor *user.User, taskID id.TaskID) (*task.Task, error) {
	t, err := uc.TaskRepo.FindByIDForCompany(ctx, taskID, actor.CompanyID())
	if err != nil {
		return nil, err
	}

	if !t.CanBeViewedBy(actor) {
		return nil, apperr.NewErrPermissionDenied("view", "task", "task is not visible to you")
	}

	return t, nil
}
