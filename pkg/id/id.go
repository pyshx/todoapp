package id

import (
	"github.com/google/uuid"
)

type ID[T any] struct {
	value uuid.UUID
}

func New[T any]() ID[T]                   { return ID[T]{value: uuid.New()} }
func From[T any](u uuid.UUID) ID[T]       { return ID[T]{value: u} }

func Parse[T any](s string) (ID[T], error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID[T]{}, err
	}
	return ID[T]{value: u}, nil
}

func MustParse[T any](s string) ID[T] {
	id, err := Parse[T](s)
	if err != nil {
		panic(err)
	}
	return id
}

func (id ID[T]) String() string      { return id.value.String() }
func (id ID[T]) UUID() uuid.UUID     { return id.value }
func (id ID[T]) IsZero() bool        { return id.value == uuid.Nil }
func (id ID[T]) Equal(other ID[T]) bool { return id.value == other.value }

type (
	companyIDType struct{}
	userIDType    struct{}
	taskIDType    struct{}
)

type (
	CompanyID = ID[companyIDType]
	UserID    = ID[userIDType]
	TaskID    = ID[taskIDType]
)

func NewCompanyID() CompanyID { return New[companyIDType]() }
func NewUserID() UserID       { return New[userIDType]() }
func NewTaskID() TaskID       { return New[taskIDType]() }

func ParseCompanyID(s string) (CompanyID, error) { return Parse[companyIDType](s) }
func ParseUserID(s string) (UserID, error)       { return Parse[userIDType](s) }
func ParseTaskID(s string) (TaskID, error)       { return Parse[taskIDType](s) }

func MustParseCompanyID(s string) CompanyID { return MustParse[companyIDType](s) }
func MustParseUserID(s string) UserID       { return MustParse[userIDType](s) }
func MustParseTaskID(s string) TaskID       { return MustParse[taskIDType](s) }
