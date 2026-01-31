package entities

import (
	"shared-modules/utils"
)

type PageTransactionRecordsForm struct {
	CardID uint64 `json:"cardId,string"`
	utils.Page
}
