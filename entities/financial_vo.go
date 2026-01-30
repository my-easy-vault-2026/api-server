package entities

import "github.com/shopspring/decimal"

type AutoYieldInfoVO struct {
	DailyInterest                  string          `json:"dailyInterest"`
	PrincipalAmount                string          `json:"principalAmount"`
	AnnualYieldRate                decimal.Decimal `json:"annualYieldRate"`
	OneMonthAccumulatedInterest    string          `json:"oneMonthAccumulatedInterest"`
	TwoMonthsAccumulatedInterest   string          `json:"twoMonthsAccumulatedInterest"`
	ThreeMonthsAccumulatedInterest string          `json:"threeMonthsAccumulatedInterest"`
	EarningStatus                  string          `json:"earningStatus,omitempty"`
	CardID                         uint64          `json:"cardId,string,omitempty"`
	Threshold                      decimal.Decimal `json:"threshold,omitempty"`
	ThresholdCurrency              string          `json:"thresholdCurrency,omitempty"`
}

type AutoYieldHistoryVO struct {
	Records []*AutoYieldHistoryDataVO `json:"records"`
}

type AutoYieldHistoryData struct {
	Timestamp       int64           `json:"timestamp"`
	PrincipalAmount decimal.Decimal `json:"principalAmount"`
	Interest        decimal.Decimal `json:"interest"`
}

type AutoYieldHistoryDataVO struct {
	Timestamp       int64  `json:"timestamp,string"`
	PrincipalAmount string `json:"principalAmount"`
	Interest        string `json:"interest"`
}

type AutoYieldInterestList struct {
	IsEarning bool                   `json:"isEarning"`
	Records   []*AutoYieldInterestVO `json:"records"`
}

type AutoYieldInterestVO struct {
	Currency        string          `json:"currency"`
	EarningStatus   string          `json:"earningStatus"`
	AnnualYieldRate decimal.Decimal `json:"annualYieldRate"`
	PrincipalAmount string          `json:"principalAmount"`
	DailyInterest   string          `json:"dailyInterest"`
}
