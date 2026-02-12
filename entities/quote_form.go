package entities

import "shared-modules/common"

type GetExchangeRateForm struct {
	Purpose common.RatePurpose `json:"purpose"`
}
