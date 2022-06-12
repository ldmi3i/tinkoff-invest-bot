package dtotapi

import (
	investapi "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"time"
)

type AccountStatus int

const (
	AccountStatusUnspecified = 0
	AccountStatusNew         = 1
	AccountStatusOpen        = 2
	AccountStatusClosed      = 3
)

func (s AccountStatus) String() string {
	switch s {
	case AccountStatusUnspecified:
		return "Unspecified"
	case AccountStatusNew:
		return "New"
	case AccountStatusOpen:
		return "Open"
	case AccountStatusClosed:
		return "Closed"
	default:
		return "Undefined"
	}
}

type AccessLevel int

const (
	AccountAccessLevelUnspecified = 0
	AccountAccessLevelFullAccess  = 1
	AccountAccessLevelReadOnly    = 2
	AccountAccessLevelNoAccess    = 3
)

type AccountsResponse struct {
	Accounts []*AccountResponse
}

type AccountResponse struct {
	Id          string
	Type        int
	Name        string
	Status      AccountStatus
	OpenedDate  time.Time
	ClosedData  time.Time
	AccessLevel AccessLevel
}

func AccountsResponseToDto(resp *investapi.GetAccountsResponse) *AccountsResponse {
	if resp == nil {
		return nil
	}
	accResps := make([]*AccountResponse, 0, len(resp.Accounts))
	for _, acc := range resp.Accounts {
		accResps = append(accResps, accountResponseToDto(acc))
	}
	return &AccountsResponse{
		accResps,
	}
}

func accountResponseToDto(resp *investapi.Account) *AccountResponse {
	if resp == nil {
		return nil
	}
	return &AccountResponse{
		Id:          resp.Id,
		Type:        int(resp.Type),
		Name:        resp.Name,
		Status:      AccountStatus(resp.Status),
		OpenedDate:  resp.OpenedDate.AsTime(),
		ClosedData:  resp.ClosedDate.AsTime(),
		AccessLevel: AccessLevel(resp.AccessLevel),
	}
}

//FindAccount simplifies search of account by id
func (ar *AccountsResponse) FindAccount(id string) (*AccountResponse, bool) {
	for _, acc := range ar.Accounts {
		if acc.Id == id {
			return acc, true
		}
	}
	return nil, false
}
