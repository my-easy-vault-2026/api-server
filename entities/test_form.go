package entities

import "github.com/shopspring/decimal"

type AddAssetsForm struct {
	UserID  uint64          `json:"userId,string"`
	AssetID uint64          `json:"assetId,string"`
	Amount  decimal.Decimal `json:"amount"`
}
