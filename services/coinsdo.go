package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	walletDao "api-server/dao/wallet"
	"context"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
)

type CoinsdoService struct {
	transactionRecordDao *orderDao.TransactionRecordDao
	walletAddressDao     *walletDao.WalletAddressDao
	parameterDao         *systemDao.ParameterDao
	cryptoCurrencyDao    *coinsdoDao.CryptoCurrencyDao
	assetTransactionDao  *accountDao.AssetTransactionDao
	userDao              *userDao.UserDao
	cardDao              *cardDao.CardDao
	assetDao             *accountDao.AssetDao
	exchangeOrderDao     *orderDao.ExchangeOrderDao
	categoryDao          *accountDao.CategoryDao
	mainCardDao          *cardDao.MainCardDao
}

func NewCoinsdoService() *CoinsdoService {
	return &CoinsdoService{
		transactionRecordDao: orderDao.NewTransactionRecordDao(),
		walletAddressDao:     walletDao.NewWalletAddressDao(),
		parameterDao:         systemDao.NewParameterDao(),
		cryptoCurrencyDao:    coinsdoDao.NewCryptoCurrencyDao(),
		assetTransactionDao:  accountDao.NewAssetTransactionDao(),
		userDao:              userDao.NewUserDao(),
		cardDao:              cardDao.NewCardDao(),
		assetDao:             accountDao.NewAssetDao(),
		categoryDao:          accountDao.NewCategoryDao(),
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

func (cs *CoinsdoService) ListMainnetNames(ctx context.Context) (map[common.Mainnet]string, error) {

	cryptos, err := cs.cryptoCurrencyDao.ListMainnetNames(ctx)

	if err != nil {
		logger.Warn("get failed,", err)
		return make(map[common.Mainnet]string), utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	mainnetNames := make(map[common.Mainnet]string)

	for _, crypto := range cryptos {
		mainnetNames[crypto.Mainnet] = crypto.MainnetFullName
	}

	return mainnetNames, nil
}

func (cs *CoinsdoService) GetCryptoCurrency(ctx context.Context, mainnet string, currency string) (*coinsdoDao.CryptoCurrency, error) {
	cryptoCurrency, err := cs.cryptoCurrencyDao.GetCryptoCurrency(ctx, common.Mainnet(0).FromString(mainnet), currency)
	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return cryptoCurrency, nil
}

func (cs *CoinsdoService) GetCryptoCurrencies(ctx context.Context) ([]*coinsdoDao.CryptoCurrency, error) {
	cryptoCurrencies, err := cs.cryptoCurrencyDao.GetCryptoTypeCurrencies(ctx)
	if err != nil {
		logger.Warn(ctx, "cryptoCurrencies get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return cryptoCurrencies, nil
}

func (cs *CoinsdoService) GetMainnetCaseSensitive(ctx context.Context, mainnet common.Mainnet) (common.CaseSensitive, error) {
	cryptoCurrency, err := cs.cryptoCurrencyDao.GetByMainnet(ctx, mainnet)
	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if cryptoCurrency == nil {
		return 0, utils.NewBusinessError(ctx, common.CODE_NO_SUCH_CURRENCY)
	}

	if cryptoCurrency.CaseSensitive == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_COINSDO_NO_CASE_SENSITIVE_DATA)
	}

	return cryptoCurrency.CaseSensitive, nil
}
