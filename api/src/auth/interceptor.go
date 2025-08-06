package auth

import (
	"context"

	"connectrpc.com/connect"
)

func (c *Controller) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		account, err := c.AccountFromRequestHeader(ctx, request.Header())
		if err != nil {
			return nil, err
		}
		if !account.Active {
			return nil, ErrDeactivated
		}
		return next(context.WithValue(ctx, ContextKey, account), request)
	})
}

func (c *Controller) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// noop
		return next(ctx, spec)
	})
}

func (c *Controller) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		account, err := c.AccountFromRequestHeader(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}
		if !account.Active {
			return ErrDeactivated
		}
		return next(context.WithValue(ctx, ContextKey, account), conn)
	})
}
