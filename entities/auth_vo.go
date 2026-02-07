package entities

type LoginOrCreateVO struct {
	ID          uint64 `json:"id,string"`
	Email       string `json:"email"`
	CountryCode uint8  `json:"countryCode,omitempty"` //可能為空
	PhoneNumber string `json:"phoneNumber,omitempty"` //可能為空
	Gender      uint8  `json:"gender,omitempty"`      //可能為空
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
