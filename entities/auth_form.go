package entities

type LoginOrCreateForm struct {
	Email   string `json:"email" validate:"required,email"`
	PINCode string `json:"newPinCode" validate:"required,number,len=6"`
}
