package user

import (
	"time"

	"github.com/pyshx/todoapp/pkg/id"
)

type User struct {
	id        id.UserID
	companyID id.CompanyID
	email     string
	role      Role
	createdAt time.Time
}

func (u *User) ID() id.UserID         { return u.id }
func (u *User) CompanyID() id.CompanyID { return u.companyID }
func (u *User) Email() string         { return u.email }
func (u *User) Role() Role            { return u.role }
func (u *User) CreatedAt() time.Time  { return u.createdAt }
func (u *User) CanEdit() bool         { return u.role.CanEdit() }

type Builder struct {
	u   *User
	err error
}

func NewBuilder() *Builder {
	return &Builder{u: &User{}}
}

func (b *Builder) ID(id id.UserID) *Builder {
	if b.err == nil {
		b.u.id = id
	}
	return b
}

func (b *Builder) CompanyID(companyID id.CompanyID) *Builder {
	if b.err == nil {
		b.u.companyID = companyID
	}
	return b
}

func (b *Builder) Email(email string) *Builder {
	if b.err == nil {
		b.u.email = email
	}
	return b
}

func (b *Builder) Role(role Role) *Builder {
	if b.err == nil {
		b.u.role = role
	}
	return b
}

func (b *Builder) CreatedAt(t time.Time) *Builder {
	if b.err == nil {
		b.u.createdAt = t
	}
	return b
}

func (b *Builder) Build() (*User, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.u, nil
}

func (b *Builder) MustBuild() *User {
	u, err := b.Build()
	if err != nil {
		panic(err)
	}
	return u
}
