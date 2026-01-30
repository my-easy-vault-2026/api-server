package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SupportedVO struct {
	Mainnet     string `json:"mainnet,omitempty"`
	MainnetName string `json:"mainnetName,omitempty"`
	Protocol    string `json:"protocol"`
	Recommended bool   `json:"recommended,omitempty"`
}

type CardSwitchVO struct {
	EnableApplePay bool `json:"enableApplePay"`
}

type CardVO struct {
	ID               uint64                 `json:"id,string,omitempty"`
	UserID           uint64                 `json:"userId,string"`
	MerchantID       uint64                 `json:"merchantId,string,omitempty"`
	Email            string                 `json:"email"`
	CategoryID       uint64                 `json:"categoryId,string"`
	Vendor           string                 `json:"vendor"`
	PreferredName    string                 `json:"preferredName"`
	Type             string                 `json:"type"`
	Organization     string                 `json:"organization,omitempty"`
	Amount           decimal.Decimal        `json:"amount,omitempty"`
	PANNumber        string                 `json:"panNumber,omitempty"`
	Issuer           string                 `json:"issuer"`
	Currency         string                 `json:"currency"`
	PointLevel       string                 `json:"pointLevel,omitempty"`       //可能為空
	PointToNextLevel *decimal.Decimal       `json:"pointToNextLevel,omitempty"` //
	NextPointLevel   string                 `json:"nextPointLevel,omitempty"`   //可能為空
	Alias            string                 `json:"alias"`
	Format           string                 `json:"format"`
	Supported        []*SupportedVO         `json:"supported,omitempty"`
	CardSwitch       map[string]interface{} `json:"cardSwitch,omitempty"`
	BalanceType      string                 `json:"balanceType,omitempty"`
	ForwardType      string                 `json:"forwardType,omitempty"`
	AutoYieldStatus  string                 `json:"autoYieldStatus,omitempty"`
	Design           string                 `json:"design,omitempty"`
	DisplayStatus    common.DisplayStatus   `json:"displayStatus,omitempty"`
	DeliveryStatus   string                 `json:"deliveryStatus,omitempty"`
	TrackingLink     string                 `json:"trackingLink,omitempty"`
	Status           string                 `json:"status"`
	FreezeStatus     string                 `json:"freezeStatus"`
	HasCardIdentity  bool                   `json:"hasCardIdentity"`
	CreatedAt        int64                  `json:"createdAt,string"`
	UpdatedAt        int64                  `json:"updatedAt,string"`
}

type ListCardVO struct {
	Records    []*CardVO `json:"records"`
	MainCardID uint64    `json:"mainCardId,string,omitempty"`
}

type PrivacyInfoVO struct {
	ID             uint64 `json:"id,string"`
	CardNumber     string `json:"cardNumber"`
	CardHolderName string `json:"cardHolderName"`
	CCV            string `json:"ccv"`
	ValidThru      string `json:"validThru"`
	Associaion     string `json:"association"`
}

type PrivacyInfoHTMLVO struct {
	URL string `json:"url"`
}

type ListPrivacyInfoVO struct {
	Records []*PrivacyInfoVO `json:"records"`
}

type ExchangeRateVO struct {
	BaseCurrency  string          `json:"baseCurrency"`
	QuoteCurrency string          `json:"quoteCurrency"`
	Rate          decimal.Decimal `json:"rate"`
	Timestamp     int64           `json:"timestamp,string"`
	Purpose       string          `json:"purpose"`
}

type ApplyCheckEligibilityVO struct {
}

type ApplyPreviewVO struct {
	Success               bool              `json:"success"`
	DepositAmount         string            `json:"depositAmount"`
	DepositAmountShortage string            `json:"depositAmountShortage,omitempty"`
	CategoryID            uint64            `json:"categoryID,string"`
	FromCategory          string            `json:"fromCategory"`
	ToCurrency            string            `json:"toCurrency"`
	ReceiveAmount         string            `json:"receiveAmount,omitempty"`
	FromDiscount          string            `json:"fromDiscount,omitempty"`
	ToDiscount            string            `json:"toDiscount,omitempty"`
	Bonus                 string            `json:"bonus,omitempty"`
	PromotionCode         string            `json:"promotionCode,omitempty"`
	Fee                   string            `json:"fee,omitempty"`
	CardFee               string            `json:"cardFee,omitempty"`
	Rate                  []*ExchangeRateVO `json:"rate"`
	DisplayRate           ExchangeRateVO    `json:"displayRate"`
	CardApplyKey          string            `json:"cardApplyKey,omitempty"`
	ExpiredAt             int64             `json:"expiredAt,string,omitempty"`
}

type ApplyConfirmVO struct {
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Parameter  string `json:"parameter,omitempty"`
}

type FeatureVO struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type FeeVO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CardCategoryLegacyVO struct {
	ID                uint64                 `json:"id,string"`
	Name              string                 `json:"name"`
	PreferredName     string                 `json:"preferredName,omitempty"`
	SecondaryName     string                 `json:"secondaryName,omitempty"`
	Type              string                 `json:"type"`
	Currency          string                 `json:"currency"`
	CurrencyType      string                 `json:"currencyType"`
	CategorySwitch    map[string]interface{} `json:"categorySwitch,omitempty"`
	Organization      string                 `json:"organization,omitempty"`
	ActivationDeposit string                 `json:"activationDeposit,omitempty"`
	LimitCountryUS    bool                   `json:"limitCountryUS"`
	Vendor            string                 `json:"vendor,omitempty"`
	Format            common.CardFormat      `json:"format,omitempty"`
	Title             string                 `json:"title,omitempty"`
	Feature           []string               `json:"feature,omitempty"`
	Description       string                 `json:"description,omitempty"`
	SpendLimit        *decimal.Decimal       `json:"spendLimit,omitempty"`
	ValidMonths       int                    `json:"validMonths,omitempty"`
	CardDesign        string                 `json:"cardDesign,omitempty"`
	Fee               *decimal.Decimal       `json:"fee,omitempty"`
	FeeCurrency       string                 `json:"feeCurrency,omitempty"`
	Supported         []*SupportedVO         `json:"supported,omitempty"`
	Code              []string               `json:"code,omitempty"`
	ProductId         string                 `json:"productId,omitempty"`
	CreatedAt         int64                  `json:"createdAt,string"`
	UpdatedAt         int64                  `json:"updatedAt,string"`
}

type ListCardCategoryLegacyVO struct {
	Records []*CardCategoryLegacyVO `json:"records"`
}

type CardCategoryVO struct {
	ID                uint64                 `json:"id,string"`
	Name              string                 `json:"name"`
	PreferredName     string                 `json:"preferredName,omitempty"`
	SecondaryName     string                 `json:"secondaryName,omitempty"`
	Type              string                 `json:"type"`
	Currency          string                 `json:"currency"`
	CurrencyType      string                 `json:"currencyType"`
	CategorySwitch    map[string]interface{} `json:"categorySwitch,omitempty"`
	Organization      string                 `json:"organization,omitempty"`
	AnnualFee         string                 `json:"annualFee,omitempty"`
	ActivationDeposit string                 `json:"activationDeposit,omitempty"`
	LimitCountryUS    bool                   `json:"limitCountryUS"`
	Vendor            string                 `json:"vendor,omitempty"`
	Format            common.CardFormat      `json:"format,omitempty"`
	Title             string                 `json:"title,omitempty"`
	Feature           []FeatureVO            `json:"feature,omitempty"`
	Fees              []FeeVO                `json:"fees,omitempty"`
	PurchasableItems  []string               `json:"purchasableItems,omitempty"`
	Wallets           []string               `json:"wallets,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Tagline           string                 `json:"tagline,omitempty"`
	SpendLimit        *decimal.Decimal       `json:"spendLimit,omitempty"`
	ValidMonths       int                    `json:"validMonths,omitempty"`
	CardDesign        string                 `json:"cardDesign,omitempty"`
	Fee               *decimal.Decimal       `json:"fee,omitempty"`
	FeeCurrency       string                 `json:"feeCurrency,omitempty"`
	Supported         []*SupportedVO         `json:"supported,omitempty"`
	Code              []string               `json:"code,omitempty"`
	ProductId         string                 `json:"productId,omitempty"`
	CreatedAt         int64                  `json:"createdAt,string"`
	UpdatedAt         int64                  `json:"updatedAt,string"`
}

type ListCardCategoryVO struct {
	Records []*CardCategoryVO `json:"records"`
}

type FreezeCardVO struct {
}

type BlockCardVO struct{}

type Check3dsVO struct {
	ID                   uint64            `json:"id,string"`
	CardID               uint64            `json:"cardId,string"`
	MerchantID           uint64            `json:"merchantId,string,omitempty"`
	Alias                string            `json:"alias" gorm:"default:null"`
	PANNumberSuffix      string            `json:"panNumberSuffix" gorm:"default:null"`
	Balance              string            `json:"balance"`
	Status               string            `json:"status"`
	MerchantName         string            `json:"merchantName"`
	TransactionAmount    string            `json:"transactionAmount"`
	MerchantCountryCode  common.NationCode `json:"merchantCountryCode"`
	TransactionCurrency  string            `json:"transactionCurrency"`
	TransactionTimestamp int64             `json:"transactionTimestamp,string"`
	ExpiredAt            int64             `json:"expiredAt,string,omitempty"`
	CreatedAt            int64             `json:"createdAt,string,omitempty"`
	UpdatedAt            int64             `json:"updatedAt,string,omitempty"`
}

type UnblockPreviewVO struct {
	FromAmount      string            `json:"fromAmount"`
	FromCardID      uint64            `json:"fromCardID,string"`
	FromCategory    string            `json:"fromCategory"`
	FromCategoryID  uint64            `json:"fromCategoryID,string"`
	FromCurrency    string            `json:"fromCurrency"`
	ToAmount        string            `json:"toAmount"`
	ToCardID        uint64            `json:"toCardID,string"`
	ToCategory      string            `json:"toCategory"`
	ToCategoryID    uint64            `json:"toCategoryID,string"`
	ToCurrency      string            `json:"toCurrency"`
	Fee             string            `json:"fee"` // 充值手續費
	ExchangeFee     string            `json:"exchangeFee,omitempty"`
	Rate            []*ExchangeRateVO `json:"rate"`
	DisplayRate     ExchangeRateVO    `json:"displayRate"`
	Key             string            `json:"key"`
	ExpiredAt       int64             `json:"expiredAt,string"`
	PointAmount     string            `json:"pointAmount"`
	PointCardID     uint64            `json:"pointCardID"`
	PointCategory   string            `json:"pointCategory"`
	PointCategoryID uint64            `json:"pointCategoryID"`
	PointCurrency   string            `json:"pointCurrency"`
}

type UnblockConfirmVO struct {
	OrderNO string `json:"orderNO"`
}

type UpdatePINVO struct {
}

type CardGetRemarkVO struct {
	CardID  uint64          `json:"cardId,string"`
	Remarks []*CardRemarkVO `json:"remarks"`
}

type CardRemarkVO struct {
	Id        uint64 `json:"id,string"`
	CreatedAt int64  `json:"createdAt,string"`
	CreatedBy uint64 `json:"createdBy,string"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
}

type CardSaveRemarkVO struct {
}

type PaycryptoPhysicalCardNumberVO []*PaycryptoPhysicalCardNumberVOUnit

type PaycryptoPhysicalCardNumberVOUnit struct {
	Id         uint64                                   `json:"id,string"`
	PanNumber  string                                   `json:"panNumber" `
	InviteCode string                                   `json:"inviteCode" `
	UserId     uint64                                   `json:"userId,string,omitempty"`
	CardID     uint64                                   `json:"cardId,string,omitempty"`
	Status     common.PaycryptoPhysicalCardNumberStatus `json:"status,omitempty"`
}

type PhysicalCardOrderPageVO struct {
	utils.PageData[[]*PhysicalCardOrderPageUnitVO]
}

type PhysicalCardOrderPageUnitVO struct {
	OrderId   uint64 `json:"orderId,string"`
	Batch     string `json:"batch"`
	BulkShiId string `json:"bulkShiId"`
	Size      uint64 `json:"size"`
	Remark    string `json:"remark"`
	CreatedAt int64  `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
	UpdatedAt int64  `json:"updatedAt"`
}

type PhysicalCardOrderGetVO struct {
	OrderId   uint64 `json:"orderId,string"`
	Batch     string `json:"batch"`
	BulkShiId string `json:"bulkShiId"`
	Size      uint64 `json:"size"`
	Remark    string `json:"remark"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	CreatedBy string `json:"createdBy"`
}

type PhysicalCardOrderCreateVO struct {
	OrderId    uint64     `json:"orderId,string"`
	BulkShipId *uuid.UUID `json:"bulkShipId,string"`
	CutoffDate time.Time  `json:"cutoffDate"`
}

type PhysicalCardOrderMetricVO struct {
	Total   uint64 `json:"total,string"`
	Success uint64 `json:"success,string"`
	Failed  uint64 `json:"failed,string"`
}

type PhysicalCardOrderRemarkListVO []*PhysicalCardOrderRemarkUnit

type PhysicalCardOrderRemarkUnit struct {
	Id            uint64 `json:"id,string"`
	OrderId       uint64 `json:"orderId,string"`
	Remark        string `json:"remark"`
	CreatedAt     int64  `json:"createdAt"`
	CreatedBy     uint64 `json:"createdBy,string"`
	CreatedByName string `json:"createdByName"`
}

type PhysicalCardOrderRemarkGetVO struct {
	Id            uint64 `json:"id,string"`
	OrderId       uint64 `json:"orderId,string"`
	Remark        string `json:"remark"`
	CreatedAt     int64  `json:"createdAt"`
	CreatedBy     uint64 `json:"createdBy,string"`
	CreatedByName string `json:"createdByName"`
}

type PhysicalCardCardListVO struct {
	UserId          uint64                                  `json:"userId,string"`
	CardId          uint64                                  `json:"cardId,string"`
	IssueId         string                                  `json:"issueId"`
	KycLevel        common.KYCLevel                         `json:"kycLevel"`
	Email           string                                  `json:"email"`
	MerchantId      uint64                                  `json:"merchantId"`
	Provider        common.CardProductVendor                `json:"provider"`
	Merchant        common.PhysicalCardRequestOrderMerchant `json:"merchant"`
	CardStatus      common.CardStatus                       `json:"cardStatus"`
	CategoryID      uint64                                  `json:"categoryID"`
	CardKind        common.CardKind                         `json:"cardKind"`
	CardDesign      string                                  `json:"cardDesign"`
	RecipientName   string                                  `json:"recipientName"`
	DeliveryAddress string                                  `json:"deliveryAddress"`
	DeliveryCity    string                                  `json:"deliveryCity"`
	DeliveryPostal  string                                  `json:"deliveryPostal"`
	DeliveryCountry string                                  `json:"deliveryCountry"`
	PhoneNumber     string                                  `json:"phoneNumber"`
	DeliveryStatus  common.DeliveryStatus                   `json:"deliveryStatus"`
	AppliedAt       time.Time                               `json:"appliedAt"`
	ItemStatus      common.PhysicalCardRequestItemStatus    `json:"itemStatus"`
	Log             string                                  `json:"log"`
	Line1           string                                  `json:"line1"`
	LastOrderId     uint64                                  `json:"lastOrderId,string"`
	LastOrderBatch  string                                  `json:"lastOrderBatch"`
}

type PhysicalCardItemListVO struct {
	OrderId         uint64                                  `json:"orderId,string"`
	ItemId          uint64                                  `json:"itemId,string"`
	UserId          uint64                                  `json:"userId,string"`
	CardId          uint64                                  `json:"cardId,string"`
	IssueId         string                                  `json:"issueId"`
	KycLevel        common.KYCLevel                         `json:"kycLevel"`
	Email           string                                  `json:"email"`
	MerchantId      uint64                                  `json:"merchantId"`
	Provider        common.CardProductVendor                `json:"provider"`
	Merchant        common.PhysicalCardRequestOrderMerchant `json:"merchant"`
	OrderStatus     common.PhysicalCardRequestItemStatus    `json:"orderStatus"`
	CardStatus      common.CardStatus                       `json:"cardStatus"`
	CardKind        common.CardKind                         `json:"cardKind"`
	CardDesign      string                                  `json:"cardDesign"`
	RecipientName   string                                  `json:"recipientName"`
	DeliveryAddress string                                  `json:"deliveryAddress"`
	DeliveryCity    string                                  `json:"deliveryCity"`
	DeliveryPostal  string                                  `json:"deliveryPostal"`
	DeliveryCountry string                                  `json:"deliveryCountry"`
	PhoneNumber     string                                  `json:"phoneNumber"`
	DeliveryStatus  common.DeliveryStatus                   `json:"deliveryStatus"`
	AppliedAt       time.Time                               `json:"appliedAt"`
	ItemStatus      common.PhysicalCardRequestItemStatus    `json:"itemStatus"`
	Log             string                                  `json:"log"`
	Line1           string                                  `json:"line1"`
}

type PhysicalCardItemDownloadVO struct {
	OrderId         uint64          `json:"orderId,string"`
	ItemId          uint64          `json:"itemId,string"`
	UserId          uint64          `json:"userId,string"`
	CardId          uint64          `json:"cardId,string"`
	IssueId         string          `json:"issueId"`
	KycLevel        common.KYCLevel `json:"kycLevel"`
	Email           string          `json:"email"`
	MerchantId      uint64          `json:"merchantId"`
	Provider        string          `json:"provider"`
	Merchant        string          `json:"merchant"`
	OrderStatus     string          `json:"orderStatus"`
	CardStatus      string          `json:"cardStatus"`
	CardKind        common.CardKind `json:"cardKind"`
	CardDesign      string          `json:"cardDesign"`
	RecipientName   string          `json:"recipientName"`
	DeliveryAddress string          `json:"deliveryAddress"`
	DeliveryCity    string          `json:"deliveryCity"`
	DeliveryPostal  string          `json:"deliveryPostal"`
	DeliveryCountry string          `json:"deliveryCountry"`
	PhoneNumber     string          `json:"phoneNumber"`
	DeliveryStatus  string          `json:"deliveryStatus"`
	AppliedAt       time.Time       `json:"appliedAt"`
	ItemStatus      string          `json:"itemStatus"`
	Log             string          `json:"log"`
	Line1           string          `json:"line1"`
}
