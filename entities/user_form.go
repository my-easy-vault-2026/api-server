package entities

import (
	"shared-modules/common"
)

type SetPinCodeForm struct {
	PinCode string `json:"pinCode" validate:"required,numeric"`
}

type ResetPinCodeForm struct {
	NewPinCode string `json:"newPinCode" validate:"required,number,len=6"`
	OldPinCode string `json:"oldPinCode"`
}

type ForgotPinCodeForm struct {
	NewPinCode string `json:"newPinCode" validate:"required,number,len=6"`
	OTPCode    string `json:"otpCode"`
}

type GetInfoForm struct {
}

type GetUserForm struct {
	ID    uint64
	Email string
	Role  common.Role
}

type ListUsersForm struct {
	UserIDs []uint64 `json:"userIds,string" validate:"required"`
}

type CoinfaceQueryKycStatus struct {
	CompanyId     int64  `json:"company_id"`
	CompanyUserId string `json:"company_user_id"`
	Sign          string `json:"sign"`
}

type CoinfaceQueryKyc3Status struct {
	CompanyId     int64  `json:"company_id"`
	CompanyUserId string `json:"company_user_id"`
	Sign          string `json:"sign"`
}

type CoinfacePreCreation struct {
	AsyncUrl    string `json:"async_url"`
	CompanyId   int64  `json:"company_id"`
	SyncUrl     string `json:"sync_url"`
	Sign        string `json:"sign"`
	FaceCompare bool   `json:"face_compare"`
	UserImg     string `json:"user_img"`

	OpenUploadCertificates bool `json:"open_upload_certificates"`
	DuplicateRegisterCheck bool `json:"duplicate_register_check"`
}

type CoinfaceKyc3PreCreation struct {
	CompanyId              int64  `json:"company_id"`
	CompanyUserId          string `json:"company_user_id"`
	SyncUrl                string `json:"sync_url"`
	AsyncUrl               string `json:"async_url"`
	OpenOcr                bool   `json:"open_ocr"`
	OpenUploadCertificates bool   `json:"open_upload_certificates"`
	DuplicateRegisterCheck bool   `json:"duplicate_register_check"`
	BlackListCheck         bool   `json:"black_list_check"`
	AmlCheck               bool   `json:"aml_check"`
	Name                   string `json:"name"`
	Birthday               string `json:"birthday"`
	Sign                   string `json:"sign"`
}

type SaveIdentityLegacyForm struct {
	FirstName        string                      `json:"firstName" validate:"required"`
	LastName         string                      `json:"lastName" validate:"required"`
	Dob              string                      `json:"dob" validate:"required"` // 出生日期 YYYY-MM-DD
	IdDocumentType   common.IdentityDocumentType `json:"idDocumentType" validate:"required"`
	IdDocumentNumber string                      `json:"idDocumentNumber" validate:"required"`
	Gender           string                      `json:"gender"`
	CountryCode      string                      `json:"countryCode" validate:"required"`
	PhoneNumber      string                      `json:"phoneNumber" validate:"required"`

	AddressLine1 string `json:"addressLine1" validate:"required"` // 街道
	AddressLine2 string `json:"addressLine2"`                     // 楼层单元号
	NationCode   string `json:"nationCode"`
	City         string `json:"city"` // 城市
	PostalCode   string `json:"postalCode"`
}

type SaveIdentityForm struct {
	FirstName        string                      `json:"firstName" validate:"required"`
	LastName         string                      `json:"lastName" validate:"required"`
	Dob              string                      `json:"dob" validate:"required"` // 出生日期 YYYY-MM-DD
	IdDocumentType   common.IdentityDocumentType `json:"idDocumentType" validate:"required"`
	IdDocumentNumber string                      `json:"idDocumentNumber" validate:"required"`
	Gender           string                      `json:"gender" validate:"required"`
	CountryCode      string                      `json:"countryCode" validate:"required"`
	PhoneNumber      string                      `json:"phoneNumber" validate:"required"`

	AddressLine1 string `json:"addressLine1" validate:"required"` // 街道
	AddressLine2 string `json:"addressLine2"`                     // 楼层单元号
	NationCode   string `json:"nationCode"`
	City         string `json:"city"` // 城市
	PostalCode   string `json:"postalCode" validate:"omitempty,postal"`
}

type ApplyKycForm struct {
	RedirectURL string `json:"redirectUrl"`
}

type CheckApplyKycForm struct {
	KycLevel common.KYCLevel `json:"kycLevel" validate:"required"`
}

type KycCallbackCertificate struct {
	CertificatesName   string `json:"certificates_name"`
	CertificatesNumber string `json:"certificates_number"`
}

type KycCallbackLivePhoto struct {
	Live1 string `json:"live_1"`
	Live2 string `json:"live_2"`
}

type KycCallbackData struct {
	SessionId   string `json:"session_id"`
	LivePhoto   string `json:"live_photo"`
	Status      int    `json:"status"`
	FaceCompare int    `json:"face_compare"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}

type Kyc3SyncData struct {
	UserID string `json:"userid"`
}

type Kyc3Callback struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data Kyc3CallbackData `json:"data"`
}

type Kyc3CallbackData struct {
	CompanyUserID             string         `json:"company_user_id"`
	BlackList                 int            `json:"black_list"`
	Status                    int            `json:"status"`
	Reason                    *string        `json:"reason"` // 因为可能为 null
	DuplicateRegistrationList string         `json:"duplicate_registration_list"`
	LivePhoto                 LivePhoto      `json:"live_photo"`
	CertificatePhoto          string         `json:"certificate_photo"`
	OcrResult                 OcrResult      `json:"ocr_result"`
	SessionID                 string         `json:"session_id"`
	AmlCheck                  AmlCheckResult `json:"aml_check"`
}

type AmlCheckResult struct {
	AmlStatus     int      `json:"aml_status"`
	CategoryName  string   `json:"category_name"`
	Name          string   `json:"name"`
	Nationality   []string `json:"nationality"`
	Citizenship   []string `json:"citizenship"`
	Country       []string `json:"country"`
	Birthdate     []string `json:"birth_date"`
	PositionTitle []string `json:"position_title"`
}

type LivePhoto struct {
	Live1 string `json:"live_1"`
	Live2 string `json:"live_2"`
}

type OcrResult struct {
	NameByCertificates        string               `json:"name_by_certificates"`
	NumberByCertificates      string               `json:"number_by_certificates"`
	EnglishNameByCertificates string               `json:"english_name_by_certificates"`
	Birthday                  string               `json:"birthday"`
	Sex                       string               `json:"sex"`
	IssueDate                 string               `json:"issue_date"`
	ExpiryDate                string               `json:"expiry_date"`
	CertificatesType          string               `json:"certificates_type"`
	CertificateCheckInfo      CertificateCheckInfo `json:"certificate_check_info"`
}

type CertificateCheckInfo struct {
	CertificateCheckResult string   `json:"certificate_check_result"`
	RiskType               []string `json:"risk_type"`
}

type KycCallbackForm struct {
	/**
	{"code": 1023,
	"msg": "\u7528\u6237\u91cd\u590d\u6ce8\u518c",
	"data": {
		"reason": "\u7528\u6237\u91cd\u590d\u6ce8\u518c,\u91cd\u590d\u7684\u7528\u6237:1002",
		"status": 3,
		"certificate": {"certificates_name": "", "certificates_number": ""},
		"live_photo": {"live_1": "xxx", "live_2": "ttt"}
		}
	 }
	*/
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data KycCallbackData `json:"data"`
}

type SavePhoneNumberForm struct {
	CountryCode string `json:"countryCode"`
	PhoneNumber string `json:"phoneNumber"`
}

type DeleteAccountForm struct {
	PinCode string `json:"pinCode" validate:"required,numeric"`
}

type RegistFirebaseTokenForm struct {
	UserID        uint64          `json:"-"`
	DeviceID      string          `json:"-"`
	FirebaseToken string          `json:"firebaseToken" validate:"required"`
	Language      common.Language `json:"language"`
}

type SaveLanguageForm struct {
	Language common.Language `json:"language"`
	UserID   uint64          `json:"-"`
}

type UploadDocumentPhotoForm struct {
	UserID uint64 `form:"userId" binding:"required,min=1"`
}
