package entities

import (
	"shared-modules/utils"

	"github.com/shopspring/decimal"
)

type AssetVO struct {
	ID           uint64          `json:"id,string"`
	CategoryID   uint64          `json:"categoryId,string"`
	Type         string          `json:"type"`
	UserID       uint64          `json:"userId,string"`
	Currency     string          `json:"currency"`
	CurrencyType string          `json:"currencyType"`
	Amount       decimal.Decimal `json:"amount"`
	FrozenAmount decimal.Decimal `json:"frozenAmount"`
	CreatedAt    int64           `json:"createdAt,string"`
	UpdatedAt    int64           `json:"updatedAt,string"`
}

type ListAssetsVO struct {
	Records []*AssetVO `json:"records"`
}

type AccountingVO struct {
	UserID   uint64          `json:"userId,string"`
	AssetID  uint64          `json:"assetId,string"`
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
	OrderNO  string          `json:"orderNO"`
}

type ManualTransferVO struct {
	OrderNO string `json:"orderNo"`
}

type BillVO struct {
	ID                  uint64          `json:"id,string"`
	UserID              uint64          `json:"userId,string"`
	AssetID             uint64          `json:"assetId,string"`
	OrderNo             string          `json:"orderNo"`
	Amount              decimal.Decimal `json:"amount"`
	CurrentAmount       decimal.Decimal `json:"currentAmount"`
	CurrentFrozenAmount decimal.Decimal `json:"currentFrozenAmount"`
	Currency            string          `json:"currency"`
	TransactionType     string          `json:"transactionType"`
	BillType            string          `json:"billType"`
	OrderType           string          `json:"orderType"`
	CreatedAt           int64           `json:"createdAt,string"`
	UpdatedAt           int64           `json:"updatedAt,string"`
}

type PageBillVO struct {
	utils.PageData[[]*BillVO]
}

type GetEquivalentValueVO struct {
	Amount       decimal.Decimal            `json:"amount"`
	FrozenAmount decimal.Decimal            `json:"frozenAmount"`
	Fees         map[string]decimal.Decimal `json:"fees"`
	Rates        []*ExchangeRateVO          `json:"rates"`
	ActualRates  []*ExchangeRateVO          `json:"autualRates"`
}
