package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	exchangeDao "api-server/dao/exchange"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"encoding/json"
	"math/rand"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ExchangeService struct {
	exchangeOrderDao     *orderDao.ExchangeOrderDao
	cardDao              *cardDao.CardDao
	assetDao             *accountDao.AssetDao
	parameterDao         *systemDao.ParameterDao
	cryptoCurrencyDao    *coinsdoDao.CryptoCurrencyDao
	previewDao           *exchangeDao.PreviewDao
	userDao              *userDao.UserDao
	transactionRecordDao *orderDao.TransactionRecordDao
	assetTransactionDao  *accountDao.AssetTransactionDao
	categoryDao          *accountDao.CategoryDao
	currencyConfigDao    *orderDao.CurrencyConfigDao
}

func NewExchangeService() *ExchangeService {
	return &ExchangeService{
		exchangeOrderDao:     orderDao.NewExchangeOrderDao(),
		cardDao:              cardDao.NewCardDao(),
		assetDao:             accountDao.NewAssetDao(),
		parameterDao:         systemDao.NewParameterDao(),
		cryptoCurrencyDao:    coinsdoDao.NewCryptoCurrencyDao(),
		previewDao:           exchangeDao.NewPreviewDao(),
		userDao:              userDao.NewUserDao(),
		transactionRecordDao: orderDao.NewTransactionRecordDao(),
		assetTransactionDao:  accountDao.NewAssetTransactionDao(),
		categoryDao:          accountDao.NewCategoryDao(),
		currencyConfigDao:    orderDao.NewCurrencyConfigDao(),
	}
}

func (es *ExchangeService) ExchangePreview(ctx context.Context, form *entities.ExchangePreviewForm, userID uint64) (*exchangeDao.Preview, string, int, int, decimal.Decimal, error) {

	// 根據form中的FromCardID取得使用者的卡片資訊
	fromCard, err := es.cardDao.GetByID(ctx, form.FromCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if fromCard == nil || fromCard.UserID != userID {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CARD)
	} else if fromCard.FreezeStatus == common.CARD_FREEZE_STATUS_FROZEN {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_FROM_CARD_FROZEN)
	}

	// 初始化接收目標卡片的變量
	toCard := (*cardDao.Card)(nil)
	if form.ToCardID != 0 {
		// 若指定了ToCardID，則直接根據ID取得卡片
		toCard, err = es.cardDao.GetByID(ctx, form.ToCardID)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	} else if form.ToCategory != "" {
		// 若未指定ToCardID但指定了ToCategory，根據分類取得目標卡片
		var toCurrency common.Currency
		if toCurrency = common.Currency(0).FromString(form.ToCategory); toCurrency == 0 {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_CARD_NO_SUCH_ASSET_CATEGORY)
		}
		// 根據使用者ID和分類取得卡片
		toCard, err = es.cardDao.GetByUserIDCategoryID(ctx, userID, uint64(toCurrency))
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if toCard == nil && toCurrency.AssetType() != common.ASSET_TYPE_CARD_PRODUCT {
			// 如果該分類不屬於卡片產品且無對應卡片，則創建一個新錢包
			walletID, err := es.createWallet(ctx, &entities.CreateWalletForm{
				Currency: form.ToCategory,
			}, userID)
			if err != nil {
				return nil, "", 0, 0, decimal.Decimal{}, err
			}
			// 創建完成後再嘗試取得該錢包卡片
			toCard, err = es.cardDao.GetByID(ctx, walletID)
			if err != nil {
				logger.Warn("get failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
		}
	}

	if toCard == nil || toCard.UserID != userID {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CARD)
	}
	if fromCard.Type == common.ASSET_TYPE_CARD_PRODUCT {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_INVALID_TARGET)
	}
	if toCard.Type == common.ASSET_TYPE_CARD_PRODUCT {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_INVALID_TARGET)
	}
	if toCard.FreezeStatus == common.CARD_FREEZE_STATUS_FROZEN {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_TO_CARD_FROZEN)
	}

	fromUser, err := es.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromUser == nil {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_USER)
	}

	currencyConfigs, err := es.currencyConfigDao.List(ctx)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	maxLimit := (*decimal.Decimal)(nil)
	minLimit := (*decimal.Decimal)(nil)
	limitCurrency := fromCard.Currency

	// 先找全域設定
	for _, c := range currencyConfigs {
		if c.Currency == 0 && c.KycLevel == "" {
			if c.Exchange == common.CURRENCY_CONFIG_STATUS_ACTIVE {
				if c.ExchangeMax != nil && c.ExchangeMax.Valid {
					maxLimit = &c.ExchangeMax.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
				if c.ExchangeMin != nil && c.ExchangeMin.Valid {
					minLimit = &c.ExchangeMin.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
			}
			break
		}
	}
	// 找個別設定
	for _, c := range currencyConfigs {
		if fromCard.Currency == c.Currency && c.KycLevel == "" && c.Mainnet == 0 {
			if c.Exchange == common.CURRENCY_CONFIG_STATUS_ACTIVE {
				if c.ExchangeMax != nil && c.ExchangeMax.Valid {
					maxLimit = &c.ExchangeMax.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
				if c.ExchangeMin != nil && c.ExchangeMin.Valid {
					minLimit = &c.ExchangeMin.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
			}
			break
		}
	}
	// 找個別設定 kyc
	for _, c := range currencyConfigs {
		if fromCard.Currency == c.Currency && c.KycLevel == fromUser.KycLevel && c.Mainnet == 0 {
			if c.Exchange == common.CURRENCY_CONFIG_STATUS_ACTIVE {
				if c.ExchangeMax != nil && c.ExchangeMax.Valid {
					maxLimit = &c.ExchangeMax.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
				if c.ExchangeMin != nil && c.ExchangeMin.Valid {
					minLimit = &c.ExchangeMin.Decimal
					limitCurrency = c.ExchangeLimitCurrency
				}
			}
			break
		}
	}

	toBuyPrice := decimal.NewFromInt(1)
	limitBuyPrice := decimal.NewFromInt(1)
	quoteCurrencies := make([]common.Currency, 0, 2)
	if fromCard.Currency != toCard.Currency {
		quoteCurrencies = append(quoteCurrencies, toCard.Currency)
	}
	if fromCard.Currency != limitCurrency {
		quoteCurrencies = append(quoteCurrencies, limitCurrency)
	}

	rates := make([]*utils.ExchangeRate, 0, 2)
	if len(quoteCurrencies) > 0 {
		rates, err = utils.ListExchangeRate(ctx, fromCard.Currency, quoteCurrencies)
		if err != nil {
			logger.Warn("get exchange rate failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		for _, rate := range rates {
			if fromCard.Currency != toCard.Currency && rate.QuoteCurrency == toCard.Currency && rate.Purpose == 0 {
				toBuyPrice = rate.Rate
				rate.Purpose = common.RATE_PURPOSE_EXCHANGE
				break
			}
		}
		for _, rate := range rates {
			if fromCard.Currency != limitCurrency && rate.QuoteCurrency == limitCurrency && rate.Purpose == 0 {
				limitBuyPrice = rate.Rate
				rate.Purpose = common.RATE_PURPOSE_LIMIT
				break
			}
		}
	}

	// 計算要交換的金額
	var toAmount, fromAmount decimal.Decimal
	if form.FromAmount != nil && !form.FromAmount.Equal(decimal.Zero) {
		fromAmount = form.FromAmount.Copy()
		toAmount = fromAmount.Div(toBuyPrice)
	} else if form.ToAmount != nil && !form.ToAmount.Equal(decimal.Zero) {
		toAmount = form.ToAmount.Copy()
		fromAmount = toAmount.Mul(toBuyPrice)
	} else {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	// 檢查使用者資產是否足夠進行交換
	asset, err := es.assetDao.GetByIDUserID(ctx, fromCard.ID, userID)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if asset == nil || asset.Amount.LessThan(fromAmount) {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_CARD_INSUFFICIENT_FUNDS)
	}

	// 取得 最新匯率
	param, err := es.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if param == nil {
		logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	// 解析交換費率
	exchangeFee := decimal.Decimal{}
	v, err := decimal.NewFromString(param.Value)
	if err != nil {
		logger.Warn("parse failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	switch param.ValueType {
	case common.PARAMETER_VALUE_TYPE_AMOUNT:
		exchangeFee = v
		if fromCard.Currency != param.Currency {
			feeRates, err := utils.ListExchangeRate(ctx, fromCard.Currency, []common.Currency{param.Currency})
			if err != nil {
				logger.Warn("get exchange rate failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			exchangeFee = v.Mul(feeRates[0].Rate)
		}
	case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
		exchangeFee = fromAmount.Mul(v)
	}

	if minLimit != nil {
		if fromAmount.LessThan(minLimit.Mul(limitBuyPrice).Sub(exchangeFee)) {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_AMOUNT_TOO_LOW)
		}
	}

	if maxLimit != nil {
		if fromAmount.GreaterThan(maxLimit.Mul(limitBuyPrice).Sub(exchangeFee)) {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_AMOUNT_TOO_HIGH)
		}
	}

	preview := &exchangeDao.Preview{
		UserID:         userID,
		FromCardID:     fromCard.ID,
		FromCategoryID: fromCard.CategoryID,
		FromCurrency:   fromCard.Currency,
		ToCardID:       toCard.ID,
		ToCategoryID:   toCard.CategoryID,
		ToCurrency:     toCard.Currency,
		FromAmount:     fromAmount,
		ExchangeFee:    exchangeFee,
		Rate:           make([]*common.ExchangeRate, 0, 10),
		ExpiredAt:      time.Now().Add(utils.Config.Exchange.PreviewExpireSeconds * time.Second),
	}

	// 計算最終的交換金額，扣除交換費
	preview.ToAmount = fromAmount.Sub(preview.ExchangeFee).Div(toBuyPrice)

	// 加入匯率詳細資訊
	for _, rate := range rates {
		previewRate := &common.ExchangeRate{}
		if err := copier.Copy(previewRate, rate); err != nil {
			logger.Warn("copy failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		preview.Rate = append(preview.Rate, previewRate)
	}

	// 取得加密貨幣的小數點顯示位數
	cryptos, err := es.cryptoCurrencyDao.ListByCurrencies(ctx, []common.Currency{
		fromCard.Currency,
		toCard.Currency,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var fromPlaces, toPlaces int
	for _, crypto := range cryptos {
		if crypto.CurrencyType == fromCard.Currency {
			fromPlaces = crypto.DisplayDecimals
		}
		if crypto.CurrencyType == toCard.Currency {
			toPlaces = crypto.DisplayDecimals
		}
	}

	preview.FromAmount = preview.FromAmount.Round(int32(fromPlaces))
	inverseRate := preview.ToAmount.Round(int32(toPlaces)).Div(preview.FromAmount).Round(int32(utils.Config.System.RatePrecision))
	preview.DisplayRate = &inverseRate
	preview.ToAmount = preview.FromAmount.Mul(inverseRate).Round(int32(toPlaces))
	preview.ExchangeFee = preview.ExchangeFee.Round(int32(fromPlaces))

	data, err := json.Marshal(preview)
	if err != nil {
		logger.Warn("marshal failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	key := utils.Md5String(string(data) + time.Now().String())

	if err := es.previewDao.Save(ctx, key, preview, utils.Config.Exchange.PreviewExpireSeconds*time.Second); err != nil {
		logger.Warn("save failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return preview, key, fromPlaces, toPlaces, inverseRate, nil
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
			// 平台帳戶加幣
			{
				UserID:          common.SYSTEM_ACCOUNT_PLATFORM,
				OrderNO:         orderNO,
				Currency:        preview.FromCurrency,
				Amount:          preview.FromAmount.Sub(preview.ExchangeFee),
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_EXCHANGE_SEND_ADD,
			},
			// 手續費帳戶加手續費
			{
				UserID:          common.SYSTEM_ACCOUNT_FEE,
				OrderNO:         orderNO,
				Currency:        preview.FromCurrency,
				Amount:          preview.ExchangeFee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_EXCHANGE_SEND_FEE_ADD,
			},
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
			// 平台扣幣
			{
				UserID:          common.SYSTEM_ACCOUNT_PLATFORM,
				OrderNO:         orderNO,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_EXCHANGE_RECEIVE_DEDUCT,
			},
		}

		tErr = es.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_EXCHANGE, false)
		if tErr != nil {
			logger.Warn("transaction failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
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

func (es *ExchangeService) createWallet(ctx context.Context, form *entities.CreateWalletForm, userID uint64) (uint64, error) {

	categoryID := uint64(0)
	currency := common.Currency(0)
	if form.CategoryID != 0 {
		currency = common.Currency(form.CategoryID)
		categoryID = uint64(form.CategoryID)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CATEGORY)
		}
	} else if form.Currency != "" {
		currency = common.Currency(0).FromString(form.Currency)
		categoryID = uint64(currency)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_NO_SUCH_CURRENCY)
		}
	}

	// 法幣、e卡不自動開卡
	if categoryID >= 200 {
		return 0, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_INVALID_CATEGORY)
	}

	category, err := es.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		logger.Warn("get failed", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category.Usage&common.CATEGORY_USAGE_USER_APPLY == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_EXCHANGE_INVALID_CATEGORY)
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
		wallet, tErr = es.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}

		id, tErr = es.cardDao.Save(ctx, &cardDao.Card{
			UserID:        userID,
			Type:          common.AssetType(currency.Type()),
			ProductName:   currency.String(),
			CategoryID:    categoryID,
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

		if _, tErr = es.assetDao.Save(ctx, &accountDao.Asset{
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
