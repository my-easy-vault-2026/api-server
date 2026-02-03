package entities

type ExchangePreviewVO struct {
	FromAmount     string            `json:"fromAmount"`
	FromWalletID   uint64            `json:"fromCardID,string"`
	FromCategory   string            `json:"fromCategory"`
	FromCategoryID uint64            `json:"fromCategoryID,string"`
	FromCurrency   string            `json:"fromCurrency"`
	ToAmount       string            `json:"toAmount"`
	ToWalletID     uint64            `json:"toCardID,string"`
	ToCategory     string            `json:"toCategory"`
	ToCategoryID   uint64            `json:"toCategoryID,string"`
	ToCurrency     string            `json:"toCurrency"`
	ExchangeFee    string            `json:"exchangeFee"`
	Rate           []*ExchangeRateVO `json:"rate"`
	Key            string            `json:"key"`
	ExpiredAt      int64             `json:"expiredAt,string"`
}

type ExchangeConfirmVO struct{}
