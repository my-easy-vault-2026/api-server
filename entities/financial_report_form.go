package entities

import "shared-modules/common"

type DownloadFinancialReportForm struct {
	From     int64                         `json:"from,string" validate:"required"`
	To       int64                         `json:"to,string" validate:"required"`
	Table    common.FinancialReportTable   `json:"table" validate:"required"`
	Tables   []common.FinancialReportTable `json:"tables"`
	WithBill bool                          `json:"withBill"`
	TempRows int64                         `json:"-"`
}

type DownloadAssetReportForm struct {
	From  string                      `json:"from" validate:"required"`
	To    string                      `json:"to" validate:"required"`
	Table common.FinancialReportTable `json:"table" validate:"required"`
}

type DownloadEtherfiReportForm struct {
	From    int64                       `json:"from,string" validate:"required"`
	To      int64                       `json:"to,string" validate:"required"`
	Source  *common.EtherfiReportSource `json:"source"`
	Account uint64                      `json:"account" validate:"required"`
}

type DownloadEtherfiReportOptionForm struct {
}
