package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
)

type MessageRecordVO struct {
	ID        uint64                       `json:"id"`
	UserID    uint64                       `json:"userID"`
	Subject   string                       `json:"subject"`
	Content   string                       `json:"content"`
	Status    common.UserMessageReadStatus `json:"status"` // 0 for unread, 1 for read
	Extra     string                       `json:"extra"`
	Code      common.MsgOPCode             `json:"code"`
	CreatedAt int64                        `json:"createdAt,string"`
}

type PageMessageRecordsVO struct {
	IsRead      bool `json:"isRead"`
	UnReadCount int  `json:"unReadCount"`
	utils.PageData[[]*MessageRecordVO]
}
