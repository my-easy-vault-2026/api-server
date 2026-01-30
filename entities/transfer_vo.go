package entities

import "github.com/shopspring/decimal"

type TransferPreviewVO struct {
	ToAmount       string            `json:"toAmount"`
	ToUserID       uint64            `json:"toUserId,string"`
	ToCardID       uint64            `json:"toCardID,string"`
	ToCategory     string            `json:"toCategory,omitempty"`
	ToCategoryID   uint64            `json:"toCategoryId,string"`
	ToCurrency     string            `json:"toCurrency"`
	ToEmail        string            `json:"toEmail,omitempty"`
	ToAddress      string            `json:"toAddress,omitempty"`
	FromAmount     string            `json:"fromAmount"`
	FromCardID     uint64            `json:"fromCardId,string"`
	FromCategory   string            `json:"fromCategory,omitempty"`
	FromCategoryID uint64            `json:"fromCategoryId,string"`
	FromCurrency   string            `json:"fromCurrency"`
	Mainnet        string            `json:"mainnet,omitempty"`
	Protocol       string            `json:"protocol,omitempty"`
	ExchangeRate   *decimal.Decimal  `json:"exchangeRate,string,omitempty"`
	ExchangeFee    string            `json:"exchangeFee,omitempty"`
	Fee            string            `json:"fee,omitempty"`
	FeeCurrency    string            `json:"feeCurrency,omitempty"`
	Rate           []*ExchangeRateVO `json:"rate"`
	DisplayRate    ExchangeRateVO    `json:"displayRate"`
	TransferKey    string            `json:"transferKey"`
	ExpiredAt      int64             `json:"expiredAt,string"`
	MainnetName    string            `json:"mainnetName"`
}

type TransferConfirmVO struct {
	OrderNO string `json:"orderNO"`
}
