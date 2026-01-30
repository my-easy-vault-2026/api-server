package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type GetCardForm struct {
	ID         uint64 `json:"id,string"`
	CategoryID uint64 `json:"categoryId,string"`
	Category   string `json:"category"`
}

type ListCardForm struct {
	ID          uint64             `json:"id,string"`
	CategoryID  uint64             `json:"categoryId,string"`
	IssueID     string             `json:"issueId"`
	Currency    string             `json:"currency"`
	IDIn        []uint64           `json:"idIn,string"`
	AssetTypeIn []common.AssetType `json:"assetTypeIn"`
}

type SetMainCardForm struct {
	CardID uint64 `json:"cardId,string"`
}

type GetMainCardForm struct {
	CategoryID uint64 `json:"categoryId,string"`
	Currency   string `json:"currency"`
}

type ListPrivacyInfoForm struct {
	PinCode string `json:"pinCode" validate:"required"`
}

type ApplyPreviewForm struct {
	FromCategory      string                      `json:"fromCategory"`
	CategoryID        uint64                      `json:"categoryId,string" validate:"required"`
	PromotionCode     string                      `json:"promotionCode"`
	FromCardID        uint64                      `json:"fromCardId,string"`
	AddressLine1      string                      `json:"addressLine1"`
	State             string                      `json:"state"`
	City              string                      `json:"city"`
	PostalCode        string                      `json:"postalCode"`
	NationCode        common.NationCode           `json:"nationCode"`
	Address           string                      `json:"-"`
	Usage             common.UserTempAddressUsage `json:"-"`
	UserTempAddressID uint64                      `json:"-"`
}

type ApplyCheckEligibilityForm struct {
	CategoryID uint64 `json:"categoryId,string" validate:"required"`
}

type ApplyConfirmByApplePayForm struct {
	TransactionId string `json:"transactionId"`
}

type ListCardCategoryForm struct {
	ID                uint64              `json:"id,string"`
	Name              string              `json:"name"`
	PreferredName     string              `json:"preferredName,omitempty"`
	SecondaryName     string              `json:"secondaryName,omitempty"`
	Type              common.AssetType    `json:"type"`
	Currency          common.Currency     `json:"currency"`
	CurrencyType      common.CurrencyType `json:"currencyType"`
	ActivationDeposit decimal.Decimal     `json:"activationDeposit,omitempty"`
	Format            common.CardFormat   `json:"format,omitempty"`
	SpendLimit        decimal.Decimal     `json:"spendLimit,omitempty"`
	ValidMonths       int                 `json:"validMonths,omitempty"`
	CardDesign        string              `json:"cardDesign,omitempty"`
	Fee               decimal.Decimal     `json:"fee,omitempty"`
	FeeCurrency       common.Currency     `json:"feeCurrency,omitempty"`
	CreatedAt         time.Time           `json:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
}

type ListCategoryByUsageForm struct {
	Usages []common.CategoryUsage `json:"usages,omitempty"`
}

type ReactivateCardForm struct {
}

type Check3dsForm struct {
}

type AddCategoryForm struct {
	ID                  uint64               `json:"id,string,omitempty"`
	Name                string               `json:"name,omitempty"`
	PreferredName       string               `json:"preferredName"`
	SecondaryName       string               `json:"secondaryName"`
	Type                common.AssetType     `json:"type,string"`
	Currency            common.Currency      `json:"currency,string"`
	CurrencyType        common.CurrencyType  `json:"currencyType"`
	Issuer              string               `json:"issuer,omitempty"`
	ActivationDeposit   *decimal.Decimal     `json:"activationDeposit,omitempty"`
	Format              common.CardFormat    `json:"format"`
	SpendLimit          *decimal.Decimal     `json:"spendLimit,omitempty"`
	ValidMonths         int                  `json:"validMonths"`
	Design              string               `json:"design,omitempty"`
	Fee                 *decimal.Decimal     `json:"fee,omitempty"`
	FeeCurrency         common.Currency      `json:"feeCurrency,string"`
	Supported           string               `json:"supported,omitempty"`
	RecommendedMainnet  common.Mainnet       `json:"recommendedMainnet,omitempty"`
	RecommendedProtocol common.Protocol      `json:"recommendedProtocol,omitempty"`
	MerchantID          uint64               `json:"merchantId,string"`
	Usage               common.CategoryUsage `json:"usage"`
	ProductId           string               `json:"productId,omitempty"`
}

type Forward3dsForm struct {
	ID                   uint64            `json:"id,string"`
	CardID               uint64            `json:"cardId,string"`
	MerchantID           uint64            `json:"merchantId,omitempty,string"`
	Alias                string            `json:"alias" gorm:"default:null"`
	PANNumberSuffix      string            `json:"panNumberSuffix" gorm:"default:null"`
	Status               string            `json:"status"`
	MerchantName         string            `json:"merchantName"`
	TransactionAmount    string            `json:"transactionAmount"`
	MerchantCountryCode  common.NationCode `json:"merchantCountryCode"`
	TransactionCurrency  string            `json:"transactionCurrency"`
	TransactionTimestamp int64             `json:"transactionTimestamp,string"`
	ExpiredAt            int64             `json:"expiredAt,omitempty,string" gorm:"default:null"`
	CreatedAt            int64             `json:"createdAt,omitempty,string" gorm:"default:null"`
	UpdatedAt            int64             `json:"updatedAt,omitempty,string" gorm:"default:null"`
}

type ShipCardForm struct {
	CardID               uint64 `json:"cardId,string" validate:"required"`
	AddressLine1         string `json:"addressLine1"`
	AddressLine2         string `json:"addressLine2"`
	AddressZone          string `json:"addressZone"`
	AddressCity          string `json:"addressCity"`
	AddressPostalCode    string `json:"addressPostalCode"`
	AddressCountry       string `json:"addressCountry"`
	RecipientFirstName   string `json:"recipientFirstName"`
	RecipientLastName    string `json:"recipientLastName"`
	RecipientPhoneNumber string `json:"recipientPhoneNumber"`
	RecipientDialCode    int    `json:"recipientDialCode"`
	RecipientEmail       string `json:"recipientEmail"`
	Overwrite            bool   `json:"overwrite"`
}

type UpdatePINForm struct {
	CardID  uint64 `json:"cardId,string" validate:"required"`
	PIN     string `json:"pin" validate:"required,digit,min=4,max=12"`
	OTPCode string `json:"otpCode"`
}

type ActivatePhysicalCardForm struct {
	CardID         uint64 `json:"cardId,string" validate:"required"`
	ActivationCode string `json:"activationCode"`
	CVV            string `json:"cvv"`
	PINCode        string `json:"pinCode" validate:"required,digit,min=4,max=12"`
}

type DappApplyPreviewForm struct {
	FromCategory  string `json:"fromCategory"`
	CategoryID    uint64 `json:"categoryId,string" validate:"required"`
	PromotionCode string `json:"promotionCode"`
	FromCardID    uint64 `json:"-"`
	Currency      string `json:"currency"`
}

type CardBlockDeleteForm struct {
	BlockedAt int64 `json:"blockedAt"`
}

type CardBlockForm struct {
	CardID uint64 `json:"cardId,string" validate:"required"`
	Reason string `json:"reason"`
}

type CardUnblockPreviewForm struct {
	CardID     uint64 `json:"cardId,string" validate:"required"`
	FromCardID uint64 `json:"fromCardId,string"`
}

type CardUnblockConfirmForm struct {
	Key string `json:"key"`
}

type CardGetRemarkForm struct {
	CardID uint64 `json:"cardId,string" validate:"required"`
}

type CardSaveRemarkForm struct {
	CardID uint64 `json:"cardId,string" validate:"required"`
	Remark string `json:"remark" validate:"required"`
}

type PaycryptoPhysicalCardNumberUnit struct {
	PanNumber  string `json:"panNumber" validate:"required"`
	InviteCode string `json:"inviteCode"`
}

type PaycryptoPhysicalCardNumberCreateForm struct {
	Cards []PaycryptoPhysicalCardNumberUnit `json:"cards" validate:"required"`
}

type PaycryptoPhysicalCardNumberUpdateForm struct {
	Cards []PaycryptoPhysicalCardNumberUnit `json:"cards" validate:"required"`
}

type PaycryptoPhysicalCardNumberGetForm struct {
	PanNumber string `json:"panNumber"`
}

type PhysicalCardOrderPageForm struct {
	OrderId uint64 `json:"orderId,string"`
	Batch   string `json:"batch"`
	utils.Page
}

type PhysicalCardOrderGetForm struct {
	OrderId uint64 `json:"orderId,string" validate:"required"`
}

type PhysicalCardOrderCreateForm struct {
	CardIds []PhysicalCardOrderCardUnit `json:"cardIds" validate:"required"`
}

type PhysicalCardOrderCardUnit uint64

func (u *PhysicalCardOrderCardUnit) UnmarshalJSON(bs []byte) (err error) {
	s := strings.Trim(string(bs), "\"")
	tu, err := strconv.ParseUint(s, 0, 64)
	*u = PhysicalCardOrderCardUnit(tu)
	return err
}

type PhysicalCardOrderMetricForm struct {
	OrderId uint64 `json:"orderId,string" validate:"required"`
}

type PhysicalCardOrderRemarkListForm struct {
	OrderId uint64 `json:"orderId,string" validate:"required"`
}

type PhysicalCardOrderRemarkGetForm struct {
	OrderId uint64 `json:"orderId,string" validate:"required"`
}

type PhysicalCardOrderRemarkNewForm struct {
	OrderId uint64 `json:"orderId,string" validate:"required"`
	Remark  string `json:"remark" validate:"required"`
}

type PhysicalCardCardListForm struct {
	CardId         uint64                                    `json:"cardId,string"`
	CardIds        []uint64                                  `json:"cardIds"`
	Merchant       []common.PhysicalCardRequestOrderMerchant `json:"merchant"`
	DeliveryStatus common.DeliveryStatus                     `json:"deliveryStatus"`
	CardStatus     []common.CardStatus                       `json:"cardStatus"`
}

type PhysicalCardItemListForm struct {
	OrderId  uint64                               `json:"orderId,string"`
	CardId   uint64                               `json:"cardId,string"`
	CardKind common.CardKind                      `json:"cardKind"`
	Status   common.PhysicalCardRequestItemStatus `json:"status"`
}

type PhysicalCardItemDownloadForm struct {
	OrderId  uint64                               `json:"orderId,string"`
	CardId   uint64                               `json:"cardId,string"`
	CardKind common.CardKind                      `json:"cardKind"`
	Status   common.PhysicalCardRequestItemStatus `json:"status"`
}
