package services

import (
	accountDao "api-server/dao/account"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
)

type AccountService struct {
	assetDao  *accountDao.AssetDao
	logger    lib.Logger
	beBuilder *lib.BEBuilder
}

func NewAccountService(assetDao *accountDao.AssetDao,
	userDao *userDao.UserDao,
	logger lib.Logger,
	BEBuilder *lib.BEBuilder) *AccountService {

	return &AccountService{
		assetDao:  assetDao,
		logger:    logger,
		beBuilder: BEBuilder,
	}
}

func (as *AccountService) ListAssetsByIDInUserID(ctx context.Context, ids []uint64, userID uint64) ([]*accountDao.Asset, error) {

	assets, err := as.assetDao.Gets(ctx, &accountDao.AssetQuery{
		Asset: accountDao.Asset{
			UserID: userID,
		},
		IDIn: ids,
	})
	if err != nil {
		as.logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) GetAssetByIDUserID(ctx context.Context, id uint64, userID uint64) (*accountDao.Asset, error) {

	asset, err := as.assetDao.Get(ctx, &accountDao.AssetQuery{
		Asset: accountDao.Asset{
			UserID: userID,
			ID:     id,
		},
	})
	if err != nil {
		as.logger.Warn("get failed,", err)
		return nil, as.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return asset, nil
}
