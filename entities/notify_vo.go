package entities

import (
	"database/sql"
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"github.com/shopspring/decimal"
)

type SendOTPVO struct {
}

type NotifyVO struct {
	Code         string `json:"code,omitempty"`
	Expire       string `json:"expire,omitempty"`
	ClientIp     string `json:"clientIp,omitempty"`
	DeviceName   string `json:"deviceName,omitempty"`
	CardNo       string `json:"cardNo,omitempty"`
	CardTitle    string `json:"cardTitle,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Channel      string `json:"channel,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Mainnet      string `json:"mainnet,omitempty"`
	OrderNo      string `json:"orderNo,omitempty"`
	OTPCode      string `json:"otpCode,omitempty"`
	Wallet       string `json:"wallet,omitempty"`
	TxFailReason string `json:"txFailReason,omitempty"`
	Threeds      string `json:"threeds,omitempty"`
	KycLevel     string `json:"kycLevel,omitempty"`
	NoticeID     uint64 `json:"noticeID,omitempty"`
	Date         string `json:"date,omitempty"`
	TrackingLink string `json:"trackingLink,omitempty"`
}

type SendVO struct {
	Email    string           `json:"email"`
	Subject  string           `json:"subject"`
	Code     common.MsgOPCode `json:"code"`
	SubCode  string           `json:"subCode"`
	UserID   uint64           `json:"userId,string"`
	Language common.Language  `json:"language"`
}

type NotifyTemplateVO struct {
	ID         uint64            `json:"id,string"`
	Code       string            `json:"code"`
	SubCode    string            `json:"subCode"`
	Subject    string            `json:"subject"`
	Template   string            `json:"template"`
	NotifyType common.NotifyType `json:"notifyType"`
	Language   common.Language   `json:"language"`
	Extra      string            `json:"extra"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	CreateUser uint64            `json:"createUser,string"`
	UpdateUser uint64            `json:"updateUser,string"`
}

type ListNotifyTemplateVO struct {
	Records []*NotifyTemplateVO `json:"records"`
}

type PageNotifyTemplateVO struct {
	utils.PageData[[]*NotifyTemplateVO]
}

type GetTemplateDetailVO struct {
	Code         string               `json:"code"`
	SubCode      string               `json:"subCode"`
	Placeholders map[string]string    `json:"placeholders"`
	CreateUser   uint64               `json:"createUser,string"`
	UpdateUser   uint64               `json:"updateUser,string"`
	CreatedAt    int64                `json:"createdAt,string"`
	UpdatedAt    int64                `json:"updatedAt,string"`
	Templates    []*GetTemplateDetail `json:"templates"`
}

type GetTemplateDetail struct {
	ID         uint64            `json:"id,string"`
	Subject    string            `json:"subject"`
	Template   string            `json:"template"`
	NotifyType common.NotifyType `json:"notifyType"`
	Language   common.Language   `json:"language"`
	Extra      string            `json:"extra"`
}

type PageAdminSendNotifyLogVO struct {
	utils.PageData[[]*AdminSendNotifyLogVO]
}
type AdminSendNotifyLogVO struct {
	ID         uint64 `json:"id,string"`         // 主鍵
	Code       string `json:"code"`              // message code
	SubCode    string `json:"subCode"`           //
	Recipients string `json:"recipients"`        // 收件用戶id（json格式或逗號分隔）
	Subject    string `json:"subject"`           // 主題
	Template   string `json:"template"`          // 內容
	CreatedAt  int64  `json:"createdAt"`         // 創建時間
	UpdatedAt  int64  `json:"updatedAt"`         // 更新時間
	CreateUser uint64 `json:"createUser,string"` // 創建人
	UpdateUser uint64 `json:"updateUser,string"` // 更新人
}

type GetAdminSendNotifyLogDetailVO struct {
	ID           uint64               `json:"id,string"`
	Code         string               `json:"code"`
	SubCode      string               `json:"subCode"`
	Language     string               `json:"language"` // 語言代碼，如：en、ja
	Placeholders map[string]string    `json:"placeholders"`
	Recipients   string               `json:"recipients"` // 收件用戶id（json格式或逗號分隔）
	CreateUser   uint64               `json:"createUser,string"`
	UpdateUser   uint64               `json:"updateUser,string"`
	CreatedAt    int64                `json:"createdAt,string"`
	UpdatedAt    int64                `json:"updatedAt,string"`
	Templates    []*GetTemplateDetail `json:"templates"`
}

type GetMessageCodeListVO struct {
	Records []GetMessageCodeVo `json:"records"`
}

type GetMessageCodeVo struct {
	Code    string `json:"code"`
	SubCode string `json:"subCode"`
}

type GetSupportLanguageVO struct {
	Records []common.Language `json:"records"`
}

type AdminNotifyMetricVO []*AdminNotifyMetricVOUnit

type AdminNotifyMetricVOUnit struct {
	Id         uint64            `json:"id,string"`
	NotifyType common.NotifyType `json:"notifyType"`
	Total      uint64            `json:"total,string"`
	Success    uint64            `json:"success,string"`
	Fail       uint64            `json:"fail,string"`
	Pending    uint64            `json:"pending,string"`
}

type SendCardMsgData struct {
	MsgOPCode common.MsgOPCode
	Card      *NotifyCard
}

type NotifyCard struct {
	ID                  uint64                   `json:"id,string"`
	UserID              uint64                   `json:"userID,string"`
	Type                common.AssetType         `json:"type,string"`
	UserFirstName       string                   `json:"userFirstName,string"`
	UserLastName        string                   `json:"userLastName,string"`
	ProductName         string                   `json:"productName,string"`
	PreferredName       string                   `json:"preferredName,string"`
	SecondaryName       string                   `json:"secondaryName,string"`
	Alias               string                   `json:"alias,string"`
	CategoryID          uint64                   `json:"categoryID,string"`
	IssueID             string                   `json:"issueID,string"`
	PANNumber           string                   `json:"panNumber,string"`
	SecurityCode        string                   `json:"securityCode,string"`
	ExpiredAt           time.Time                `json:"expiredAt"`
	Currency            common.Currency          `json:"currency,string"`
	CurrencyType        common.CurrencyType      `json:"currencyType,string"`
	Amount              *decimal.Decimal         `json:"amount,string"`
	Organization        common.CardOrganization  `json:"organization,string"`
	Vendor              common.CardProductVendor `json:"vendor,string"`
	WhaleCardType       common.WhaleCardType     `json:"whaleCardType,string"`
	WhaleCardBin        string                   `json:"whaleCardBin,string"`
	Issuer              string                   `json:"issuer,string"`
	Format              common.CardFormat        `json:"format,string"`
	SpendLimit          decimal.Decimal          `json:"spendLimit,string"`
	Design              string                   `json:"design,string"`
	CustomDesign        string                   `json:"customDesign,string"`
	MerchantID          uint64                   `json:"merchantID,string"`
	WhaleUserID         uint64                   `json:"whaleUserID,string"`
	WhaleCardID         uint64                   `json:"whaleCardID,string"`
	WhaleERCAddress     string                   `json:"whaleERCAddress,string"`
	WhaleTRCAddress     string                   `json:"whaleTRCAddress,string"`
	PaycryptoTypeID     string                   `json:"paycryptoTypeID,string"`
	PaycryptoCardNO     string                   `json:"paycryptoCardNO,string"`
	EtherfiUserID       string                   `json:"etherfiUserID,string"`
	BalanceType         common.BalanceType       `json:"balanceType,string"`
	ForwardType         common.ForwardType       `json:"forwardType,string"`
	FromAutoTopUp       common.AutoTopUpStatus   `json:"fromAutoTopUp,string"`
	ToAutoTopUp         common.AutoTopUpStatus   `json:"toAutoTopUp,string"`
	Auto3DS             common.Auto3DSStatus     `json:"auto3DS,string"`
	ATMToggle           common.ATMToggle         `json:"atmToggle,string"`
	AccumulatedEarnings *decimal.Decimal         `json:"accumulatedEarnings,string"`
	LastEarnings        *decimal.Decimal         `json:"lastEarnings,string"`
	BlockReason         *common.CardBlockReason  `json:"blockReason,string"`
	FreezeReason        *sql.NullInt32           `json:"freezeReason,string"` // CardFreezeReason
	ReapDeleteStatus    common.ReapDeleteStatus  `json:"reapDeleteStatus,string"`
	Status              common.CardStatus        `json:"status,string"`
	DeliveryStatus      common.DeliveryStatus    `json:"deliveryStatus,string"`
	FreezeStatus        common.CardFreezeStatus  `json:"freezeStatus,string"`
	EarningStatus       common.CardEarningStatus `json:"earningStatus,string"`
	RiskyStatus         common.CardRiskyStatus   `json:"riskyStatus,string"`
	FrozenAt            time.Time                `json:"frozenAt"`
	BlockedAt           time.Time                `json:"blockedAt"`
	DeletedAt           time.Time                `json:"deletedAt"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           time.Time                `json:"updatedAt"`
}
