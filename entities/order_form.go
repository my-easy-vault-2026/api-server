package entities

import "api-server/common"

type PageTransactionRecordsForm struct {
	CardID uint64 `json:"cardId,string"`
	common.Page
}
