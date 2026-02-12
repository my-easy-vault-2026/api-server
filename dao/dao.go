package dao

import (
	accountDao "github.com/my-easy-vault-2026/api-server/dao/account"
	authDao "github.com/my-easy-vault-2026/api-server/dao/auth"
	cardDao "github.com/my-easy-vault-2026/api-server/dao/card"
	coinsdoDao "github.com/my-easy-vault-2026/api-server/dao/coinsdo"
	exchangeDao "github.com/my-easy-vault-2026/api-server/dao/exchange"
	orderDao "github.com/my-easy-vault-2026/api-server/dao/order"
	systemDao "github.com/my-easy-vault-2026/api-server/dao/system"
	transferDao "github.com/my-easy-vault-2026/api-server/dao/transfer"
	userDao "github.com/my-easy-vault-2026/api-server/dao/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(accountDao.NewAssetDao),
	fx.Provide(accountDao.NewBillDao),
	fx.Provide(accountDao.NewAssetTransactionDao),
	fx.Provide(accountDao.NewCategoryDao),
	fx.Provide(authDao.NewTokenBucketDao),
	fx.Provide(authDao.NewTokenDao),
	fx.Provide(authDao.NewAPIAuthorityDao),
	fx.Provide(cardDao.NewCardDao),
	fx.Provide(coinsdoDao.NewCurrencyConfigDao),
	fx.Provide(exchangeDao.NewPreviewDao),
	fx.Provide(orderDao.NewTransactionRecordDao),
	fx.Provide(orderDao.NewExchangeOrderDao),
	fx.Provide(orderDao.NewPreviewDao),
	fx.Provide(orderDao.NewTransferOrderDao),
	fx.Provide(systemDao.NewParameterDao),
	fx.Provide(userDao.NewUserDao),
	fx.Provide(userDao.NewUserGroupDao),
	fx.Provide(transferDao.NewPreviewDao),
)
