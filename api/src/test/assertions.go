package test

import (
	"fmt"

	"connectrpc.com/connect"
	iamv1 "github.com/openhexes/proto/iam/v1"
	"github.com/stretchr/testify/require"
)

type Assertions struct {
	*require.Assertions
}

func (a *Assertions) ErrorCode(err error, code *connect.Code, msgAndArgs ...any) {
	if code == nil {
		a.NoError(err)
		return
	}
	if err == nil {
		a.Fail(fmt.Sprintf("expected error code %q but got nil", code), msgAndArgs...)
	}
	a.Equal(code.String(), connect.CodeOf(err).String(), msgAndArgs...)
}

func (a *Assertions) ActiveAccount(account *iamv1.Account, msgAndArgs ...any) {
	a.True(account.GetMeta().GetActive(), msgAndArgs...)
}

func (a *Assertions) InactiveAccount(account *iamv1.Account, msgAndArgs ...any) {
	a.False(account.GetMeta().GetActive(), msgAndArgs...)
}
