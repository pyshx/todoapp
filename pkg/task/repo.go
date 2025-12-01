package task

import (
	"context"
	"time"

	"github.com/pyshx/todoapp/pkg/id"
)

type PageCursor struct {
	CreatedAt time.Time
	ID        id.TaskID
}

type ListOptions struct {
	PageSize int
	Cursor   *PageCursor
}

type ListResult struct {
	Tasks      []*Task
	NextCursor *PageCursor
}

type Repo interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id id.TaskID) (*Task, error)
	FindByIDForCompany(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) (*Task, error)
	ListByCompany(ctx context.Context, companyID id.CompanyID, opts ListOptions) (*ListResult, error)
	ListByAssignee(ctx context.Context, companyID id.CompanyID, assigneeID id.UserID, opts ListOptions) (*ListResult, error)
	Update(ctx context.Context, task *Task, expectedVersion int) error
	Delete(ctx context.Context, taskID id.TaskID, companyID id.CompanyID) error
}
