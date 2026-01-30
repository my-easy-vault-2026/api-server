package entities

type LoginOrCreateForm struct {
	Email         string `json:"email" validate:"required,email"`
	OTPCode       string `json:"otpCode" validate:"required,numeric"`
	Platform      string `json:"-"`
	DeviceName    string `json:"deviceName"`
	DeviceID      string `json:"-"`
	PromotionCode string `json:"promotionCode"`
	Ip            string `json:"-"`
	AppVersion    string `json:"-"`
}

type AdminLoginForm struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CheckPasskeyForm struct {
	Email    string `json:"email" validate:"required"`
	DeviceID string `json:"-"`
}

type AgentLoginForm struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}
