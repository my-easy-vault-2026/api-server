package entities

import (
	"shared-modules/common"
	"time"

	"github.com/shopspring/decimal"
)

type SetPinCodeVO struct {
}

type ResetPinCodeVO struct {
}

type IdentityVO struct {
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Dob              string `json:"dob"`
	IdDocumentType   int    `json:"idDocumentType"`
	IdDocumentNumber string `json:"idDocumentNumber"`

	AddressLine1 string           `json:"addressLine1"`
	AddressLine2 string           `json:"addressLine2"`
	NationCode   string           `json:"nationCode"`
	PostalCode   string           `json:"postalCode"`
	City         string           `json:"city"`
	Status       common.KYCStatus `json:"status"`
	FailReason   string           `json:"failReason"`
	Country      string           `json:"country"`
	State        string           `json:"state"`
	ValidatedAt  time.Time        `json:"ValidatedAt"`
	DayOfBornAt  time.Time        `json:"dayOfBornAt"`
}

type DeviceInfo struct {
	DeviceID      string `json:"deviceID"`
	IsMaster      bool   `json:"isMaster"`
	DeviceCount   int    `json:"deviceCount"`
	HasPasskey    bool   `json:"hasPasskey"`
	DevicePasskey bool   `json:"devicePasskey"`
	DeviceName    string `json:"deviceName"`
	Platform      string `json:"platform"`
}

type GetInfoVO struct {
	ID               uint64                 `json:"id,string"`
	Email            string                 `json:"email"`
	FinishedIdentity string                 `json:"finishedIdentity"`
	CountryCode      int                    `json:"countryCode,omitempty"` //可能為空
	PhoneNumber      string                 `json:"phoneNumber,omitempty"` //可能為空
	KycLevel         common.KYCLevel        `json:"kycLevel,omitempty"`
	KycStatus        common.KYCStatus       `json:"kycStatus,omitempty"`
	Gender           string                 `json:"gender"`
	Channel          string                 `json:"channel,omitempty"`
	EPoint           decimal.Decimal        `json:"ePoint,omitempty"`
	EPointLevel      string                 `json:"ePointLevel,omitempty"`
	EPointCardID     uint64                 `json:"ePointCardId,string,omitempty"`
	Auto3DS          string                 `json:"auto3ds,omitempty"`
	AutoTopUp        string                 `json:"autoTopUp,omitempty"`
	ATMToggle        string                 `json:"atmToggle,omitempty"`
	CreatedAt        int64                  `json:"createdAt,string"`
	UpdatedAt        int64                  `json:"updatedAt,string"`
	Identity         IdentityVO             `json:"identity,omitempty"`
	Device           DeviceInfo             `json:"device"`
	Language         common.Language        `json:"language,omitempty"`
	PromotionCode    string                 `json:"promotionCode,omitempty"`
	PromotionLink    string                 `json:"promotionLink,omitempty"`
	ReferrerCode     string                 `json:"referrerCode,omitempty"`
	BlockStatus      common.UserBlockStatus `json:"blockStatus,omitempty"`
	LastLoginAt      int64                  `json:"LastLoginAt,string"`
	Role             common.Role            `json:"role"`
	GroupIDs         []uint64               `json:"groupIDs"`
}

type ApplyKycVO struct {
	Link      string `json:"link"`
	SessionId string `json:"session_id"`
}

type QueryKycVO struct {
	Status        int    `json:"status"`
	PassTime      string `json:"pass_time"`
	VerifyCounter int    `json:"verify_counter"`
	SyncUrl       string `json:"sync_url"`
}

type QueryKycResponse struct {
	Code int        `json:"code"`
	Data QueryKycVO `json:"data"`
	Msg  string     `json:"msg"`
}

type ApplyKycResponse struct {
	Code int        `json:"code"`
	Data ApplyKycVO `json:"data"`
	Msg  string     `json:"msg"`
}

type SaveIdentityVO struct {
}

type UserLimitsVO struct {
	ID                  uint64                `json:"id"`                  // 主鍵
	UserID              uint64                `json:"userID"`              // 用戶ID，與用戶表關聯
	LimitType           common.UserLimitsType `json:"limitType"`           // 限額類型：卡片限額或轉帳限額
	MonthlyTotal        decimal.Decimal       `json:"monthlyTotal"`        // 每月總交易量限額
	CurrentMonthlyTotal decimal.Decimal       `json:"currentMonthlyTotal"` // 當前每月總交易量
	DailyTotal          decimal.Decimal       `json:"dailyTotal"`          // 每日總交易量限額
	CurrentDailyTotal   decimal.Decimal       `json:"currentDailyTotal"`   // 當前每日總交易量
	DailyCount          int                   `json:"dailyCount"`          // 每日消費次數限額
	CurrentDailyCount   int                   `json:"currentDailyCount"`   // 當前每日消費次數
	MonthlyCount        int                   `json:"monthlyCount"`        // 每月消費次數限額
	CurrentMonthlyCount int                   `json:"currentMonthlyCount"` // 當前每月消費次數
	PerTransaction      decimal.Decimal       `json:"perTransaction"`      // 每筆交易金額限額
	MonthlyTotalMax     decimal.Decimal       `json:"monthlyTotalMax"`     //每月交易限額上限
	DailyTotalMax       decimal.Decimal       `json:"dailyTotalMax"`       //每日交易限額上限
	PerTransactionMax   decimal.Decimal       `json:"perTransactionMax"`   //單次交易金額上限
}

type GetDappInfoVO struct {
	Email      string          `json:"email"`
	CardAmount decimal.Decimal `json:"cardAmount,string"`
	CardList   []*DappCardVO   `json:"cardList"`
	ApplyList  []*DappApplyVO  `json:"applyList"`
}

type DappCardVO struct {
	ID           uint64          `json:"id,string,omitempty"`
	UserID       uint64          `json:"userId,string"`
	Amount       decimal.Decimal `json:"amount,omitempty"`
	PANNumber    string          `json:"panNumber,omitempty"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	FreezeStatus string          `json:"freezeStatus"`
}

type DappApplyVO struct {
	FromCardID     uint64 `json:"fromCardID,string"`
	FromCategoryID uint64 `json:"fromCategoryID,string"`
	FromCurrency   string `json:"fromCurrency"`
	ToCategoryID   uint64 `json:"toCategoryID,string"`
	ToCurrency     string `json:"toCurrency"`
}

type CheckApplyKycVO struct {
	Status common.KYCStatus `json:"status"`
}

type UserTempAddressVO struct {
	UserID     uint64                      `json:"userID"`
	Protocol   common.Protocol             `json:"protocol"`
	Mainnet    common.Mainnet              `json:"mainnet"`
	Currency   common.Currency             `json:"currency"`
	UsageType  common.UserTempAddressUsage `json:"usageType"`
	AssignedAt *time.Time                  `json:"assignedAt"`
}

type FirebaseTokenVO struct {
	FirebaseToken string          `json:"firebaseToken"`
	Language      common.Language `json:"language"`
}

type KYCRequestStatusVO struct {
	KYCLevel      common.KYCLevel         `json:"kycLevel"`
	RequestStatus common.KYCRequestStatus `json:"requestStatus"`
}
