package entities

type TransferPreviewVO struct {
	ToAmount       string `json:"toAmount"`
	ToUserID       uint64 `json:"toUserId,string"`
	ToWalletID     uint64 `json:"toWalletId,string"`
	ToCategoryID   uint64 `json:"toCategoryId,string"`
	ToCurrency     string `json:"toCurrency"`
	FromAmount     string `json:"fromAmount"`
	FromWalletID   uint64 `json:"fromWalletId,string"`
	FromCategoryID uint64 `json:"fromCategoryId,string"`
	FromCurrency   string `json:"fromCurrency"`
	Fee            string `json:"fee,omitempty"`
	FeeCurrency    string `json:"feeCurrency"`
	Key            string `json:"key"`
	ExpiredAt      int64  `json:"expiredAt,string"`
}

type TransferConfirmVO struct {
	OrderNO string `json:"orderNO,omitempty"`
}
