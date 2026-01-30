package services

import (
	coinsdoDao "api-server/dao/coinsdo"
	"context"
	"shared-modules/common"
	"shared-modules/logger"
)

type CryptoCurrencyService struct {
	cryptoCurrencyDao *coinsdoDao.CryptoCurrencyDao
}

func NewCryptoCurrencyService() *CryptoCurrencyService {
	return &CryptoCurrencyService{
		cryptoCurrencyDao: coinsdoDao.NewCryptoCurrencyDao(),
	}
}

func (cs *CryptoCurrencyService) GetCryptoCurrencies(ctx context.Context) ([]*coinsdoDao.CryptoCurrency, error) {
	cryptos, err := cs.cryptoCurrencyDao.GetCryptoCurrencies(ctx)

	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed, ", err)
		return nil, err
	}
	return cryptos, nil
}

func (cs *CryptoCurrencyService) ListByType(ctx context.Context, t common.AssetType) ([]*coinsdoDao.CryptoCurrency, error) {
	cryptos, err := cs.cryptoCurrencyDao.ListByType(ctx, t)

	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed, ", err)
		return nil, err
	}
	return cryptos, nil
}

func (cs *CryptoCurrencyService) GetCryptoCurrencyByCurrencyType(ctx context.Context, currency common.Currency) (*coinsdoDao.CryptoCurrency, error) {
	crypto, err := cs.cryptoCurrencyDao.GetCryptoCurrencyByCurrencyType(ctx, currency)

	if err != nil {
		logger.Error(ctx, "cryptoCurrency get failed,", err)
		return nil, err
	}
	return crypto, nil
}

func (cs *CryptoCurrencyService) GetCryptoCurrencyByMainnetCurrency(ctx context.Context, mainnet common.Mainnet, currency common.Currency) (*coinsdoDao.CryptoCurrency, error) {
	crypto, err := cs.cryptoCurrencyDao.GetByMainnetCurrency(ctx, mainnet, currency)

	if err != nil {
		logger.Error(ctx, "cryptoCurrency get failed,", err)
		return nil, err
	}
	return crypto, nil
}
