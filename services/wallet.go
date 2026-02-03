package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	"api-server/infra"
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
	categoryDao *accountDao.CategoryDao
	assetDao    *accountDao.AssetDao
	cardDao     *cardDao.CardDao
	logger      lib.Logger
	beBuilder   *lib.BEBuilder
	lockers     infra.Lockers
	env         *lib.Env
	db          infra.Database
}

func NewWalletService(
	categoryDao *accountDao.CategoryDao,
	assetDao *accountDao.AssetDao,
	cardDao *cardDao.CardDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	lockers infra.Lockers,
	env *lib.Env,
	db infra.Database,
) *WalletService {
	return &WalletService{
		categoryDao: categoryDao,
		assetDao:    assetDao,
		cardDao:     cardDao,
		logger:      logger,
		beBuilder:   beBuilder,
		lockers:     lockers,
		env:         env,
		db:          db,
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
func (ws *WalletService) ListWalletByUserIDWalletID(ctx context.Context, userID uint64, walletIDs []uint64) ([]*cardDao.Card, error) {

	cards, err := ws.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID: userID,
		},
		IDIn: walletIDs,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
}

func (ws *WalletService) CreateWallet(ctx context.Context, categoryID uint64, userID uint64) (uint64, error) {

	currency := common.Currency(categoryID)
	category, err := ws.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		ws.logger.Warn("get failed", err)
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category == nil {
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_WALLET_NO_SUCH_CATEGORY)
	}

	locker := ws.lockers.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		ws.env.LockDuration*time.Microsecond,
		ws.env.LockWaitDuration*time.Microsecond,
	); err != nil {
		ws.logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		return 0, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			ws.logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		}
	}()
	var id uint64
	var tErr error
	err = utils.WithTX(ws.db.DB, func(tx *gorm.DB) error {
		cardDaoTX := ws.cardDao.WithTx(ws.db.DB)
		assetDaoTX := ws.assetDao.WithTx(ws.db.DB)
		var wallet *cardDao.Card
		wallet, tErr = ws.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, categoryID)
		if tErr != nil {
			ws.logger.Warn("get failed", tErr)
			tErr = ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}
		id, tErr = cardDaoTX.Save(ctx, &cardDao.Card{
			UserID:       userID,
			Type:         common.AssetType(currency.Type()),
			CategoryID:   categoryID,
			Nation:       category.Nation,
			NationCode:   category.NationCode,
			Currency:     currency,
			CurrencyType: currency.Type(),
			Status:       common.CARD_STATUS_ACTIVATED,
		})
		if tErr != nil {
			ws.logger.Warn("save failed,", tErr)
			tErr = ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if _, tErr = assetDaoTX.Save(ctx, &accountDao.Asset{
			ID:           id,
			UserID:       userID,
			Type:         currency.AssetType(),
			CategoryID:   categoryID,
			Currency:     currency,
			CurrencyType: currency.Type(),
		}); tErr != nil {
			ws.logger.Warn("save failed,", tErr)
			tErr = ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		return nil
	})
	if tErr != nil {
		ws.logger.Warn("transaction failed,", tErr)
		return 0, tErr
	}
	if err != nil {
		ws.logger.Warn("transaction failed,", err)
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return id, nil
}
