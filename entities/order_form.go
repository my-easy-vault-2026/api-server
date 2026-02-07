package entities

import "shared-modules/common"

type PageTransactionRecordsForm struct {
	CardID uint64 `json:"cardId,string"`
	common.Page
}
