package card

import (
	"context"
	"shared-modules/utils"
)

type DeleteCard struct {
	Card `gorm:"embedded"`
}

type DeleteCardDao struct {
}

func NewDeleteCardDao() *DeleteCardDao {
	return &DeleteCardDao{}
}

func (DeleteCard) TableName() string {
	return "delete_card"
}

func (ad *DeleteCardDao) Save(ctx context.Context, model *DeleteCard) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}
