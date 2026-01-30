package entities

import "shared-modules/common"

// SystemParameterForm represents the form data for creating or updating a system parameter.
type ListSystemParametersForm struct {
	CategoryID common.ParameterCategory `json:"categoryId,string"`
	Key        common.ParameterKey      `json:"key" validate:"required,max=255"`
}

// CurrencyForm represents the form data for creating or updating a currency.
type ListCurrenciesForm struct {
	Type common.CurrencyType `json:"type"`
}

type QueryVersionLimitForm struct {
	Platform string                `json:"platform"`
	Version  string                `json:"version"`
	Keys     []common.ParameterKey `json:"keys"`
}

type UpdateAPPVersionLimitForm struct {
	Platform string `json:"platform" `
	Version  string `json:"version" validate:"required"`
}

type GetAssetCategoryForm struct {
	Type common.AssetType `json:"type"`
}

type DeviceInfoForm struct {
	ID uint64 `json:"id,string"`
}
