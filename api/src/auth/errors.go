package auth

import (
	"errors"

	"connectrpc.com/connect"
)

var (
	ErrDeactivated = connect.NewError(connect.CodePermissionDenied, errors.New("account deactivated"))
	ErrDenied      = connect.NewError(connect.CodePermissionDenied, errors.New(""))
)
