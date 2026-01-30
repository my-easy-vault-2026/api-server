package entities

import (
	"shared-modules/common"

	"github.com/shopspring/decimal"
)

type WalletVO struct {
	ID         uint64              `json:"id,string"`
	UserID     uint64              `json:"userId,string"`
	CategoryID uint64              `json:"categoryId,string"`
	Currency   string              `json:"currency"`
	Status     common.WalletStatus `json:"status"`
	CreatedAt  int64               `json:"createdAt,string"`
	UpdatedAt  int64               `json:"updatedAt,string"`
}

type ListWalletsVO struct {
	Records []*WalletVO `json:"records"`
}

type WalletAddressVO struct {
	ID         uint64                     `json:"id,string"`
	CardID     uint64                     `json:"cardId,string"`
	Address    string                     `json:"address"`
	CategoryID uint64                     `json:"categoryId,string"`
	Mainnet    string                     `json:"mainnet"`
	Protocol   string                     `json:"protocol,omitempty"`
	Currency   string                     `json:"currency"`
	Status     common.WalletAddressStatus `json:"status"`
	CreatedAt  int64                      `json:"createdAt,string"`
	UpdatedAt  int64                      `json:"updatedAt,string"`
}

type ListWalletAddressesVO struct {
	Records []*WalletAddressVO `json:"records"`
}

type CreateWalletVO struct {
	ID            uint64 `json:"id,string"`
	WalletAddress string `json:"walletAdress"`
}

type CryptoCurrencyVO struct {
	ID               uint64
	Mainnet          common.Mainnet `gorm:"default:null"`
	MainnetName      string         `gorm:"default:null"`
	MainnetFullName  string         `gorm:"default:null"`
	Protocol         string         `gorm:"default:null"`
	ProtocolName     string         `gorm:"default:null"`
	CurrencyType     common.Currency
	CurrencyName     string
	CurrencyFullName string
	Flag             string               `gorm:"default:null"`
	Decimals         int                  `gorm:"default:null"`
	DisplayDecimals  int                  `gorm:"default:null"`
	CoinType         common.CoinTokenType `gorm:"default:null"`
}

type ListCryptoCurrencyVO struct {
	Records []*CryptoCurrencyVO `json:"records"`
}

type CheckWalletAndPreviewVO struct {
	ContractAddress string            `json:"contractAddress"`             // 鏈上幣種合約地址
	WalletAddress   string            `json:"walletAddress"`               // 鏈上收款地址
	Decimals        int               `json:"decimals"`                    // 鏈上交易精度
	Amount          decimal.Decimal   `json:"amount,string"`               // 支付最小金額
	ToDiscount      *decimal.Decimal  `json:"toDiscount,string,omitempty"` // 幣種為目的幣種
	TopUpFee        decimal.Decimal   `json:"topUpFee,string"`             // 交易手續費
	Rate            []*ExchangeRateVO `json:"rate"`
	DisplayRate     ExchangeRateVO    `json:"displayRate"`
	ExpiredAt       int64             `json:"expiredAt,omitempty"`
	ToCurrency      string            `json:"toCurrency"`
	FromCategory    string            `json:"fromCategory"`
	PromotionCode   string            `json:"promotionCode,omitempty"`
}

type NetCurrency struct {
	Mainnet  common.Mainnet
	Currency common.Currency
}
