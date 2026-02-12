package entities

import (
	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/shopspring/decimal"
)

type WalletVO struct {
	ID         uint64            `json:"id,string"`
	UserID     uint64            `json:"userId,string"`
	CategoryID uint64            `json:"categoryId,string"`
	Currency   string            `json:"currency"`
	Amount     decimal.Decimal   `json:"amount,string"`
	Nation     string            `json:"nation"`
	Status     common.CardStatus `json:"status"`
	CreatedAt  int64             `json:"createdAt,string"`
	UpdatedAt  int64             `json:"updatedAt,string"`
}

type ListWalletsVO struct {
	Records []*WalletVO `json:"records"`
}

type CreateWalletVO struct {
	ID uint64 `json:"id,string"`
}

type CategoryVO struct {
	ID           uint64 `json:"id,string"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Currency     string `json:"currency"`
	CurrencyType string `json:"currencyType"`
	Nation       string
	NationCode   common.NationCode
	CreatedAt    int64 `json:"createdAt,string"`
	UpdatedAt    int64 `json:"updatedAt,string"`
}

type ListCategoryVO struct {
	Records []*CategoryVO `json:"records"`
}
