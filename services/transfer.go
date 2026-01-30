package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	transferDao "api-server/dao/transfer"
	userDao "api-server/dao/user"
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

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ITransferConfirm interface {
	TransferConfirm(ctx context.Context, form *entities.TransferConfirmForm, userID uint64) (*string, error)
}

type TransferService struct {
	transferOrderDao     *orderDao.TransferOrderDao
	cardDao              *cardDao.CardDao
	assetDao             *accountDao.AssetDao
	parameterDao         *systemDao.ParameterDao
	cryptoCurrencyDao    *coinsdoDao.CryptoCurrencyDao
	previewDao           *transferDao.PreviewDao
	userDao              *userDao.UserDao
	transactionRecordDao *orderDao.TransactionRecordDao
	assetTransactionDao  *accountDao.AssetTransactionDao
	categoryDao          *accountDao.CategoryDao
	logger               lib.Logger
}

var _ ITransferConfirm = (*TransferService)(nil)

func NewTransferService() *TransferService {
	return &TransferService{
		transferOrderDao:     orderDao.NewTransferOrderDao(),
		cardDao:              cardDao.NewCardDao(),
		assetDao:             accountDao.NewAssetDao(),
		parameterDao:         systemDao.NewParameterDao(),
		cryptoCurrencyDao:    coinsdoDao.NewCryptoCurrencyDao(),
		previewDao:           transferDao.NewPreviewDao(),
		userDao:              userDao.NewUserDao(),
		transactionRecordDao: orderDao.NewTransactionRecordDao(),
		assetTransactionDao:  accountDao.NewAssetTransactionDao(),
		categoryDao:          accountDao.NewCategoryDao(),
	}
}

func (ts *TransferService) TransferPreview(ctx context.Context, form *entities.TransferPreviewForm, userID uint64) (*transferDao.Preview, string, int, int, decimal.Decimal, error) {
	fromUser, err := ts.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromUser == nil {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_USER)
	}
	if fromUser.Role != common.ROLE_MERCHANT && fromUser.Role != common.ROLE_MERCHANT_USER {
		if fromUser.BlockReason != nil && *fromUser.BlockReason != common.USER_STATUS_EUSD_REAP_KYC && fromUser.BlockStatus == common.USER_BLOCK_STATUS_BLOCK {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_FROM_USER_BLOCKED)
		}
	}

	fromCard, err := ts.cardDao.GetByID(ctx, form.FromCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromCard == nil {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
	}
	if fromCard.UserID != userID {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_NO_PERMISSION)
	}
	if fromCard.Status == common.CARD_STATUS_BLOCKED {
		if func() bool {
			if fromCard.Type == common.ASSET_TYPE_CRYPTO && fromCard.BlockReason != nil && *fromCard.BlockReason == common.CARD_STATUS_EUSD_REAP_KYC {
				logger.Warnf("skip block card. id: %v, type: %v, reason: %v", fromCard.ID, fromCard.Type, *fromCard.BlockReason)
				return false
			}
			return true
		}() {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_FROM_CARD_BLOCKED)
		}
	}
	if fromCard.FreezeStatus == common.CARD_FREEZE_STATUS_FROZEN {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_FROM_CARD_FROZEN)
	}

	// 因為有個需求是eth主幣要帶協議erc-20，實際上主幣是沒有協議的，所以這邊要手動清掉
	if form.Mainnet == common.MAINNET_ETH.String() && fromCard.Currency == common.CURRENCY_ETH {
		form.Protocol = ""
	}

	var toCard *cardDao.Card
	var fromAddress, fromEmail string
	var channel common.TransferChannel
	if form.ToCardID != 0 {
		toCard, err = ts.cardDao.GetByID(ctx, form.ToCardID)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if toCard == nil {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
		}
		channel = common.TRANSFER_CHANNEL_CARD_ID
	} else if form.ToUserID != 0 || form.ToEmail != "" {
		var toUser *userDao.User
		if form.ToUserID != 0 {
			toUser, err = ts.userDao.GetByUserID(ctx, form.ToUserID)
			channel = common.TRANSFER_CHANNEL_USER_ID
		} else if form.ToEmail != "" {
			toUser, err = ts.userDao.GetByEmailRole(ctx, form.ToEmail, common.ROLE_USER)
			fromEmail = fromUser.Email
			channel = common.TRANSFER_CHANNEL_EMAIL
		}
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if toUser == nil {
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_USER)
		}

		toCard, err = ts.cardDao.GetByUserIDCurrencyTypes(ctx, toUser.ID, fromCard.Currency, []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT})
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if toCard == nil {
			walletID, err := ts.createWallet(ctx, &entities.CreateWalletForm{
				CategoryID: uint64(fromCard.Currency),
				Currency:   fromCard.Currency.String(),
			}, toUser.ID)
			if err != nil {
				logger.Warn("create failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			toCard, err = ts.cardDao.GetByID(ctx, walletID)
			if err != nil {
				logger.Warn("get failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
		}

	} else {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_INVALID_TARGET)
	}
	if toCard.Status == common.CARD_STATUS_BLOCKED && toCard.Type != common.ASSET_TYPE_CRYPTO {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_TO_CARD_BLOCKED)
	}
	if toCard.FreezeStatus == common.CARD_FREEZE_STATUS_FROZEN {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_TO_CARD_FROZEN)
	}
	if fromCard.UserID == toCard.UserID {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_SELF_TRANSFER)
	}

	toUser, err := ts.userDao.GetByUserID(ctx, toCard.UserID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if toUser.Role == common.ROLE_MERCHANT || toUser.Role == common.ROLE_MERCHANT_USER {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_TRANSFER_TO_MERCHANT_CARD)
	}
	if toUser.BlockStatus == common.USER_BLOCK_STATUS_BLOCK {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_TO_USER_BLOCKED)
	}

	toBuyPrice := decimal.NewFromInt(1)
	quoteCurrencies := make([]common.Currency, 0, 2)
	if fromCard.Currency != toCard.Currency {
		quoteCurrencies = append(quoteCurrencies, toCard.Currency)
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
				rate.Purpose = common.RATE_PURPOSE_TRANSFER
				break
			}
		}
	}

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

	asset, err := ts.assetDao.GetByIDUserID(ctx, fromCard.ID, userID)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if asset == nil || asset.Amount.LessThan(fromAmount) {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_INSUFFICIENT_FUND)
	}

	params, err := ts.parameterDao.ListByKeys(ctx, []common.ParameterKey{
		common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE,
		common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE,
		common.PARAMETER_KEY_TRANSFER_TRANSFER_EXCHANGE_FEE,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if len(params) < 2 {
		logger.Warnf("no parameter: [%s][%s]", common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	transferFee, exchangeFee := decimal.Zero, (*decimal.Decimal)(nil)
	for _, param := range params {
		if param.Key == common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE {
			v, err := decimal.NewFromString(param.Value)
			if err != nil {
				logger.Warn("parse failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			switch param.ValueType {
			case common.PARAMETER_VALUE_TYPE_AMOUNT:
				transferFee = v
				if fromCard.Currency != param.Currency {
					feeRates, err := utils.ListExchangeRate(ctx, fromCard.Currency, []common.Currency{param.Currency})
					if err != nil {
						logger.Warn("get exchange rate failed,", err)
						return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
					}
					transferFee = v.Mul(feeRates[0].Rate)
				}
			case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
				transferFee = fromAmount.Mul(v)
			}
		}
		if param.Key == common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE {
			v, err := decimal.NewFromString(param.Value)
			if err != nil {
				logger.Warn("parse failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			switch param.ValueType {
			case common.PARAMETER_VALUE_TYPE_AMOUNT:
				exchangeFee = utils.Ptr(v)
				if fromCard.Currency != param.Currency {
					feeRates, err := utils.ListExchangeRate(ctx, fromCard.Currency, []common.Currency{param.Currency})
					if err != nil {
						logger.Warn("get exchange rate failed,", err)
						return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
					}
					exchangeFee = utils.Ptr(v.Mul(feeRates[0].Rate))
				}
			case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
				exchangeFee = utils.Ptr(fromAmount.Mul(v))
			}
		}
	}

	for _, param := range params {
		if param.Key == common.PARAMETER_KEY_TRANSFER_TRANSFER_EXCHANGE_FEE {
			if param.SpecialValue == common.SPECIAL_VALUE_NULL {
				continue
			}
			v, err := decimal.NewFromString(param.Value)
			if err != nil {
				logger.Warn("parse failed,", err)
				return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			}
			switch param.ValueType {
			case common.PARAMETER_VALUE_TYPE_AMOUNT:
				exchangeFee = utils.Ptr(v)
				if fromCard.Currency != param.Currency {
					feeRates, err := utils.ListExchangeRate(ctx, fromCard.Currency, []common.Currency{param.Currency})
					if err != nil {
						logger.Warn("get exchange rate failed,", err)
						return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
					}
					exchangeFee = utils.Ptr(v.Mul(feeRates[0].Rate))
				}
			case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
				exchangeFee = utils.Ptr(fromAmount.Mul(v))
			}
		}
	}

	if fromCard.Currency == toCard.Currency {
		exchangeFee = nil
	}

	if fromAmount.LessThanOrEqual(transferFee.Add(func() decimal.Decimal {
		if exchangeFee == nil {
			return decimal.Zero
		}
		return *exchangeFee
	}())) {
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_TRANSFER_AMOUNT_TOO_LOW)
	}

	preview := &transferDao.Preview{
		UserID:         userID,
		ToCardID:       toCard.ID,
		ToCategoryID:   toCard.CategoryID,
		ToCurrency:     toCard.Currency,
		ToUserID:       toCard.UserID,
		ToEmail:        form.ToEmail,
		ToAddress:      form.ToAddress,
		FromAmount:     fromAmount,
		FromCardID:     fromCard.ID,
		FromCategoryID: fromCard.CategoryID,
		FromCurrency:   fromCard.Currency,
		FromEmail:      fromEmail,
		FromAddress:    fromAddress,
		Mainnet:        common.Mainnet(0).FromString(form.Mainnet),
		Protocol:       common.Protocol(0).FromString(form.Protocol),
		ExchangeFee:    exchangeFee,
		Fee:            transferFee,
		Rate:           make([]*common.ExchangeRate, 0, 10),
		Channel:        channel,
		ExpiredAt:      time.Now().Add(utils.Config.TopUp.PreviewExpireSeconds * time.Second),
		Remark:         form.Remark,
	}

	for _, rate := range rates {
		previewRate := &common.ExchangeRate{}
		if err := copier.Copy(previewRate, rate); err != nil {
			logger.Warn("copy failed,", err)
			return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		preview.Rate = append(preview.Rate, previewRate)
	}

	preview.ToAmount = fromAmount.Sub(preview.Fee).Sub(func() decimal.Decimal {
		if exchangeFee == nil {
			return decimal.Zero
		}
		return *exchangeFee
	}()).Div(toBuyPrice)

	cryptos, err := ts.cryptoCurrencyDao.ListByCurrencies(ctx, []common.Currency{
		fromCard.Currency,
		toCard.Currency,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var fromPlaces, toPlaces int
	var mainnetName string
	for _, crypto := range cryptos {
		if crypto.CurrencyType == fromCard.Currency {
			fromPlaces = crypto.DisplayDecimals
		}
		if crypto.CurrencyType != toCard.Currency {
			continue
		}
		if form.Mainnet == "" {
			toPlaces = crypto.DisplayDecimals
		} else if crypto.Mainnet == common.Mainnet(0).FromString(form.Mainnet) {
			toPlaces = crypto.DisplayDecimals
			mainnetName = crypto.MainnetFullName
		}
	}

	preview.MainnetName = mainnetName
	preview.FromAmount = preview.FromAmount.Round(int32(fromPlaces))
	preview.Fee = preview.Fee.Round(int32(fromPlaces))
	inverseRate := preview.ToAmount.Round(int32(toPlaces)).Div(preview.FromAmount.Sub(preview.Fee)).Round(int32(utils.Config.System.RatePrecision))
	preview.DisplayRate = &inverseRate
	preview.ToAmount = preview.FromAmount.Sub(preview.Fee).Mul(inverseRate).Round(int32(toPlaces))
	preview.ExchangeFee = func() *decimal.Decimal {
		if exchangeFee == nil {
			return nil
		}
		f := preview.ExchangeFee.RoundFloor(int32(fromPlaces))
		return &f
	}()

	data, err := json.Marshal(preview)
	if err != nil {
		logger.Warn("marshal failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	key := utils.Md5String(string(data) + time.Now().String())

	if err := ts.previewDao.Save(ctx, key, preview, utils.Config.Transfer.PreviewExpireSeconds*time.Second); err != nil {
		logger.Warn("save failed,", err)
		return nil, "", 0, 0, decimal.Decimal{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return preview, key, fromPlaces, toPlaces, inverseRate, nil
}

func (ts *TransferService) TransferConfirm(ctx context.Context, form *entities.TransferConfirmForm, userID uint64) (*string, error) {
	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, form.TransferKey),
		utils.Config.Transfer.PreviewExpireSeconds*time.Second,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, form.TransferKey), err)
		return nil, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, form.TransferKey), err)
		}
	}()

	user, err := ts.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if utils.CheckBcryptHash(form.PinCode, user.Salt, user.PinCode) {
		return nil, utils.NewBusinessError(ctx, common.CODE_USER_PIN_CODE_CONFIRMATION_FAILED)
	}

	preview, err := ts.previewDao.Get(ctx, form.TransferKey)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	if preview == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_TRANSFER_EXPIRED)
	}

	if preview.Channel == common.TRANSFER_CHANNEL_WITHDRAW {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_CHANGE_TO_WITHDRAW)
	}

	defer func() {
		if err := ts.previewDao.Remove(ctx, form.TransferKey); err != nil {
			logger.Warn("delete failed,", err)
		}
	}()

	if fromAsset, err := ts.assetDao.GetByIDUserID(ctx, preview.FromCardID, userID); err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if fromAsset == nil {
		logger.Warn("no card asset, ID: %d", preview.FromCardID)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	} else if fromAsset.Amount.LessThan(preview.FromAmount) {
		// TODO: transfer failed
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_INSUFFICIENT_FUND)
	}

	fromCard, err := ts.cardDao.GetByID(ctx, preview.FromCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromCard == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
	}

	toCard, err := ts.cardDao.GetByID(ctx, preview.ToCardID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if toCard == nil {
		return nil, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CARD)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNO := "TRF_" + strconv.FormatUint(preview.FromCardID, 10) + "_" + strconv.FormatUint(preview.ToCardID, 10) +
		"_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

	logger.Infof("start transfering, card ID: [%d] -> [%d], order NO: [%s]", preview.FromCardID, preview.ToCardID, orderNO)

	var tErr error
	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)

		order := &orderDao.TransferOrder{
			OrderNO:        orderNO,
			UserID:         userID,
			ToAmount:       preview.ToAmount,
			ToUserID:       preview.ToUserID,
			ToCardID:       preview.ToCardID,
			ToCategoryID:   preview.ToCategoryID,
			ToCurrency:     preview.ToCurrency,
			ToEmail:        preview.ToEmail,
			ToAddress:      preview.ToAddress,
			FromAmount:     preview.FromAmount,
			FromCardID:     preview.FromCardID,
			FromCategoryID: preview.FromCategoryID,
			FromCurrency:   preview.FromCurrency,
			FromEmail:      preview.FromEmail,
			FromAddress:    preview.FromAddress,
			Mainnet:        preview.Mainnet,
			Protocol:       preview.Protocol,
			ExchangeFee:    preview.ExchangeFee,
			TransferFee:    preview.Fee,
			Channel:        preview.Channel,
			Status:         common.TRANSFER_STATUS_SUCCESS,
		}

		for _, rate := range preview.Rate {
			if rate.BaseCurrency == preview.FromCurrency &&
				rate.QuoteCurrency == preview.ToCurrency {
				order.ExchangeRate = &rate.Rate
				break
			}
		}

		_, tErr = ts.transferOrderDao.Save(ctx, order)
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		var rowsAffected int64
		rowsAffected, tErr = ts.transactionRecordDao.Saves(ctx, []*orderDao.TransactionRecord{
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_TRANSFER,
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
				FromUserID:              preview.UserID,
				FromCardID:              preview.FromCardID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				FromEmail:               preview.FromEmail,
				FromAddress:             preview.FromAddress,
				FromAlias:               fromCard.Alias,
				FromPANNumber:           fromCard.PANNumber,
				ToUserID:                preview.ToUserID,
				ToCardID:                preview.ToCardID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				ToEmail:                 preview.ToEmail,
				ToAddress:               preview.ToAddress,
				ToAlias:                 toCard.Alias,
				ToPANNumber:             toCard.PANNumber,
				Mainnet:                 preview.Mainnet,
				Protocol:                preview.Protocol,
				TransferChannel:         preview.Channel,
				DisplayRate:             preview.DisplayRate,
				ExchangeFee:             preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				TransferFee:             &preview.Fee,
				TransferFeeCurrency:     preview.FromCurrency,
				ExecutorRole:            common.ROLE_USER,
				Status:                  common.TRANSACTION_STATUS_TRANSFER_SUCCESS,
				Remark:                  preview.Remark,
			},
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_TRANSFER,
				TransactionNO:           orderNO,
				UserID:                  preview.ToUserID,
				CardID:                  preview.ToCardID,
				Income:                  decimal.NewNullDecimal(preview.ToAmount),
				IncomeCategoryID:        preview.ToCategoryID,
				IncomeCurrency:          preview.ToCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.FromAmount.Neg()),
				AgainstIncomeCategoryID: preview.FromCategoryID,
				AgainstIncomeCurrency:   preview.FromCurrency,
				Side:                    common.TRANSACTION_SIDE_TO,
				FromUserID:              preview.UserID,
				FromCardID:              preview.FromCardID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				FromEmail:               preview.FromEmail,
				FromAddress:             preview.FromAddress,
				ToUserID:                preview.ToUserID,
				ToCardID:                preview.ToCardID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				ToEmail:                 preview.ToEmail,
				ToAddress:               preview.ToAddress,
				Mainnet:                 preview.Mainnet,
				Protocol:                preview.Protocol,
				TransferChannel:         preview.Channel,
				DisplayRate:             preview.DisplayRate,
				ExchangeFee:             preview.ExchangeFee,
				ExchangeFeeCurrency:     preview.FromCurrency,
				TransferFee:             &preview.Fee,
				TransferFeeCurrency:     preview.FromCurrency,
				ExecutorRole:            common.ROLE_USER,
				Status:                  common.TRANSACTION_STATUS_TRANSFER_SUCCESS,
				Remark:                  preview.Remark,
			},
		})
		if tErr != nil {
			logger.Warn("saves failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if rowsAffected != 2 {
			logger.Warnf("duplicated save: [%+v]", preview)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		var card *cardDao.Card
		card, tErr = ts.cardDao.GetsByFreezeStatusIDForShare(ctx, common.CARD_FREEZE_STATUS_UNFROZEN, order.FromCardID)
		if tErr != nil {
			logger.Warn("get failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if card == nil {
			tErr = utils.NewBusinessError(ctx, common.CODE_TRANSFER_FROM_CARD_FROZEN)
			return tErr
		}

		card, tErr = ts.cardDao.GetsByFreezeStatusIDForShare(ctx, common.CARD_FREEZE_STATUS_UNFROZEN, order.ToCardID)
		if tErr != nil {
			logger.Warn("get failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if card == nil {
			tErr = utils.NewBusinessError(ctx, common.CODE_TRANSFER_FROM_CARD_FROZEN)
			return tErr
		}

		transactions := []*accountDao.AssetTransaction{
			// 平台帳戶加幣
			{
				UserID:   common.SYSTEM_ACCOUNT_PLATFORM,
				OrderNO:  orderNO,
				Currency: preview.FromCurrency,
				Amount: preview.FromAmount.Sub(preview.Fee).Sub(func() decimal.Decimal {
					if preview.ExchangeFee == nil {
						return decimal.Zero
					}
					return *preview.ExchangeFee
				}()),
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_ADD,
			},
			// 手續費帳戶加手續費
			{
				UserID:          common.SYSTEM_ACCOUNT_FEE,
				OrderNO:         orderNO,
				Currency:        preview.FromCurrency,
				Amount:          preview.Fee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_FEE_ADD,
			},
			// 用戶扣幣
			{
				UserID:     userID,
				CardID:     preview.FromCardID,
				OrderNO:    orderNO,
				CategoryID: preview.FromCategoryID,
				Currency:   preview.FromCurrency,
				Amount: preview.FromAmount.Sub(preview.Fee).Sub(func() decimal.Decimal {
					if preview.ExchangeFee == nil {
						return decimal.Zero
					}
					return *preview.ExchangeFee
				}()),
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_DEDUCT,
			},
			// 用戶扣手續費
			{
				UserID:          userID,
				CardID:          preview.FromCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.FromCategoryID,
				Currency:        preview.FromCurrency,
				Amount:          preview.Fee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_FEE_DEDUCT,
			},
			// 用戶加幣
			{
				UserID:          preview.ToUserID,
				CardID:          preview.ToCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.ToCategoryID,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_TRANSFER_RECEIVE_ADD,
			},
			// 平台扣幣
			{
				UserID:          common.SYSTEM_ACCOUNT_PLATFORM,
				OrderNO:         orderNO,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_TRANSFER_RECEIVE_DEDUCT,
			},
		}

		if preview.ExchangeFee != nil {
			// 手續費帳戶加 匯差費
			systemTransaction := &accountDao.AssetTransaction{
				UserID:          common.SYSTEM_ACCOUNT_FEE,
				OrderNO:         orderNO,
				Currency:        preview.FromCurrency,
				Amount:          *preview.ExchangeFee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_EXCHANGE_FEE_ADD,
			}
			transactions = append(transactions, systemTransaction)
			// 用戶扣 匯差費
			userTransaction := &accountDao.AssetTransaction{
				UserID:          userID,
				CardID:          preview.FromCardID,
				OrderNO:         orderNO,
				CategoryID:      preview.FromCategoryID,
				Currency:        preview.FromCurrency,
				Amount:          *preview.ExchangeFee,
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_EXCHANGE_FEE_DEDUCT,
			}
			transactions = append(transactions, userTransaction)
		}

		tErr = ts.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_TRANSFER, false)
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
		logger.Warn("transaction failed,", err)
		return nil, err
	}

	logger.Infof("transfered, card ID: [%d] -> [%d], order NO: [%s]", preview.FromCardID, preview.ToCardID, orderNO)

	if fromCard.Type != common.ASSET_TYPE_CARD_PRODUCT {
		return &orderNO, nil
	}

	return &orderNO, nil
}

func (ts *TransferService) createWallet(ctx context.Context, form *entities.CreateWalletForm, userID uint64) (uint64, error) {

	categoryID := uint64(0)
	currency := common.Currency(0)
	if form.CategoryID != 0 {
		currency = common.Currency(form.CategoryID)
		categoryID = uint64(form.CategoryID)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CATEGORY)
		}
	} else if form.Currency != "" {
		currency = common.Currency(0).FromString(form.Currency)
		categoryID = uint64(currency)
		if !currency.IsValid() {
			return 0, utils.NewBusinessError(ctx, common.CODE_TRANSFER_NO_SUCH_CURRENCY)
		}
	}

	// 法幣、e卡不自動開卡
	if categoryID >= 200 {
		return 0, utils.NewBusinessError(ctx, common.CODE_TRANSFER_INVALID_CATEGORY)
	}

	category, err := ts.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		logger.Warn("get failed", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if category.Usage&common.CATEGORY_USAGE_USER_APPLY == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_TRANSFER_INVALID_CATEGORY)
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
		wallet, tErr = ts.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if wallet != nil {
			id = wallet.ID
			return nil
		}

		id, tErr = ts.cardDao.Save(ctx, &cardDao.Card{
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

		if _, tErr = ts.assetDao.Save(ctx, &accountDao.Asset{
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

func (ts *TransferService) GetCategory(ctx context.Context, form *entities.TransferConfirmForm) (fromCategory *accountDao.Category, toCategory *accountDao.Category, err error) {
	preview, err := ts.previewDao.Get(ctx, form.TransferKey)
	if err != nil {
		logger.Warn("get failed,", err)
		return
	}

	if preview == nil {
		err = utils.NewBusinessError(ctx, common.CODE_CARD_APPLY_EXPIRED)
		return
	}

	fromCategory, err = ts.categoryDao.GetByID(ctx, preview.FromCategoryID)
	if err != nil {
		logger.Warn("get failed,", err)
		return
	}
	if fromCategory == nil {
		err = utils.NewBusinessError(ctx, common.CODE_CARD_NO_SUCH_ASSET_CATEGORY)
		return
	}

	toCategory, err = ts.categoryDao.GetByID(ctx, preview.ToCategoryID)
	if err != nil {
		logger.Warn("get failed,", err)
		return
	}
	if toCategory == nil {
		err = utils.NewBusinessError(ctx, common.CODE_CARD_NO_SUCH_ASSET_CATEGORY)
		return
	}

	return
}
