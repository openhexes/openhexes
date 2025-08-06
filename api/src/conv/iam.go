package conv

import (
	"github.com/openhexes/openhexes/api/src/db"
	v1 "github.com/openhexes/proto/iam/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AccountToProto(account *db.Account) *v1.Account {
	if account == nil {
		return nil
	}

	return &v1.Account{
		Id: account.ID.String(),
		Meta: &v1.Account_Meta{
			Active:      account.Active,
			CreatedAt:   timestamppb.New(account.CreatedAt.Time),
			DisplayName: account.DisplayName,
			Picture:     account.Picture,
		},
		Email: account.Email,
	}
}
