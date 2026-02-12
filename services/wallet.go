package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	"api-server/infra"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/utils"
	"strconv"

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
		cs.logger.Warn("get failed,", err)
		return []*cardDao.Card{}, cs.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
}

func (cs *WalletService) ListCategory(ctx context.Context) ([]*accountDao.Category, error) {

	categories, err := cs.categoryDao.List(ctx)
	if err != nil {
		cs.logger.Warn("get failed,", err)
		return []*accountDao.Category{}, cs.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return categories, nil
}

func (ws *WalletService) CreateWallet(ctx context.Context, categoryID uint64, userID uint64) (uint64, error) {

	currency := common.Currency(categoryID)
	category, err := ws.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		ws.logger.Warn("get failed", err)
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category == nil {
		return 0, ws.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_CURRENCY)
	}

	locker := ws.lockers.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		ws.env.LockDuration,
		ws.env.LockWaitDuration,
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
