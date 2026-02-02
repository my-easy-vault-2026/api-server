package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	"api-server/lib"
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
	assetCategoryDao *accountDao.AssetCategoryDao
	categoryDao      *accountDao.CategoryDao
	assetDao         *accountDao.AssetDao
	cardDao          *cardDao.CardDao
	logger           lib.Logger
	beBuilder        *lib.BEBuilder
}

func NewWalletService(
	assetCategoryDao *accountDao.AssetCategoryDao,
	categoryDao *accountDao.CategoryDao,
	assetDao *accountDao.AssetDao,
	cardDao *cardDao.CardDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
) *WalletService {
	return &WalletService{
		assetCategoryDao: assetCategoryDao,
		categoryDao:      categoryDao,
		assetDao:         assetDao,
		cardDao:          cardDao,
		logger:           logger,
		beBuilder:        beBuilder,
	}
}

func (cs *WalletService) ListWalletsByUserID(ctx context.Context, userID uint64) ([]*cardDao.Card, error) {

	cards, err := cs.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID: userID,
		},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, cs.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
}

func (cs *WalletService) ListCategory(ctx context.Context) ([]*accountDao.Category, error) {

	categories, err := cs.categoryDao.ListByTypeUsage(ctx, 0, []common.CategoryUsage{common.CATEGORY_USAGE_USER_DISPLAY})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Category{}, cs.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return categories, nil
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

func (ws *WalletService) CreateWallet(ctx context.Context, categoryID uint64, userID uint64) (uint64, error) {

	currency = common.Currency(form.CategoryID)
	category, err := ws.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		ws.logger.Warn("get failed", err)
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category == nil {
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
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
	return id, nil
}
