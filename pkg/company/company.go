package company

import (
	"time"

	"github.com/pyshx/todoapp/pkg/id"
)

type Company struct {
	id        id.CompanyID
	name      string
	createdAt time.Time
}

func (c *Company) ID() id.CompanyID    { return c.id }
func (c *Company) Name() string        { return c.name }
func (c *Company) CreatedAt() time.Time { return c.createdAt }

type Builder struct {
	c   *Company
	err error
}

func NewBuilder() *Builder {
	return &Builder{c: &Company{}}
}

func (b *Builder) ID(id id.CompanyID) *Builder {
	if b.err == nil {
		b.c.id = id
	}
	return b
}

func (b *Builder) Name(name string) *Builder {
	if b.err == nil {
		b.c.name = name
	}
	return b
}

func (b *Builder) CreatedAt(t time.Time) *Builder {
	if b.err == nil {
		b.c.createdAt = t
	}
	return b
}

func (b *Builder) Build() (*Company, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *Builder) MustBuild() *Company {
	c, err := b.Build()
	if err != nil {
		panic(err)
	}
	return c
}
