package apperr

import "fmt"

type ErrorKind string

const (
	ErrorKindValidation ErrorKind = "validation"
	ErrorKindAuth       ErrorKind = "auth"
	ErrorKindNotFound   ErrorKind = "not_found"
	ErrorKindConflict   ErrorKind = "conflict"
	ErrorKindInternal   ErrorKind = "internal"
)

type ErrNotFound struct {
	ResourceType string
	ID           string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ID)
}

func NewErrNotFound(resourceType, id string) *ErrNotFound {
	return &ErrNotFound{ResourceType: resourceType, ID: id}
}

type ErrPermissionDenied struct {
	Action   string
	Resource string
	Reason   string
}

func (e *ErrPermissionDenied) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("permission denied: cannot %s %s - %s", e.Action, e.Resource, e.Reason)
	}
	return fmt.Sprintf("permission denied: cannot %s %s", e.Action, e.Resource)
}

func NewErrPermissionDenied(action, resource, reason string) *ErrPermissionDenied {
	return &ErrPermissionDenied{Action: action, Resource: resource, Reason: reason}
}

type ErrVersionMismatch struct {
	Expected int
	Actual   int
}

func (e *ErrVersionMismatch) Error() string {
	return fmt.Sprintf("version mismatch: expected %d, got %d", e.Expected, e.Actual)
}

func NewErrVersionMismatch(expected, actual int) *ErrVersionMismatch {
	return &ErrVersionMismatch{Expected: expected, Actual: actual}
}

type ErrInvalidInput struct {
	Field  string
	Reason string
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid input: %s - %s", e.Field, e.Reason)
}

func NewErrInvalidInput(field, reason string) *ErrInvalidInput {
	return &ErrInvalidInput{Field: field, Reason: reason}
}

type ErrUnauthenticated struct {
	Reason string
}

func (e *ErrUnauthenticated) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("unauthenticated: %s", e.Reason)
	}
	return "unauthenticated"
}

func NewErrUnauthenticated(reason string) *ErrUnauthenticated {
	return &ErrUnauthenticated{Reason: reason}
}

func KindFromError(err error) ErrorKind {
	switch err.(type) {
	case *ErrNotFound:
		return ErrorKindNotFound
	case *ErrPermissionDenied:
		return ErrorKindAuth
	case *ErrVersionMismatch:
		return ErrorKindConflict
	case *ErrInvalidInput:
		return ErrorKindValidation
	case *ErrUnauthenticated:
		return ErrorKindAuth
	default:
		return ErrorKindInternal
	}
}

func IsNotFound(err error) bool             { _, ok := err.(*ErrNotFound); return ok }
func IsPermissionDenied(err error) bool     { _, ok := err.(*ErrPermissionDenied); return ok }
func IsVersionMismatch(err error) bool      { _, ok := err.(*ErrVersionMismatch); return ok }
func IsInvalidInput(err error) bool         { _, ok := err.(*ErrInvalidInput); return ok }
func IsUnauthenticated(err error) bool      { _, ok := err.(*ErrUnauthenticated); return ok }
