package entities

type SignedVO struct {
	Data string `json:"data"`
	Sign string `json:"sign"`
}

type EncryptedVO struct {
	Chunks []string `json:"chunks"`
}
