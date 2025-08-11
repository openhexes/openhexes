package e2e

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/test"
)

func TestIAMService(t *testing.T) {
	e := test.NewEnv(t)

	tokenO := e.Config.Test.Tokens.Owner
	tokenA := e.Config.Test.Tokens.Alfa

	idO := e.ResolveAccount(test.WithToken(tokenO)).Id
	accountO := e.GetAccount(idO, test.WithToken(tokenO))
	e.Require.Contains(accountO.Roles, config.RoleOwner)

	idA := e.ResolveAccount(test.WithToken(tokenA)).Id
	accountA := e.GetAccount(idA, test.WithToken(tokenO))
	e.Require.Empty(accountA.Roles)

	e.UpdateAccountActivation(map[string]bool{idA: true}, test.WithToken(tokenO))

	e.Run("list accounts", func(e *test.Env) {
		accounts := e.ListAccounts(test.WithToken(tokenO))
		e.Require.Len(accounts, 2)

		account := accounts[0]
		e.Require.Equal("alfa@test.com", account.Email)
		e.Require.ActiveAccount(account)

		account = accounts[1]
		e.Require.Equal("owner@test.com", account.Email)
		e.Require.ActiveAccount(account)
	})

	e.Run("account activation", func(e *test.Env) {
		// Try deactivating the account without having "owner" role
		e.UpdateAccountActivation(
			map[string]bool{idA: false},
			test.WithToken(tokenA),
			test.WithExpectedCode(connect.CodePermissionDenied),
		)

		// Deactivate account using owner's token
		e.UpdateAccountActivation(
			map[string]bool{idA: false},
			test.WithToken(tokenO),
		)

		// Verify the account is now inactive
		accountA = e.GetAccount(idA, test.WithToken(tokenA))
		e.Require.InactiveAccount(accountA)

		// Reactivate the account
		e.UpdateAccountActivation(
			map[string]bool{idA: true},
			test.WithToken(tokenO),
		)

		// Verify the account is now active again
		accountA = e.GetAccount(idA, test.WithToken(tokenA))
		e.Require.ActiveAccount(accountA)
	})
}
