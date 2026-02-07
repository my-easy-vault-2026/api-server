package dao

import (
	accountDao "api-server/dao/account"
	authDao "api-server/dao/auth"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(accountDao.NewAssetDao),
	fx.Provide(accountDao.NewBillDao),
	fx.Provide(accountDao.NewAssetTransactionDao),
	fx.Provide(accountDao.NewCategoryDao),
	fx.Provide(authDao.NewTokenBucketDao),
	fx.Provide(authDao.NewTokenDao),
	fx.Provide(coinsdoDao.NewCurrencyConfigDao),
	fx.Provide(orderDao.NewTransactionRecordDao),
	fx.Provide(orderDao.NewExchangeOrderDao),
	fx.Provide(orderDao.NewPreviewDao),
	fx.Provide(orderDao.NewTransferOrderDao),
	fx.Provide(systemDao.NewParameterDao),
	fx.Provide(userDao.NewUserDao),
	fx.Provide(userDao.NewUserGroupDao),
	fx.Provide(cardDao.NewCardDao),
)
