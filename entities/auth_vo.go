package entities

import (
	"shared-modules/common"
	"time"

	"github.com/shopspring/decimal"
)

type LoginOrCreateVO struct {
	ID           uint64                 `json:"id,string"`
	Token        string                 `json:"token"`
	ExpiredAt    int64                  `json:"expiredAt,string"`
	Email        string                 `json:"email"`
	HasPinCode   bool                   `json:"hasPinCode"`
	CountryCode  uint8                  `json:"countryCode,omitempty"` //可能為空
	PhoneNumber  string                 `json:"phoneNumber,omitempty"` //可能為空
	Gender       uint8                  `json:"gender,omitempty"`      //可能為空
	KycLevel     common.KYCLevel        `json:"kycLevel,omitempty"`
	Channel      string                 `json:"channel,omitempty"`
	EPoint       decimal.Decimal        `json:"ePoint,omitempty"`
	EPointLevel  string                 `json:"ePointLevel,omitempty"`
	EPointCardID uint64                 `json:"ePointCardId,string,omitempty"`
	Device       DeviceInfo             `json:"device"`
	BlockStatus  common.UserBlockStatus `json:"blockStatus,omitempty"`
	CreatedAt    int64                  `json:"createdAt"`
	UpdatedAt    int64                  `json:"updatedAt"`
}

type AdminLoginVO struct {
	ID        uint64            `json:"id,string"`
	Token     string            `json:"token"`
	WsToken   string            `json:"wsToken"`
	Name      string            `json:"name"`
	Level     common.AdminLevel `json:"level"`
	GroupIDs  []uint64          `json:"groupIds,omitempty"`
	CreatedAt int64             `json:"createdAt,string"`
	UpdatedAt int64             `json:"updatedAt,string"`
	Role      common.Role       `json:"role"`
}

type AgentLoginVO struct {
	ID        uint64      `json:"id,string"`
	Token     string      `json:"token"`
	WsToken   string      `json:"wsToken"`
	Name      string      `json:"name"`
	Level     string      `json:"level"`
	CreatedAt int64       `json:"createdAt,string"`
	UpdatedAt int64       `json:"updatedAt,string"`
	Role      common.Role `json:"role"`
}

type IsExistPasskeyUserDevice struct {
	Exist bool `json:"exist"`
}

type AccountGroup struct {
	ID           uint64            `json:"id,string"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Role         common.Role       `json:"role"`
	Level        common.AdminLevel `json:"level"`
	ExclusiveIDs string            `json:"exclusiveIDs"`
	CreatedAt    time.Time         `json:"createdAt,string"`
	UpdatedAt    time.Time         `json:"updatedAt,string"`
}
