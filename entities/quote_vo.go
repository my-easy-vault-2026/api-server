package entities

import "github.com/shopspring/decimal"

type ExchangeRateVO struct {
	Base      string          `json:"base"`
	Quote     string          `json:"quote"`
	Rate      decimal.Decimal `json:"rate"`
	Timestamp int64           `json:"timestamp,string"`
	Purpose   string          `json:"purpose"`
}
