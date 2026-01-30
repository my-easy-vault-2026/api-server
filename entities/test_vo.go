package entities

type MerchantSimulateAuthorizeVO struct {
	OrderNO string `json:"orderNo"`
}

type MerchantSimulateClearingVO struct {
	OrderNO string `json:"orderNo"`
}

type MerchantSimulateReversalVO struct {
	OrderNO string `json:"orderNo"`
}

type MerchantSimulateRefundVO struct {
	OrderNO string `json:"orderNo"`
}

type UploadS3VO struct {
	FileID uint64 `json:"fileId,string"`
}
