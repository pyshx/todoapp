package user

import (
	"context"

	"github.com/pyshx/todoapp/pkg/id"
)

type Repo interface {
	FindByID(ctx context.Context, id id.UserID) (*User, error)
}
