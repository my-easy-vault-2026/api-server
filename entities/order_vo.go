package entities

import (
	"shared-modules/utils"

	"github.com/shopspring/decimal"
)

type TransactionRecordVO struct {
	ID                      uint64           `json:"id,string"` //可能為空
	Type                    string           `json:"type"`
	TransactionNO           string           `json:"transactionNo"` //可能為空
	UserID                  uint64           `json:"userId,string"`
	CardID                  uint64           `json:"cardId,string,omitempty"` //可能為空
	Income                  string           `json:"income"`
	IncomeCategoryID        uint64           `json:"incomeCategoryId,string"`
	IncomeCurrency          string           `json:"incomeCurrency"`
	AgainstIncome           string           `json:"againstIncome,omitempty"`                  //可能為空
	AgainstIncomeCategoryID uint64           `json:"againstIncomeCategoryId,string,omitempty"` //可能為空
	AgainstIncomeCurrency   string           `json:"againstIncomeCurrency,omitempty"`          //可能為空
	Side                    string           `json:"side,omitempty"`                           //可能為空
	FromCardID              uint64           `json:"fromCardId,string,omitempty"`              //可能為空
	FromCategoryID          uint64           `json:"fromCategoryId,string,omitempty"`          //可能為空
	FromCategory            string           `json:"fromCategory,omitempty"`                   //可能為空
	FromCurrency            string           `json:"fromCurrency,omitempty"`                   //可能為空
	FromAmount              string           `json:"fromAmount,omitempty"`                     //可能為空
	FromUserID              uint64           `json:"fromUserId,string,omitempty"`              //可能為空
	FromEmail               string           `json:"fromEmail,omitempty"`                      //可能為空
	ToCardID                uint64           `json:"toCardId,string,omitempty"`                //可能為空
	ToCategoryID            uint64           `json:"toCategoryId,string,omitempty"`            //可能為空
	ToCategory              string           `json:"toCategory,omitempty"`                     //可能為空
	ToCurrency              string           `json:"toCurrency,omitempty"`                     //可能為空
	ToAmount                string           `json:"toAmount,omitempty"`                       //可能為空
	ToUserID                uint64           `json:"toUserId,string,omitempty"`                //可能為空
	ToEmail                 string           `json:"toEmail,omitempty"`                        //可能為空
	ExchangeRate            *decimal.Decimal `json:"exchangeRate,string,omitempty"`            //可能為空
	ExchangeFee             string           `json:"exchangeFee,omitempty"`                    //可能為空
	ExchangeFeeCurrency     string           `json:"exchangeFeeCurrency,omitempty"`            //可能為空
	TransferFee             string           `json:"transferFee,omitempty"`                    //可能為空
	TransferFeeCurrency     string           `json:"transferFeeCurrency,omitempty"`            //可能為空
	Status                  string           `json:"status"`
	CreatedAt               int64            `json:"createdAt,string"`
	UpdatedAt               int64            `json:"updatedAt,string"`
}

type PageTransactionRecordsVO struct {
	utils.PageData[[]*TransactionRecordVO]
}
