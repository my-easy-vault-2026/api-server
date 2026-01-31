package dao

import (
	accountDao "api-server/dao/account"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(accountDao.NewAssetDao),
	fx.Provide(accountDao.NewBillDao),
	fx.Provide(accountDao.NewAssetTransactionDao),
	fx.Provide(accountDao.NewAssetDailySnapshotDao),
	fx.Provide(accountDao.NewCategoryDao),
	fx.Provide(authDao.NewTokenBucketDao),
)
