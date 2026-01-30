package entities

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"shared-modules/utils"
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

func (m RateResponseMatrix) ExistSource(source string) bool {
	return utils.SliceContain(m, func(e RateResponseUnit) bool {
		return e.Source == source
	})
}

func (m RateResponseMatrix) ExistDest(dest string) bool {
	return utils.SliceContain(m, func(e RateResponseUnit) bool {
		return e.Dest == dest
	})
}

func (m RateResponseMatrix) Exist(src, dest string) bool {
	return utils.SliceContain(m, func(e RateResponseUnit) bool {
		return e.Source == src && e.Dest == dest
	})
}

func (m RateResponseMatrix) GetBySource(source string) RateResponseMatrix {
	var resp RateResponseMatrix
	for _, v := range m {
		if v.Source == source {
			tv := v
			resp = append(resp, tv)
		}
	}
	return resp
}

func (m RateResponseMatrix) GetByDest(dest string) RateResponseMatrix {
	var resp RateResponseMatrix
	for _, v := range m {
		if v.Dest == dest {
			tv := v
			resp = append(resp, tv)
		}
	}
	return resp
}

func (m RateResponseMatrix) GetBySourceAndDest(source, dest string) *RateResponseUnit {
	for _, v := range m {
		if v.Source == source && v.Dest == dest {
			tv := v
			return &tv
		}
	}
	return nil
}

// bitop
type BitopRateResponseVO struct {
	Code    int                       `json:"code"`
	Data    map[string]float64        `json:"data"`
	List    []BitopRateResponseUnitVO `json:"list"`
	Message string                    `json:"message"`
}

type BitopRateResponseUnitVO struct {
	Name  string  `json:"name"` // src currency
	Value float64 `json:"value"`
}

func (vo BitopRateResponseVO) Get(src string) float64 {
	for _, v := range vo.List {
		if v.Name == src {
			return v.Value
		}
	}
	return 0.0
}

// binance
type BinanceRateResponseVO []BinanceRateResponseUnitVO

type BinanceRateResponseUnitVO struct {
	Symbol string `json:"symbol"` // "srcdest" currency pair
	Price  string `json:"price"`
}

func (vos BinanceRateResponseVO) Gets(src string) BinanceRateResponseVO {
	var resp []BinanceRateResponseUnitVO
	for _, v := range vos {
		if strings.HasPrefix(v.Symbol, src) {
			r := v
			resp = append(resp, r)
		}
	}
	return resp
}

func (vos BinanceRateResponseVO) Get(src, dest string) *BinanceRateResponseUnitVO {
	s := fmt.Sprintf("%s%s", src, dest)
	for _, v := range vos {
		if v.Symbol == s {
			r := v
			return &r
		}
	}
	return nil
}
