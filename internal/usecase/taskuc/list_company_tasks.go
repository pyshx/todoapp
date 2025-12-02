package taskuc

import (
	"context"

	"github.com/pyshx/todoapp/pkg/task"
	"github.com/pyshx/todoapp/pkg/user"
)

type ListCompanyTasksInput struct {
	PageSize int
	Cursor   *task.PageCursor
}

type ListCompanyTasksOutput struct {
	Tasks      []*task.Task
	NextCursor *task.PageCursor
}

type ListCompanyTasks struct {
	TaskRepo task.Repo
}

func NewListCompanyTasks(taskRepo task.Repo) *ListCompanyTasks {
	return &ListCompanyTasks{TaskRepo: taskRepo}
}

func (uc *ListCompanyTasks) Execute(ctx context.Context, actor *user.User, input ListCompanyTasksInput) (*ListCompanyTasksOutput, error) {
	result, err := uc.TaskRepo.ListByCompany(ctx, actor.CompanyID(), task.ListOptions{
		PageSize: input.PageSize,
		Cursor:   input.Cursor,
	})
	if err != nil {
		return nil, err
	}

	visibleTasks := make([]*task.Task, 0, len(result.Tasks))
	for _, t := range result.Tasks {
		if t.CanBeViewedBy(actor) {
			visibleTasks = append(visibleTasks, t)
		}
	}

	return &ListCompanyTasksOutput{
		Tasks:      visibleTasks,
		NextCursor: result.NextCursor,
	}, nil
}
