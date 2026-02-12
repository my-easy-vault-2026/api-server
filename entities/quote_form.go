package entities

import "github.com/my-easy-vault-2026/shared-modules/common"

type GetExchangeRateForm struct {
	Purpose common.RatePurpose `json:"purpose"`
}
