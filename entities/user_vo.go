package entities

import (
	"shared-modules/common"
)

type GetInfoVO struct {
	ID        uint64      `json:"id,string"`
	Email     string      `json:"email"`
	Gender    string      `json:"gender"`
	CreatedAt int64       `json:"createdAt,string"`
	UpdatedAt int64       `json:"updatedAt,string"`
	Role      common.Role `json:"role"`
	GroupIDs  []uint64    `json:"groupIDs"`
}
