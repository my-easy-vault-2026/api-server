package entities

import (
	"github.com/shopspring/decimal"
)

type RateResponseVO struct {
	Code    int                      `json:"code"`
	Data    map[string]float64       `json:"data"`
	List    []map[string]interface{} `json:"list"`
	Message string                   `json:"message"`
}

type RateResponseMatrix []RateResponseUnit
type RateResponseUnit struct {
	Source string          `json:"source"`
	Dest   string          `json:"dest"`
	Value  decimal.Decimal `json:"value"`
}
