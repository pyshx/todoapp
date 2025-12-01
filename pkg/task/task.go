package task

import (
	"time"

	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/user"
)

type Task struct {
	id          id.TaskID
	companyID   id.CompanyID
	creatorID   id.UserID
	assigneeID  *id.UserID
	title       string
	description *string
	dueDate     *time.Time
	visibility  Visibility
	status      Status
	version     int
	createdAt   time.Time
	updatedAt   time.Time
}

func (t *Task) ID() id.TaskID             { return t.id }
func (t *Task) CompanyID() id.CompanyID   { return t.companyID }
func (t *Task) CreatorID() id.UserID      { return t.creatorID }
func (t *Task) AssigneeID() *id.UserID    { return t.assigneeID }
func (t *Task) Title() string             { return t.title }
func (t *Task) Description() *string      { return t.description }
func (t *Task) DueDate() *time.Time       { return t.dueDate }
func (t *Task) Visibility() Visibility    { return t.visibility }
func (t *Task) Status() Status            { return t.status }
func (t *Task) Version() int              { return t.version }
func (t *Task) CreatedAt() time.Time      { return t.createdAt }
func (t *Task) UpdatedAt() time.Time      { return t.updatedAt }

func (t *Task) CanBeViewedBy(u *user.User) bool {
	if !t.companyID.Equal(u.CompanyID()) {
		return false
	}
	if t.visibility == VisibilityCompanyWide {
		return true
	}
	if t.creatorID.Equal(u.ID()) {
		return true
	}
	if t.assigneeID != nil && t.assigneeID.Equal(u.ID()) {
		return true
	}
	return false
}

func (t *Task) BelongsToCompany(companyID id.CompanyID) bool {
	return t.companyID.Equal(companyID)
}

type Builder struct {
	t   *Task
	err error
}

func NewBuilder() *Builder {
	return &Builder{t: &Task{status: StatusTodo, version: 1}}
}

func (b *Builder) ID(id id.TaskID) *Builder {
	if b.err == nil {
		b.t.id = id
	}
	return b
}

func (b *Builder) CompanyID(companyID id.CompanyID) *Builder {
	if b.err == nil {
		b.t.companyID = companyID
	}
	return b
}

func (b *Builder) CreatorID(creatorID id.UserID) *Builder {
	if b.err == nil {
		b.t.creatorID = creatorID
	}
	return b
}

func (b *Builder) AssigneeID(assigneeID *id.UserID) *Builder {
	if b.err == nil {
		b.t.assigneeID = assigneeID
	}
	return b
}

func (b *Builder) Title(title string) *Builder {
	if b.err == nil {
		b.t.title = title
	}
	return b
}

func (b *Builder) Description(description *string) *Builder {
	if b.err == nil {
		b.t.description = description
	}
	return b
}

func (b *Builder) DueDate(dueDate *time.Time) *Builder {
	if b.err == nil {
		b.t.dueDate = dueDate
	}
	return b
}

func (b *Builder) Visibility(visibility Visibility) *Builder {
	if b.err == nil {
		b.t.visibility = visibility
	}
	return b
}

func (b *Builder) Status(status Status) *Builder {
	if b.err == nil {
		b.t.status = status
	}
	return b
}

func (b *Builder) Version(version int) *Builder {
	if b.err == nil {
		b.t.version = version
	}
	return b
}

func (b *Builder) CreatedAt(t time.Time) *Builder {
	if b.err == nil {
		b.t.createdAt = t
	}
	return b
}

func (b *Builder) UpdatedAt(t time.Time) *Builder {
	if b.err == nil {
		b.t.updatedAt = t
	}
	return b
}

func (b *Builder) Build() (*Task, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.t, nil
}

func (b *Builder) MustBuild() *Task {
	t, err := b.Build()
	if err != nil {
		panic(err)
	}
	return t
}

type Update struct {
	Title       *string
	Description **string
	AssigneeID  **id.UserID
	DueDate     **time.Time
	Visibility  *Visibility
	Status      *Status
}

func (t *Task) ApplyUpdate(u Update, now time.Time) *Task {
	newTask := *t
	newTask.version++
	newTask.updatedAt = now

	if u.Title != nil {
		newTask.title = *u.Title
	}
	if u.Description != nil {
		newTask.description = *u.Description
	}
	if u.AssigneeID != nil {
		newTask.assigneeID = *u.AssigneeID
	}
	if u.DueDate != nil {
		newTask.dueDate = *u.DueDate
	}
	if u.Visibility != nil {
		newTask.visibility = *u.Visibility
	}
	if u.Status != nil {
		newTask.status = *u.Status
	}

	return &newTask
}
