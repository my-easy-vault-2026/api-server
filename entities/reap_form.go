package entities

import "shared-modules/common"

type CardTransactionCheckForm struct {
	CardID    uint64 `json:"cardID"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	// 設定取 前幾天的資料, 例如: 7 代表取最近7天的資料
	Days int `json:"days" validate:"omitempty,min=1,max=30"`
}

type ClearCardTransactionStagingForm struct {
	ClearDate string `json:"clearDate"`
}

type ReapBillTaskForm struct {
	ReportRange common.ReapReportRange `json:"reportRange" validate:"required,oneof=1 2"`
	TargetDate  string                 `json:"targetDate"`
}

type CardStatusCheckForm struct {
	CardID uint64 `json:"cardId,string"`
}
