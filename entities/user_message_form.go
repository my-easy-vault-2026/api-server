package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
)

type PageMessageRecordsForm struct {
	Language common.Language         `json:"language"`
	Type     *common.UserMessageType `json:"type"`
	utils.Page
}

type ReadMessageForm struct {
	Id uint64 `json:"id"`
}
