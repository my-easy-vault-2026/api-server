package entities

import (
	"shared-modules/common"

	"github.com/shopspring/decimal"
)

// SystemParameterVO represents the system parameter data for response.
type ListSystemParametersVO struct {
	Records []*SystemParameterVO `json:"records"`
}

// SystemParameter represents a system parameter.
type SystemParameterVO struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CurrencyVO represents the currency data for response.
type ListCurrenciesVO struct {
	Records []*CurrencyVO `json:"records"`
}

// Currency represents a currency.
type CurrencyVO struct {
	Code   common.Currency     `json:"code"`
	Symbol string              `json:"name"`
	Type   common.CurrencyType `json:"type"`
}

type VersionLimitVO struct {
	IsLimit               bool `json:"isLimit"`
	IsLimitCryptoExchange bool `json:"isLimitCryptoExchange"`
}

type KycSettingVO struct {
	ID                uint64          `json:"id"`
	KycLevel          common.KYCLevel `json:"kycLevel"`
	Name              string          `json:"name"`
	MonthlyTotal      decimal.Decimal `json:"monthlyTotal"`
	DailyTotal        decimal.Decimal `json:"dailyTotal"`
	DailyCount        int             `json:"dailyCount"`
	MonthlyCount      int             `json:"monthlyCount"`
	PerTransaction    decimal.Decimal `json:"perTransaction"`
	AtmMonthlyTotal   decimal.Decimal `json:"atmMonthlyTotal"`
	AtmDailyTotal     decimal.Decimal `json:"atmDailyTotal"`
	AtmDailyCount     int             `json:"atmDailyCount"`
	AtmMonthlyCount   int             `json:"atmMonthlyCount"`
	AtmPerTransaction decimal.Decimal `json:"atmPerTransaction"`
}

type ListKycVO struct {
	Records []*KycSettingVO `json:"records"`
}

type AssetCategoryVO struct {
	ID           uint64           `json:"id,string"`
	Name         string           `json:"name"`
	Type         common.AssetType `json:"type"`
	CardType     common.CardType  `json:"cardType"`
	CardKind     common.CardKind  `json:"cardKind,string"`
	WhaleCardBin string           `json:"whaleCardBin"`
}

type ListAssetCategoryVO struct {
	Records []*AssetCategoryVO `json:"records"`
}

type DeviceInfoVO struct {
	Platform    string          `json:"platform"`
	DeviceId    string          `json:"deviceId"`
	DeviceName  string          `json:"deviceName"`
	AppVersion  string          `json:"appVersion"`
	AppLanguage common.Language `json:"appLanguage"`
	LoginTime   int64           `json:"loginTime"`
	Activated   bool            `json:"activated"`
}

type DeviceListVO struct {
	Devices []*DeviceInfoVO `json:"devices"`
}
