package iam

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

var (
	ErrIdToActivationRequired = connect.NewError(connect.CodeInvalidArgument, errors.New("id_to_activation required"))
	ErrDenied                 = connect.NewError(connect.CodePermissionDenied, errors.New(""))
)

func InvalidID(id string, err error) error {
	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid id: %q: %w", id, err))
}
