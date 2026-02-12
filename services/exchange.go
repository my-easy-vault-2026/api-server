package services

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/my-easy-vault-2026/shared-modules/common"
	"github.com/my-easy-vault-2026/shared-modules/utils"

	accountDao "github.com/my-easy-vault-2026/api-server/dao/account"
	cardDao "github.com/my-easy-vault-2026/api-server/dao/card"
	coinsdoDao "github.com/my-easy-vault-2026/api-server/dao/coinsdo"
	exchangeDao "github.com/my-easy-vault-2026/api-server/dao/exchange"
	orderDao "github.com/my-easy-vault-2026/api-server/dao/order"
	systemDao "github.com/my-easy-vault-2026/api-server/dao/system"
	userDao "github.com/my-easy-vault-2026/api-server/dao/user"
	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"

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
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	toWallet := (*cardDao.Card)(nil)
	if toCurrency == 0 {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_CURRENCY)
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
	if toWallet.Type != common.ASSET_TYPE_FIAT {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_TARGET)
		return
	}
	if fromWallet.ID == toWallet.ID {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_TARGET)
		return
	}

	fromUser, err := es.userDao.GetByUserID(ctx, userID)
	if err != nil {
		es.logger.Warn("get failed,", err)
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
	exchangeRate.Purpose = common.RATE_PURPOSE_EXCHANGE
	rates := []*common.ExchangeRate{exchangeRate}

	// 檢查使用者資產是否足夠進行交換
	asset, err := es.assetDao.GetByIDUserID(ctx, fromWallet.ID, userID)

	if err != nil {
		es.logger.Warn("get failed,", err)
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
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if param == nil {
		es.logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	exchangeFee := decimal.Decimal{}
	v, err := decimal.NewFromString(param.Value)
	if err != nil {
		es.logger.Warn("parse failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	switch param.ValueType {
	case common.PARAMETER_VALUE_TYPE_AMOUNT:
		exchangeFee = v
		if fromWallet.Currency != param.Currency {
			feeRates, err2 := es.quoteService.GetExchangeRates(ctx, 0, param.Currency, fromWallet.Currency)
			if err2 != nil {
				es.logger.Warn("get exchange rate failed,", err2)
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
		ExpiredAt:      time.Now().Add(es.env.PreviewExpiryTime),
	}

	preview.ToAmount = fromAmount.Sub(preview.ExchangeFee).Div(exchangeRate.Rate)

	// 取得加密貨幣的小數點顯示位數
	currencyConfigs, err := es.currencyConfigDao.ListByCurrencies(ctx, []common.Currency{
		fromWallet.Currency,
		toWallet.Currency,
	})
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	for _, conf := range currencyConfigs {
		if conf.Currency == fromWallet.Currency {
			fromPlaces = conf.Decimals
		}
		if conf.Currency == toWallet.Currency {
			toPlaces = conf.Decimals
		}
	}

	preview.FromAmount = preview.FromAmount.RoundFloor(int32(fromPlaces))
	preview.ToAmount = preview.ToAmount.RoundFloor(int32(toPlaces))
	preview.ExchangeFee = preview.ExchangeFee.RoundFloor(int32(fromPlaces))

	data, err := json.Marshal(preview)
	if err != nil {
		es.logger.Warn("marshal failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	key = utils.Md5String(string(data) + time.Now().String())

	if err2 := es.previewDao.Save(ctx, key, preview, es.env.PreviewExpiryTime); err2 != nil {
		es.logger.Warn("save failed,", err2)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	return
}

func (es *ExchangeService) ExchangeConfirm(ctx context.Context, key string, userID uint64) (orderNO string, err error) {

	locker := es.lockers.NewLocker()
	if err = locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, key),
		es.env.PreviewExpiryTime,
		es.env.LockWaitDuration,
	); err != nil {
		es.logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, key), err)
		return
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			es.logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_EXCHANGE_CONFIRM, key), err)
		}
	}()

	user, err := es.userDao.GetByUserID(ctx, userID)
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if user == nil {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	preview, err := es.previewDao.Get(ctx, key)
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if preview == nil {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_PREVIEW_EXPIRED)
		return
	}

	defer func() {
		if err := es.previewDao.Remove(ctx, key); err != nil {
			es.logger.Warn("delete failed,", err)
		}
	}()

	if fromAsset, err2 := es.assetDao.GetByIDUserID(ctx, preview.FromWalletID, userID); err2 != nil {
		es.logger.Warn("get failed,", err2)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	} else if fromAsset == nil {
		es.logger.Warn("no card asset, ID: %d", preview.FromWalletID)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	} else if fromAsset.Amount.LessThan(preview.FromAmount) {
		es.logger.Warn("insufficient funds")
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_INSUFFICIENT_FUNDS)
		return
	}

	fromWallet, err := es.cardDao.GetByID(ctx, preview.FromWalletID)
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if fromWallet == nil {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	toWallet, err := es.cardDao.GetByID(ctx, preview.ToWalletID)
	if err != nil {
		es.logger.Warn("get failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if toWallet == nil {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNO = "EXG_" + strconv.FormatUint(preview.FromWalletID, 10) + "_" + strconv.FormatUint(preview.ToWalletID, 10) +
		"_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

	es.logger.Infof("start exchanging, card ID: [%d] -> [%d], order NO: [%s]", preview.FromWalletID, preview.ToWalletID, orderNO)

	var tErr error
	err = utils.WithTX(es.db.DB, func(tx *gorm.DB) error {
		exchangeOrderDaoTX := es.exchangeOrderDao.WithTx(es.db.DB)
		transactionRecordDaoTX := es.transactionRecordDao.WithTx(es.db.DB)
		assetTransactionDaoTX := es.assetTransactionDao.WithTx(es.db.DB)

		order := &orderDao.ExchangeOrder{
			UserID:         userID,
			OrderNO:        orderNO,
			FromAmount:     preview.FromAmount,
			FromWalletID:   preview.FromWalletID,
			FromCategoryID: preview.FromCategoryID,
			FromCurrency:   preview.FromCurrency,
			ToAmount:       preview.ToAmount,
			ToWalletID:     preview.ToWalletID,
			ToCategoryID:   preview.ToCategoryID,
			ToCurrency:     preview.ToCurrency,
			Fee:            preview.ExchangeFee,
			Status:         common.EXCHANGE_STATUS_SUCCESS,
		}

		for _, rate := range preview.Rates {
			if rate.Purpose == common.RATE_PURPOSE_EXCHANGE {
				order.ExchangeRate = rate.Rate
				break
			}
		}

		var orderID uint64
		orderID, tErr = exchangeOrderDaoTX.Save(ctx, order)
		if tErr != nil {
			es.logger.Warn("save failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if orderID == 0 {
			es.logger.Warn("save failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_DB_INSERT_FAILED)
			return tErr
		}

		var rowsAffected int64
		rowsAffected, tErr = transactionRecordDaoTX.Saves(ctx, []*orderDao.TransactionRecord{
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_EXCHANGE,
				TransactionNO:           orderNO,
				UserID:                  userID,
				WalletID:                preview.FromWalletID,
				Income:                  decimal.NewNullDecimal(preview.FromAmount.Neg()),
				IncomeCategoryID:        preview.FromCategoryID,
				IncomeCurrency:          preview.FromCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.ToAmount),
				AgainstIncomeCategoryID: preview.ToCategoryID,
				AgainstIncomeCurrency:   preview.ToCurrency,
				Side:                    common.TRANSACTION_SIDE_FROM,
				FromWalletID:            preview.FromWalletID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				ToWalletID:              preview.ToWalletID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				ExchangeRate:            &order.ExchangeRate,
				ExchangeFee:             &preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				Status:                  common.TRANSACTION_STATUS_EXCHANGE_SUCCESS,
			},
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_EXCHANGE,
				TransactionNO:           orderNO,
				UserID:                  userID,
				WalletID:                preview.ToWalletID,
				Income:                  decimal.NewNullDecimal(preview.ToAmount),
				IncomeCategoryID:        preview.ToCategoryID,
				IncomeCurrency:          preview.ToCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.FromAmount.Neg()),
				AgainstIncomeCategoryID: preview.FromCategoryID,
				AgainstIncomeCurrency:   preview.FromCurrency,
				Side:                    common.TRANSACTION_SIDE_TO,
				FromWalletID:            preview.FromWalletID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				ToWalletID:              preview.ToWalletID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				ExchangeRate:            &order.ExchangeRate,
				ExchangeFee:             &preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				Status:                  common.TRANSACTION_STATUS_EXCHANGE_SUCCESS,
			},
		})
		if err != nil {
			es.logger.Warn("saves failed,", err)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if rowsAffected != 2 {
			es.logger.Warnf("duplicated save: [%#v]", preview)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_DB_INSERT_FAILED)
			return tErr
		}

		transactions := []*accountDao.AssetTransaction{
			// 用戶扣幣
			{
				UserID:          userID,
				WalletID:        preview.FromWalletID,
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
				WalletID:        preview.FromWalletID,
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
				WalletID:        preview.ToWalletID,
				OrderNO:         orderNO,
				CategoryID:      preview.ToCategoryID,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_EXCHANGE_RECEIVE_ADD,
			},
		}

		tErr = assetTransactionDaoTX.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_EXCHANGE, false)
		if tErr != nil {
			es.logger.Warn("transaction failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_ASSET_TRANSACTION_FAILED)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		es.logger.Warn("transaction failed,", tErr)
		return
	}

	if err != nil {
		es.logger.Warn("transaction failed,", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_ASSET_TRANSACTION_FAILED)
		return
	}

	es.logger.Infof("exchanged, card ID: [%d] -> [%d], order NO: [%s]", preview.FromWalletID, preview.ToWalletID, orderNO)

	return
}

func (es *ExchangeService) createWallet(ctx context.Context, currency common.Currency, userID uint64) (walletID uint64, err error) {

	categoryID := uint64(currency)
	if !currency.IsValid() {
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_CURRENCY)
		return
	}

	category, err := es.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		es.logger.Warn("get failed", err)
		err = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	locker := es.lockers.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		es.env.LockDuration,
		es.env.LockWaitDuration,
	); err != nil {
		es.logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		return 0, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			es.logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
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
			es.logger.Warn("get failed", tErr)
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
			es.logger.Warn("save failed,", tErr)
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
			es.logger.Warn("save failed,", tErr)
			tErr = es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		return 0, tErr
	}

	if err != nil {
		es.logger.Warn("transaction failed,", err)
		return 0, es.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return id, nil
}
