package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	systemDao "api-server/dao/system"
	walletDao "api-server/dao/wallet"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type WalletService struct {
	assetCategoryDao  *accountDao.AssetCategoryDao
	categoryDao       *accountDao.CategoryDao
	assetDao          *accountDao.AssetDao
	cardDao           *cardDao.CardDao
	mainCardDao       *cardDao.MainCardDao
	walletAddressDao  *walletDao.WalletAddressDao
	cryptoCurrencyDao *coinsdoDao.CryptoCurrencyDao
	parameterDao      *systemDao.ParameterDao
}

func NewWalletService() *WalletService {

	return &WalletService{
		assetCategoryDao:  accountDao.NewAssetCategoryDao(),
		categoryDao:       accountDao.NewCategoryDao(),
		assetDao:          accountDao.NewAssetDao(),
		cardDao:           cardDao.NewCardDao(),
		mainCardDao:       cardDao.NewMainCardDao(),
		walletAddressDao:  walletDao.NewWalletAddressDao(),
		cryptoCurrencyDao: coinsdoDao.NewCryptoCurrencyDao(),
		parameterDao:      systemDao.NewParameterDao(),
	}
}

func (ws *WalletService) GetChangellyDefaultUsdtMainet(ctx context.Context, currency common.Currency) (string, string, error) {
	mainnetParameter, err := ws.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_CHANGELLY_USDT_DEFAULT_MAINNET)
	if err != nil {
		logger.Warn("get failed,", err)
		return "", "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	mainnet := mainnetParameter.Value
	crypto, err := ws.cryptoCurrencyDao.GetByMainnetCurrency(ctx, common.Mainnet(0).FromString(mainnet), currency)
	if err != nil {
		logger.Warn("get failed,", err)
		return "", "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if crypto == nil {
		return "", "", utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CURRENCY)
	}
	protocol := crypto.Protocol.String()
	return mainnet, protocol, nil
}

func (ws *WalletService) ListWallets(ctx context.Context, form *entities.ListWalletsForm, userID uint64) ([]*cardDao.Card, error) {

	wallets, err := ws.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID: userID,
		},
		TypeIn: []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return wallets, nil
}

func (ws *WalletService) ListWalletAddresses(ctx context.Context, form *entities.ListWalletAddressesForm, userID uint64) ([]*walletDao.WalletAddress, error) {

	walletAddresses, err := ws.walletAddressDao.Gets(ctx, &walletDao.WalletAddressQuery{
		WalletAddress: walletDao.WalletAddress{
			UserID:   userID,
			Mainnet:  common.Mainnet(0).FromString(form.Mainnet),
			Protocol: common.Protocol(0).FromString(form.Protocol),
			Currency: common.Currency(0).FromString(form.Currency),
		},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*walletDao.WalletAddress{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return walletAddresses, nil
}

func (ws *WalletService) CreateWallet(ctx context.Context, form *entities.CreateWalletForm, userID uint64) (uint64, error) {
	categoryID := uint64(0)
	currency := common.Currency(0)
	if form.CategoryID != 0 {
		currency = common.Currency(form.CategoryID)
		categoryID = uint64(form.CategoryID)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
		}
	} else if form.Currency != "" {
		currency = common.Currency(0).FromString(form.Currency)
		categoryID = uint64(currency)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CURRENCY)
		}
	}
	if categoryID >= 200 { // TODO: 目前不支援法幣錢包
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_INVALID_CATEGORY)
	}
	category, err := ws.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		logger.Warn("get failed", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category == nil {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
	}
	if category.Usage&common.CATEGORY_USAGE_USER_APPLY == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_INVALID_CATEGORY)
	}
	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		utils.Config.System.LockMicroseconds*time.Microsecond,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		return 0, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		}
	}()
	var id uint64
	var tErr error
	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)
		var wallet *cardDao.Card
		wallet, tErr = ws.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}
		id, tErr = ws.cardDao.Save(ctx, &cardDao.Card{
			UserID:        userID,
			Type:          common.AssetType(currency.Type()),
			ProductName:   currency.String(),
			CategoryID:    categoryID,
			Alias:         currency.String(),
			Currency:      currency,
			CurrencyType:  currency.Type(),
			FromAutoTopUp: common.AUTO_TOP_UP_STATUS_ENABLED,
			ToAutoTopUp:   common.AUTO_TOP_UP_STATUS_DISABLED,
			Status:        common.CARD_STATUS_ACTIVATED,
			FreezeStatus:  common.CARD_FREEZE_STATUS_UNFROZEN,
		})
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if _, tErr = ws.assetDao.Save(ctx, &accountDao.Asset{
			ID:           id,
			UserID:       userID,
			Type:         currency.AssetType(),
			CategoryID:   categoryID,
			Currency:     currency,
			CurrencyType: currency.Type(),
		}); tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		return nil
	})
	if tErr != nil {
		return 0, tErr
	}
	if err != nil {
		logger.Warn("transaction failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	var mainCards []*cardDao.MainCard
	mainCards, tErr = ws.mainCardDao.ListByUserID(ctx, userID)
	if tErr != nil {
		logger.Warn("get failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	var hasCurrencyMC, hasCategoryIDMC bool
	for _, c := range mainCards {
		if c.Currency == category.Currency && c.CategoryID == 0 {
			hasCurrencyMC = true
		}
		if c.CategoryID == category.ID {
			hasCategoryIDMC = true
		}
	}
	if !hasCurrencyMC {
		_, tErr = ws.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:   userID,
			Currency: category.Currency,
			CardID:   id,
		})
		if tErr != nil {
			logger.Warn("save failed,", err) // 主卡儲存失敗就算了
		}
	}
	if !hasCategoryIDMC {
		_, tErr = ws.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:     userID,
			Currency:   category.Currency,
			CategoryID: category.ID,
			CardID:     id,
		})
		if tErr != nil {
			logger.Warn("save failed,", err) // 主卡儲存失敗就算了
		}
	}
	return id, nil
}

func (ws *WalletService) CreatePoint(ctx context.Context, form *entities.CreatePointForm, userID uint64) (uint64, error) {

	currency := common.Currency(form.CategoryID)
	if !currency.IsValid() {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
	}

	if form.CategoryID < 1000 || form.CategoryID >= 2000 {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_INVALID_CATEGORY)
	}

	category, err := ws.categoryDao.GetByID(ctx, form.CategoryID)
	if err != nil {
		logger.Warn("get failed", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category == nil {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
	}
	if category.Usage&common.CATEGORY_USAGE_USER_REWARD == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_WALLET_INVALID_CATEGORY)
	}

	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_POINT, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)),
		utils.Config.System.LockMicroseconds*time.Microsecond,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_POINT, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)), err)
		return 0, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_POINT, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)), err)
		}
	}()

	var id uint64

	var tErr error
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)

		var wallet *cardDao.Card
		wallet, tErr = ws.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}

		id, tErr = ws.cardDao.Save(ctx, &cardDao.Card{
			UserID:        userID,
			Type:          common.AssetType(currency.Type()),
			ProductName:   currency.String(),
			CategoryID:    form.CategoryID,
			Alias:         currency.String(),
			Currency:      currency,
			CurrencyType:  currency.Type(),
			FromAutoTopUp: common.AUTO_TOP_UP_STATUS_DISABLED,
			ToAutoTopUp:   common.AUTO_TOP_UP_STATUS_DISABLED,
			Status:        common.CARD_STATUS_ACTIVATED,
			FreezeStatus:  common.CARD_FREEZE_STATUS_UNFROZEN,
		})
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		if _, tErr = ws.assetDao.Save(ctx, &accountDao.Asset{
			ID:           id,
			UserID:       userID,
			Type:         currency.AssetType(),
			CategoryID:   form.CategoryID,
			Currency:     currency,
			CurrencyType: currency.Type(),
		}); tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		return 0, tErr
	}

	if err != nil {
		logger.Warn("transaction failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var mainCards []*cardDao.MainCard
	mainCards, tErr = ws.mainCardDao.ListByUserID(ctx, userID)
	if tErr != nil {
		logger.Warn("get failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	var hasCurrencyMC, hasCategoryIDMC bool
	for _, c := range mainCards {
		if c.Currency == category.Currency && c.CategoryID == 0 {
			hasCurrencyMC = true
		}
		if c.CategoryID == category.ID {
			hasCategoryIDMC = true
		}
	}

	if !hasCurrencyMC {
		_, tErr = ws.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:   userID,
			Currency: category.Currency,
			CardID:   id,
		})
		if tErr != nil {
			logger.Warn("save failed,", err) // 主卡儲存失敗就算了
		}
	}

	if !hasCategoryIDMC {
		_, tErr = ws.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:     userID,
			Currency:   category.Currency,
			CategoryID: category.ID,
			CardID:     id,
		})
		if tErr != nil {
			logger.Warn("save failed,", err) // 主卡儲存失敗就算了
		}
	}

	return id, nil
}
