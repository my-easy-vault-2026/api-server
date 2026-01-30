package services

import (
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
)

type CoinsdoService struct {
	cryptoCurrencyDao *coinsdoDao.CryptoCurrencyDao
	userDao           *userDao.UserDao
	cardDao           *cardDao.CardDao
	logger            lib.Logger
}

func NewCoinsdoService(
	cryptoCurrencyDao *coinsdoDao.CryptoCurrencyDao,
	userDao *userDao.UserDao,
	cardDao *cardDao.CardDao,
	logger lib.Logger,
) *CoinsdoService {
	return &CoinsdoService{
		cryptoCurrencyDao: cryptoCurrencyDao,
		userDao:           userDao,
		cardDao:           cardDao,
		logger:            logger,
	}
}

func (cs *CoinsdoService) ListDisplayDecimals(ctx context.Context) (map[common.Currency]int, error) {
	currencies, err := cs.cryptoCurrencyDao.ListDisplayDecimalsByDistCurrencies(ctx)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	decimals := make(map[common.Currency]int)

	for _, currency := range currencies {
		decimals[currency.CurrencyType] = currency.DisplayDecimals
	}

	return decimals, nil
}

func (cs *CoinsdoService) GetCryptoCurrency(ctx context.Context, mainnet string, currency string) (*coinsdoDao.CryptoCurrency, error) {
	cryptoCurrency, err := cs.cryptoCurrencyDao.GetCryptoCurrency(ctx, common.Mainnet(0).FromString(mainnet), currency)
	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return cryptoCurrency, nil
}
