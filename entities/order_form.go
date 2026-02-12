package entities

import "github.com/my-easy-vault-2026/shared-modules/common"

type PageTransactionRecordsForm struct {
	CardID uint64 `json:"cardId,string"`
	common.Page
}
