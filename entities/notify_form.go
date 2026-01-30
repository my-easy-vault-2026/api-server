package entities

import (
	"shared-modules/common"
	"shared-modules/utils"
)

type SendOTPForm struct {
	Email   string                `json:"email" validate:"email"`
	Purpose common.MessagePurpose `json:"purpose" validate:"required"`
}

type RetryCallbackForm struct {
	Types  []common.CallbackType  `json:"types"`
	Scenes []common.CallbackScene `json:"scenes"`
}

type GetTemplateListForm struct {
	NotifyType common.NotifyType `json:"notifyType"`
	Language   common.Language   `json:"language"`
	Code       string            `json:"code"`
	SubCode    string            `json:"subCode"`
}

type GetTemplatePageForm struct {
	Code    string `json:"code"`
	SubCode string `json:"subCode"`
	utils.Page
}

type SendNotifyForm struct {
	UserIds      []string          `json:"userIds"`
	EMails       []string          `json:"emails"`
	Language     common.Language   `json:"language" `
	Code         string            `json:"code" validate:"required"`
	SubCode      string            `json:"subCode"`
	Placeholders map[string]string `json:"placeholders"`
}

type GetTemplateForm struct {
	Code    string `json:"code" validate:"required"`
	SubCode string `json:"subCode"`
}

type EditTemplateForm struct {
	Code      string          `json:"code" validate:"required"`
	SubCode   string          `json:"subCode"`
	Templates []*EditTemplate `json:"templates" validate:"required"`
}

type CreateTemplateForm struct {
	Code      string          `json:"code"`
	SubCode   string          `json:"subCode"`
	Templates []*EditTemplate `json:"templates"`
}

type EditTemplate struct {
	ID         uint64            `json:"id,string"`
	Subject    string            `json:"subject"`
	Template   string            `json:"template"`
	NotifyType common.NotifyType `json:"notifyType"`
	Language   common.Language   `json:"language"`
}

type AdminGetLogPageForm struct {
	Code    string `json:"code"`
	SubCode string `json:"subCode"`
	utils.Page
}

type AdminGetLogDetailForm struct {
	ID uint64 `json:"id,string" validate:"required"`
}

type AdminNotifyMetricForm struct {
	Id uint64 `json:"id,string" validate:"required"`
}

type AdminNotifyResendForm struct {
	Id uint64 `json:"id,string" validate:"required"`
}
