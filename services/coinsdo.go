package services

import (
	"context"

	"github.com/my-easy-vault-2026/shared-modules/common"

	cardDao "github.com/my-easy-vault-2026/api-server/dao/card"
	coinsdoDao "github.com/my-easy-vault-2026/api-server/dao/coinsdo"
	userDao "github.com/my-easy-vault-2026/api-server/dao/user"
	"github.com/my-easy-vault-2026/api-server/lib"
)

type CoinsdoService struct {
	currencyConfigDao *coinsdoDao.CurrencyConfigDao
	userDao           *userDao.UserDao
	cardDao           *cardDao.CardDao
	logger            lib.Logger
	beBuilder         *lib.BEBuilder
}

func NewCoinsdoService(
	currencyConfigDao *coinsdoDao.CurrencyConfigDao,
	userDao *userDao.UserDao,
	cardDao *cardDao.CardDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
) *CoinsdoService {
	return &CoinsdoService{
		currencyConfigDao: currencyConfigDao,
		userDao:           userDao,
		cardDao:           cardDao,
		logger:            logger,
		beBuilder:         beBuilder,
	}
}

func (cs *CoinsdoService) ListDisplayDecimals(ctx context.Context) (map[common.Currency]int, error) {
	currencies, err := cs.currencyConfigDao.ListDisplayDecimalsByDistCurrencies(ctx)

	if err != nil {
		cs.logger.Warn("get failed,", err)
		return nil, cs.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	decimals := make(map[common.Currency]int)

	for _, currency := range currencies {
		decimals[currency.Currency] = currency.Decimals
	}

	return decimals, nil
}
