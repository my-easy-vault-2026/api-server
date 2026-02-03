package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	exchangeDao "api-server/dao/exchange"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"api-server/infra"
	"api-server/lib"
	"context"
	"encoding/json"
	"math/rand"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ExchangeService struct {
	exchangeOrderDao     *orderDao.ExchangeOrderDao
	cardDao              *cardDao.CardDao
	assetDao             *accountDao.AssetDao
	parameterDao         *systemDao.ParameterDao
	currencyConfigDao    *coinsdoDao.CurrencyConfigDao
	previewDao           *exchangeDao.PreviewDao
	userDao              *userDao.UserDao
	transactionRecordDao *orderDao.TransactionRecordDao
	assetTransactionDao  *accountDao.AssetTransactionDao
	categoryDao          *accountDao.CategoryDao
	logger               lib.Logger
	beBuilder            *lib.BEBuilder
	lockers              infra.Lockers
	db                   infra.Database
	env                  *lib.Env
	quoteService         *QuoteService
}

func NewExchangeService(
	exchangeOrderDao *orderDao.ExchangeOrderDao,
	cardDao *cardDao.CardDao,
	assetDao *accountDao.AssetDao,
	parameterDao *systemDao.ParameterDao,
	currencyConfigDao *coinsdoDao.CurrencyConfigDao,
	previewDao *exchangeDao.PreviewDao,
	userDao *userDao.UserDao,
	transactionRecordDao *orderDao.TransactionRecordDao,
	assetTransactionDao *accountDao.AssetTransactionDao,
	categoryDao *accountDao.CategoryDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	lockers infra.Lockers,
	db infra.Database,
	env *lib.Env,
	quoteService *QuoteService,
) *ExchangeService {

	return &ExchangeService{
		exchangeOrderDao:     exchangeOrderDao,
		cardDao:              cardDao,
		assetDao:             assetDao,
		parameterDao:         parameterDao,
		currencyConfigDao:    currencyConfigDao,
		previewDao:           previewDao,
		userDao:              userDao,
		transactionRecordDao: transactionRecordDao,
		assetTransactionDao:  assetTransactionDao,
		categoryDao:          categoryDao,
		logger:               logger,
		beBuilder:            beBuilder,
		lockers:              lockers,
		db:                   db,
		env:                  env,
		quoteService:         quoteService,
	}
}

func (es *ExchangeService) ExchangePreview(ctx context.Context,
	fromWalletID uint64,
	toCurrency common.Currency,
	fromAmount decimal.Decimal,
	userID uint64) (preview *exchangeDao.Preview, key string, fromPlaces int, toPlaces int, err error) {

	// 根據form中的FromCardID取得使用者的卡片資訊
	fromWallet, err := es.cardDao.GetByID(ctx, fromWalletID)
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	} else if fromWallet == nil || fromWallet.UserID != userID {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CARD)
		return
	}

	toWallet := (*cardDao.Card)(nil)
	if toCurrency == 0 {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CATEGORY)
		return
	}
	toWallet, err = es.cardDao.GetByUserIDCategoryID(ctx, userID, uint64(toCurrency))
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if toWallet == nil && toCurrency.AssetType() == common.ASSET_TYPE_FIAT {
		walletID, err2 := es.createWallet(ctx, toCurrency, userID)
		if err2 != nil {
			err = err2
			return
		}

		toWallet, err = es.cardDao.GetByID(ctx, walletID)
		if err != nil {
			es.logger.Warn("get failed,", err)
			err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}
	}

	if toWallet == nil || toWallet.UserID != userID {
		es.logger.Warnf("not matched,%d, %d", userID, toWallet.UserID)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}
	if fromWallet.Type != common.ASSET_TYPE_FIAT {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_TARGET)
		return
	}
	if toWallet.Type == common.ASSET_TYPE_FIAT {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_TARGET)
		return
	}

	fromUser, err := es.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if fromUser == nil {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_USER)
		return
	}

	quoteCurrency := toWallet.Currency
	baseCurrency := fromWallet.Currency

	exchangeRate, err := es.quoteService.GetExchangeRates(ctx, common.RATE_PURPOSE_EXCHANGE, quoteCurrency, baseCurrency)
	if err != nil {
		es.logger.Warn("get exchange rate failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	exchangeRate.Purpose = common.RATE_PURPOSE_FEE
	rates := []*common.ExchangeRate{exchangeRate}

	// 計算要交換的金額

	toAmount := fromAmount.Div(exchangeRate.Rate)

	// 檢查使用者資產是否足夠進行交換
	asset, err := es.assetDao.GetByIDUserID(ctx, fromWallet.ID, userID)

	if err != nil {
		logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if asset == nil || asset.Amount.LessThan(fromAmount) {
		es.logger.Warn("insufficient funds")
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INSUFFICIENT_FUNDS)
		return
	}

	param, err := es.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if param == nil {
		logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	exchangeFee := decimal.Decimal{}
	v, err := decimal.NewFromString(param.Value)
	if err != nil {
		logger.Warn("parse failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	switch param.ValueType {
	case common.PARAMETER_VALUE_TYPE_AMOUNT:
		exchangeFee = v
		if fromWallet.Currency != param.Currency {
			feeRates, err2 := es.quoteService.GetExchangeRates(ctx, 0, param.Currency, fromWallet.Currency)
			if err2 != nil {
				logger.Warn("get exchange rate failed,", err2)
				err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
				return
			}
			exchangeFee = v.Mul(feeRates.Rate)
			feeRates.Purpose = common.RATE_PURPOSE_FEE
			rates = append(rates, feeRates)
		}
	case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
		exchangeFee = fromAmount.Mul(v)
	}

	preview = &exchangeDao.Preview{
		UserID:         userID,
		FromWalletID:   fromWallet.ID,
		FromCategoryID: fromWallet.CategoryID,
		FromCurrency:   fromWallet.Currency,
		ToWalletID:     toWallet.ID,
		ToCategoryID:   toWallet.CategoryID,
		ToCurrency:     toWallet.Currency,
		FromAmount:     fromAmount,
		ExchangeFee:    exchangeFee,
		Rates:          rates,
		ExpiredAt:      time.Now().Add(es.env.PreviewExpiryTime * time.Second),
	}

	preview.ToAmount = fromAmount.Sub(preview.ExchangeFee).Div(exchangeRate.Rate)

	// 取得加密貨幣的小數點顯示位數
	currencyConfigs, err := es.currencyConfigDao.ListByCurrencies(ctx, []common.Currency{
		fromWallet.Currency,
		toWallet.Currency,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	for _, conf := range currencyConfigs {
		if conf.CurrencyType == fromWallet.Currency {
			fromPlaces = conf.Decimals
		}
		if conf.CurrencyType == toWallet.Currency {
			toPlaces = conf.Decimals
		}
	}

	preview.FromAmount = preview.FromAmount.Round(int32(fromPlaces))
	preview.ToAmount = preview.ToAmount.Round(int32(toPlaces))
	preview.ExchangeFee = preview.ExchangeFee.Round(int32(fromPlaces))

	data, err := json.Marshal(preview)
	if err != nil {
		logger.Warn("marshal failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	key = utils.Md5String(string(data) + time.Now().String())

	if err2 := es.previewDao.Save(ctx, key, preview, es.env.PreviewExpiryTime*time.Second); err2 != nil {
		logger.Warn("save failed,", err2)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	return
}

func (es *ExchangeService) ExchangeConfirm(ctx context.Context, form *entities.ExchangeConfirmForm, userID uint64) (*string, error) {

	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, form.ExchangeKey),
		utils.Config.Exchange.PreviewExpireSeconds*time.Second,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, form.ExchangeKey), err)
		return nil, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, form.ExchangeKey), err)
		}
	}()

	user, err := es.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if user == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TOP_UP_INCORRECT_PIN_CODE)
	}

	preview, err := es.previewDao.Get(ctx, form.ExchangeKey)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if preview == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TOP_UP_TOP_UP_EXPIRED)
	}

	defer func() {
		if err := es.previewDao.Remove(ctx, form.ExchangeKey); err != nil {
			logger.Warn("delete failed,", err)
		}
	}()

	if fromAsset, err := es.assetDao.GetByIDUserID(ctx, preview.FromCardID, userID); err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if fromAsset == nil {
		logger.Warn("no card asset, ID: %d", preview.FromCardID)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if fromAsset.Amount.LessThan(preview.FromAmount) {
		// TODO: exchange failed
		return nil, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_INSUFFICIENT_FUND)
	}

	fromCard, err := es.cardDao.GetByID(ctx, preview.FromCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromCard == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
	}

	toCard, err := es.cardDao.GetByID(ctx, preview.ToCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if toCard == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNO := "EXG_" + strconv.FormatUint(preview.FromCardID, 10) + "_" + strconv.FormatUint(preview.ToCardID, 10) +
		"_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

	logger.Infof("start exchanging, card ID: [%d] -> [%d], order NO: [%s]", preview.FromCardID, preview.ToCardID, orderNO)

	var tErr error
	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)

		order := &orderDao.ExchangeOrder{
			UserID:         userID,
			OrderNO:        orderNO,
			FromAmount:     preview.FromAmount,
			FromCardID:     preview.FromCardID,
			FromCategoryID: preview.FromCategoryID,
			FromCurrency:   preview.FromCurrency,
			ToAmount:       preview.ToAmount,
			ToCardID:       preview.ToCardID,
			ToCategoryID:   preview.ToCategoryID,
			ToCurrency:     preview.ToCurrency,
			Fee:            preview.ExchangeFee,
			TriggerMode:    common.EXCHANGE_TRIGGER_MODE_ON_REQUEST,
			Status:         common.EXCHANGE_STATUS_SUCCESS,
		}

		for _, rate := range preview.Rate {
			if rate.BaseCurrency == preview.FromCurrency &&
				rate.QuoteCurrency == preview.ToCurrency {
				order.ExchangeRate = rate.Rate
				break
			}
		}

		_, tErr = es.exchangeOrderDao.Save(ctx, order)
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		var rowsAffected int64
		rowsAffected, tErr = es.transactionRecordDao.Saves(ctx, []*orderDao.TransactionRecord{
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_EXCHANGE,
				TransactionNO:           orderNO,
				UserID:                  userID,
				CardID:                  preview.FromCardID,
				Income:                  decimal.NewNullDecimal(preview.FromAmount.Neg()),
				IncomeCategoryID:        preview.FromCategoryID,
				IncomeCurrency:          preview.FromCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.ToAmount),
				AgainstIncomeCategoryID: preview.ToCategoryID,
				AgainstIncomeCurrency:   preview.ToCurrency,
				Side:                    common.TRANSACTION_SIDE_FROM,
				FromCardID:              preview.FromCardID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				FromAlias:               fromCard.Alias,
				ToCardID:                preview.ToCardID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				ToAlias:                 toCard.Alias,
				DisplayRate:             preview.DisplayRate,
				ExchangeRate:            &order.ExchangeRate,
				ExchangeFee:             &preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				ExecutorRole:            common.ROLE_USER,
				Status:                  common.TRANSACTION_STATUS_EXCHANGE_SUCCESS,
			},
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_EXCHANGE,
				TransactionNO:           orderNO,
				UserID:                  userID,
				CardID:                  preview.ToCardID,
				Income:                  decimal.NewNullDecimal(preview.ToAmount),
				IncomeCategoryID:        preview.ToCategoryID,
				IncomeCurrency:          preview.ToCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.FromAmount.Neg()),
				AgainstIncomeCategoryID: preview.FromCategoryID,
				AgainstIncomeCurrency:   preview.FromCurrency,
				Side:                    common.TRANSACTION_SIDE_TO,
				FromCardID:              preview.FromCardID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				ToCardID:                preview.ToCardID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				DisplayRate:             preview.DisplayRate,
				ExchangeRate:            &order.ExchangeRate,
				ExchangeFee:             &preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				ExecutorRole:            common.ROLE_USER,
				Status:                  common.TRANSACTION_STATUS_EXCHANGE_SUCCESS,
			},
		})
		if err != nil {
			logger.Warn("saves failed,", err)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if rowsAffected != 2 {
			logger.Warnf("duplicated save: [%+v]", preview)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		var card *cardDao.Card
		card, tErr = es.cardDao.GetsByFreezeStatusIDForShare(ctx, common.CARD_FREEZE_STATUS_UNFROZEN, order.FromCardID)
		if tErr != nil {
			logger.Warn("get failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if card == nil {
			tErr = utils.NewBusinessError(ctx, common.CODE_EXCHANGE_FROM_CARD_FROZEN)
			return tErr
		}

		card, tErr = es.cardDao.GetsByFreezeStatusIDForShare(ctx, common.CARD_FREEZE_STATUS_UNFROZEN, order.ToCardID)
		if tErr != nil {
			logger.Warn("get failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if card == nil {
			tErr = utils.NewBusinessError(ctx, common.CODE_EXCHANGE_FROM_CARD_FROZEN)
			return tErr
		}

		transactions := []*accountDao.AssetTransaction{
			// 用戶扣幣
			{
				UserID:          userID,
				CardID:          preview.FromCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.FromCategoryID,
				Currency:        preview.FromCurrency,
				Amount:          preview.FromAmount.Sub(preview.ExchangeFee),
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_EXCHANGE_SEND_DEDUCT,
			},
			// 用戶扣手續費
			{
				UserID:          userID,
				CardID:          preview.FromCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.FromCategoryID,
				Currency:        preview.FromCurrency,
				Amount:          preview.ExchangeFee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_EXCHANGE_SEND_FEE_DEDUCT,
			},
			// 用戶加幣
			{
				UserID:          userID,
				CardID:          preview.ToCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.ToCategoryID,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_EXCHANGE_RECEIVE_ADD,
			},
		}

		tErr = es.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_EXCHANGE, false)
		if tErr != nil {
			logger.Warn("transaction failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_ASSET_TRANSACTION_FAILED)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		logger.Warn("transaction failed,", tErr)
		return nil, tErr
	}

	if err != nil {
		logger.Warn("transaction failed,", tErr)
		return nil, err
	}

	logger.Infof("exchanged, card ID: [%d] -> [%d], order NO: [%s]", preview.FromCardID, preview.ToCardID, orderNO)

	return &orderNO, nil
}

func (es *ExchangeService) createWallet(ctx context.Context, currency common.Currency, userID uint64) (walletID uint64, err error) {

	categoryID := uint64(currency)
	if !currency.IsValid() {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CATEGORY)
		return
	}

	category, err := es.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		logger.Warn("get failed", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	locker := es.lockers.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		es.env.LockDuration*time.Microsecond,
		es.env.LockWaitDuration*time.Microsecond,
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
	err = utils.WithTX(es.db.DB, func(tx *gorm.DB) error {
		cardDaoTX := es.cardDao.WithTx(es.db.DB)
		assetDaoTX := es.assetDao.WithTx(es.db.DB)
		var wallet *cardDao.Card
		wallet, tErr = cardDaoTX.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}

		id, tErr = cardDaoTX.Save(ctx, &cardDao.Card{
			UserID:       userID,
			Name:         category.Name,
			Type:         common.AssetType(currency.Type()),
			CategoryID:   categoryID,
			Currency:     currency,
			CurrencyType: currency.Type(),
			Nation:       category.Nation,
			NationCode:   category.NationCode,
			Status:       common.CARD_STATUS_ACTIVATED,
		})
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
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
			logger.Warn("save failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		return 0, tErr
	}

	if err != nil {
		logger.Warn("transaction failed,", err)
		return 0, es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return id, nil
}
