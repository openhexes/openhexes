package test

import (
	"github.com/openhexes/proto/game/v1/gamev1connect"
	"github.com/openhexes/proto/iam/v1/iamv1connect"
)

type Clients struct {
	IAM  iamv1connect.IAMServiceClient
	Game gamev1connect.GameServiceClient
}