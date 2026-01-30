package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
)

type PageTransactionRecordsForm struct {
	CardID     uint64   `json:"cardId,string"`
	CategoryID uint64   `json:"categoryId,string"`
	Category   string   `json:"category"`
	Types      []string `json:"types"`
	utils.Page
}

type PageAutoYieldTransactionRecordsForm struct {
	Currency string   `json:"currency"`
	Types    []string `json:"types"`
	utils.Page
}

type AdminPageTransactionRecordsForm struct {
	CardID        uint64                       `json:"cardId,string"`
	CategoryID    uint64                       `json:"categoryId,string"`
	MerchantID    uint64                       `json:"merchantId,string"`
	UserID        uint64                       `json:"userId,string"`
	Category      string                       `json:"category"`
	Type          common.TransactionRecordType `json:"type"`
	StatusIn      []common.TransactionStatus   `json:"statusIn"`
	TransactionId uint64                       `json:"transactionId,string"`
	TransactionNo string                       `json:"transactionNo"`
	CreatedAtFrom int64                        `json:"createdAtFrom,string"`
	CreatedAtTo   int64                        `json:"createdAtTo,string"`
	PromotionCode string                       `json:"promotionCode"`
	utils.Page
}

type AdminGetTransactionRecordForm struct {
	OrderNO string `json:"orderNo"`
	Side    string `json:"side"`
}

type MerchantPageTransactionRecordsForm struct {
	CardID        uint64 `json:"cardId,string"`
	CategoryID    uint64 `json:"categoryId,string"`
	Category      string `json:"category"`
	CreatedAtFrom int64  `json:"createdAtFrom,string" validate:"required"`
	CreatedAtTo   int64  `json:"createdAtTo,string" validate:"required"`
	utils.Page
}

type MerchantPageUserTransactionRecordsForm struct {
	CardID        uint64 `json:"cardId,string"`
	CategoryID    uint64 `json:"categoryId,string"`
	UserID        uint64 `json:"userId,string"`
	Email         string `json:"email"`
	Category      string `json:"category"`
	CreatedAtFrom int64  `json:"createdAtFrom,string" validate:"required"`
	CreatedAtTo   int64  `json:"createdAtTo,string" validate:"required"`
	utils.Page
}

type MerchantGetTransactionRecordForm struct {
	OrderNO string `json:"orderNo"`
}

type GetTransactionRecordForm struct {
	OrderNO string `json:"orderNO"`
	Side    string `json:"side"`
	CardID  uint64 `json:"cardId,string"`
}

type PageCardIdentiyForm struct {
	utils.Page
}

type GetCardIdentiyForm struct {
	ID uint64 `json:"id,string"`
}

type AdminTransactionRecordsDownloadPageForm struct {
	CardID        uint64                       `json:"cardId,string"`
	CategoryID    uint64                       `json:"categoryId,string"`
	MerchantID    uint64                       `json:"merchantId,string"`
	UserID        uint64                       `json:"userId,string"`
	Category      string                       `json:"category"`
	Type          common.TransactionRecordType `json:"type"`
	StatusIn      []common.TransactionStatus   `json:"statusIn"`
	TransactionId uint64                       `json:"transactionId,string"`
	TransactionNo string                       `json:"transactionNo"`
	CreatedAtFrom int64                        `json:"createdAtFrom,string"`
	CreatedAtTo   int64                        `json:"createdAtTo,string"`
	Size          int                          `json:"size"`
	PromotionCode string                       `json:"promotionCode"`
}
