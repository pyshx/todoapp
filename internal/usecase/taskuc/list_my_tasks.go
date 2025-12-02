package taskuc

import (
	"context"

	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type ListMyTasksInput struct {
	PageSize int
	Cursor   *task.PageCursor
}

type ListMyTasksOutput struct {
	Tasks      []*task.Task
	NextCursor *task.PageCursor
}

type ListMyTasks struct {
	TaskRepo task.Repo
}

func NewListMyTasks(taskRepo task.Repo) *ListMyTasks {
	return &ListMyTasks{TaskRepo: taskRepo}
}

func (uc *ListMyTasks) Execute(ctx context.Context, actor *user.User, input ListMyTasksInput) (*ListMyTasksOutput, error) {
	result, err := uc.TaskRepo.ListByAssignee(ctx, actor.CompanyID(), actor.ID(), task.ListOptions{
		PageSize: input.PageSize,
		Cursor:   input.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &ListMyTasksOutput{
		Tasks:      result.Tasks,
		NextCursor: result.NextCursor,
	}, nil
}
