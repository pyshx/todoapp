package grpc

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/pyshx/todoapp/pkg/apperr"
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	if connectErr, ok := err.(*connect.Error); ok {
		return connectErr
	}

	var notFound *apperr.ErrNotFound
	if errors.As(err, &notFound) {
		return connect.NewError(connect.CodeNotFound, notFound)
	}

	var permDenied *apperr.ErrPermissionDenied
	if errors.As(err, &permDenied) {
		return connect.NewError(connect.CodePermissionDenied, permDenied)
	}

	var versionMismatch *apperr.ErrVersionMismatch
	if errors.As(err, &versionMismatch) {
		return connect.NewError(connect.CodeAborted, versionMismatch)
	}

	var invalidInput *apperr.ErrInvalidInput
	if errors.As(err, &invalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, invalidInput)
	}

	var unauth *apperr.ErrUnauthenticated
	if errors.As(err, &unauth) {
		return connect.NewError(connect.CodeUnauthenticated, unauth)
	}

	return connect.NewError(connect.CodeInternal, err)
}
