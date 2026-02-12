package entities

type ListWalletForm struct {
}

type ListWalletCategoryForm struct {
}

type CreateWalletForm struct {
	CategoryID uint64 `json:"categoryId,string"`
}
