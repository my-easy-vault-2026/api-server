package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/my-easy-vault-2026/shared-modules/common"
	"github.com/my-easy-vault-2026/shared-modules/utils"

	accountDao "github.com/my-easy-vault-2026/api-server/dao/account"
	cardDao "github.com/my-easy-vault-2026/api-server/dao/card"
	coinsdoDao "github.com/my-easy-vault-2026/api-server/dao/coinsdo"
	orderDao "github.com/my-easy-vault-2026/api-server/dao/order"
	systemDao "github.com/my-easy-vault-2026/api-server/dao/system"
	transferDao "github.com/my-easy-vault-2026/api-server/dao/transfer"
	userDao "github.com/my-easy-vault-2026/api-server/dao/user"
	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TransferService struct {
	transferOrderDao     *orderDao.TransferOrderDao
	cardDao              *cardDao.CardDao
	assetDao             *accountDao.AssetDao
	parameterDao         *systemDao.ParameterDao
	currencyConfigDao    *coinsdoDao.CurrencyConfigDao
	previewDao           *transferDao.PreviewDao
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
	mq                   *infra.MQ
}

func NewTransferService(transferOrderDao *orderDao.TransferOrderDao,
	cardDao *cardDao.CardDao,
	assetDao *accountDao.AssetDao,
	parameterDao *systemDao.ParameterDao,
	currencyConfigDao *coinsdoDao.CurrencyConfigDao,
	previewDao *transferDao.PreviewDao,
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
	mq *infra.MQ) *TransferService {
	return &TransferService{
		transferOrderDao:     transferOrderDao,
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
		mq:                   mq,
	}
}

func (ts *TransferService) TransferPreview(ctx context.Context,
	fromAmount decimal.Decimal,
	fromWalletID uint64,
	toEmail string,
	toUserID uint64,
	userID uint64) (preview *transferDao.Preview, key string, places int, err error) {
	fromUser, err := ts.userDao.GetByUserID(ctx, userID)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if fromUser == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	fromWallet, err := ts.cardDao.GetByIDUserID(ctx, fromWalletID, userID)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if fromWallet == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	var toUser *userDao.User
	if toUserID != 0 {
		toUser, err = ts.userDao.GetByUserID(ctx, toUserID)
	} else if toEmail != "" {
		toUser, err = ts.userDao.GetByEmailRole(ctx, toEmail, common.ROLE_USER)
	} else {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_MISSING_PARAMETER)
	}
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if toUser == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_USER)
		return
	}

	toWallet, err := ts.cardDao.GetByUserIDCurrencyType(ctx, toUser.ID, fromWallet.Currency, common.ASSET_TYPE_FIAT)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if toWallet == nil {
		var walletID uint64
		walletID, err = ts.createWallet(ctx, fromWallet.Currency, toUser.ID)
		if err != nil {
			ts.logger.Warn("create failed,", err)
			err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}
		toWallet, err = ts.cardDao.GetByID(ctx, walletID)
		if err != nil {
			ts.logger.Warn("get failed,", err)
			err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}
	}

	if fromWallet.UserID == toWallet.UserID {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SELF_TRANSFER)
		return
	}

	asset, err := ts.assetDao.GetByIDUserID(ctx, fromWallet.ID, userID)

	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if asset == nil || asset.Amount.LessThan(fromAmount) {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_INSUFFICIENT_FUNDS)
		return
	}

	params, err := ts.parameterDao.ListByKeys(ctx, []common.ParameterKey{
		common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE,
	})
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if len(params) < 1 {
		ts.logger.Warnf("no parameter: [%s][%s]", common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	transferFee := decimal.Zero
	for _, param := range params {
		if param.Key == common.PARAMETER_KEY_TRANSFER_TRANSFER_FEE {
			v, err2 := decimal.NewFromString(param.Value)
			if err2 != nil {
				ts.logger.Warn("parse failed,", err2)
				err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
				return
			}
			switch param.ValueType {
			case common.PARAMETER_VALUE_TYPE_AMOUNT:
				transferFee = v
			case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
				transferFee = fromAmount.Mul(v)
			}
		}
	}

	preview = &transferDao.Preview{
		UserID:         userID,
		ToWalletID:     toWallet.ID,
		ToCategoryID:   toWallet.CategoryID,
		ToCurrency:     toWallet.Currency,
		ToUserID:       toWallet.UserID,
		FromAmount:     fromAmount,
		FromWalletID:   fromWallet.ID,
		FromCategoryID: fromWallet.CategoryID,
		FromCurrency:   fromWallet.Currency,
		Fee:            transferFee,
		ExpiredAt:      time.Now().Add(ts.env.PreviewExpiryTime),
	}

	preview.ToAmount = fromAmount.Sub(preview.Fee)

	configs, err := ts.currencyConfigDao.ListByCurrencies(ctx, []common.Currency{fromWallet.Currency})
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if len(configs) == 0 {
		ts.logger.Warnf("no currency config: [%s]", fromWallet.Currency)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	places = configs[0].Decimals

	preview.FromAmount = preview.FromAmount.Round(int32(places))
	preview.Fee = preview.Fee.Round(int32(places))
	preview.ToAmount = preview.FromAmount.Sub(preview.Fee).Round(int32(places))

	data, err := json.Marshal(preview)
	if err != nil {
		ts.logger.Warn("marshal failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	key = utils.Md5String(string(data) + time.Now().String())

	if err = ts.previewDao.Save(ctx, key, preview, ts.env.PreviewExpiryTime); err != nil {
		ts.logger.Warn("save failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	return
}

func (ts *TransferService) TransferConfirm(ctx context.Context, key string, pinCode string, userID uint64) (orderNO string, err error) {
	locker := ts.lockers.NewLocker()
	if err = locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, key),
		ts.env.PreviewExpiryTime,
		ts.env.LockWaitDuration,
	); err != nil {
		ts.logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, key), err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			ts.logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_TRANSFER_CONFIRM, key), err)
		}
	}()

	user, err := ts.userDao.GetByUserID(ctx, userID)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	saltedPassword := pinCode + user.Salt
	err = bcrypt.CompareHashAndPassword([]byte(user.PinCode), []byte(saltedPassword))
	if err != nil {
		ts.logger.Infof("invalid pin code for user: %d", userID)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_INVALED_PIN_CODE)
		return
	}

	preview, err := ts.previewDao.Get(ctx, key)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if preview == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_PREVIEW_EXPIRED)
		return
	}

	defer func() {
		if err := ts.previewDao.Remove(ctx, key); err != nil {
			ts.logger.Warn("delete failed,", err)
		}
	}()

	if fromAsset, err2 := ts.assetDao.GetByIDUserID(ctx, preview.FromWalletID, userID); err2 != nil {
		ts.logger.Warn("get failed,", err2)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	} else if fromAsset == nil {
		ts.logger.Warn("no card asset, ID: %d", preview.FromWalletID)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	} else if fromAsset.Amount.LessThan(preview.FromAmount) {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_INSUFFICIENT_FUNDS)
		return
	}

	fromWallet, err := ts.cardDao.GetByID(ctx, preview.FromWalletID)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if fromWallet == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	toWallet, err := ts.cardDao.GetByID(ctx, preview.ToWalletID)
	if err != nil {
		ts.logger.Warn("get failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if toWallet == nil {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
		return
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNO = "TRF_" + strconv.FormatUint(preview.FromWalletID, 10) + "_" + strconv.FormatUint(preview.ToWalletID, 10) +
		"_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

	ts.logger.Infof("start transfering, wallet ID: [%d] -> [%d], order NO: [%s]", preview.FromWalletID, preview.ToWalletID, orderNO)

	var tErr error
	err = utils.WithTX(ts.db.DB, func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)
		transferOrderDaoTX := ts.transferOrderDao.WithTx(ts.db.DB)
		transactionRecordDaoTX := ts.transactionRecordDao.WithTx(ts.db.DB)
		assetTransactionDaoTX := ts.assetTransactionDao.WithTx(ts.db.DB)

		order := &orderDao.TransferOrder{
			OrderNO:        orderNO,
			UserID:         userID,
			ToAmount:       preview.ToAmount,
			ToUserID:       preview.ToUserID,
			ToWalletID:     preview.ToWalletID,
			ToCategoryID:   preview.ToCategoryID,
			ToCurrency:     preview.ToCurrency,
			FromAmount:     preview.FromAmount,
			FromWalletID:   preview.FromWalletID,
			FromCategoryID: preview.FromCategoryID,
			FromCurrency:   preview.FromCurrency,
			Fee:            preview.Fee,
			FeeCurrency:    preview.FeeCurrency,
			Status:         common.TRANSFER_STATUS_SUCCESS,
		}

		var id uint64
		id, tErr = transferOrderDaoTX.Save(ctx, order)
		if tErr != nil {
			ts.logger.Warn("save failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if id == 0 {
			ts.logger.Warn("save failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_DB_INSERT_FAILED)
			return tErr
		}

		var rowsAffected int64
		rowsAffected, tErr = transactionRecordDaoTX.Saves(ctx, []*orderDao.TransactionRecord{
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_TRANSFER,
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
				FromUserID:              preview.UserID,
				FromWalletID:            preview.FromWalletID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				ToUserID:                preview.ToUserID,
				ToWalletID:              preview.ToWalletID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				TransferFee:             &preview.Fee,
				TransferFeeCurrency:     preview.FromCurrency,
				Status:                  common.TRANSACTION_STATUS_TRANSFER_SUCCESS,
			},
			{
				Type:                    common.TRANSACTION_RECORD_TYPE_TRANSFER,
				TransactionNO:           orderNO,
				UserID:                  preview.ToUserID,
				WalletID:                preview.ToWalletID,
				Income:                  decimal.NewNullDecimal(preview.ToAmount),
				IncomeCategoryID:        preview.ToCategoryID,
				IncomeCurrency:          preview.ToCurrency,
				AgainstIncome:           decimal.NewNullDecimal(preview.FromAmount.Neg()),
				AgainstIncomeCategoryID: preview.FromCategoryID,
				AgainstIncomeCurrency:   preview.FromCurrency,
				Side:                    common.TRANSACTION_SIDE_FROM,
				FromUserID:              preview.UserID,
				FromWalletID:            preview.FromWalletID,
				FromCategoryID:          preview.FromCategoryID,
				FromCurrency:            preview.FromCurrency,
				FromAmount:              decimal.NewNullDecimal(preview.FromAmount),
				ToUserID:                preview.ToUserID,
				ToWalletID:              preview.ToWalletID,
				ToCategoryID:            preview.ToCategoryID,
				ToCurrency:              preview.ToCurrency,
				ToAmount:                decimal.NewNullDecimal(preview.ToAmount),
				TransferFee:             &preview.Fee,
				TransferFeeCurrency:     preview.FromCurrency,
				Status:                  common.TRANSACTION_STATUS_TRANSFER_SUCCESS,
			},
		})
		if tErr != nil {
			ts.logger.Warn("saves failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if rowsAffected != 2 {
			ts.logger.Warnf("duplicated save: [%#v]", preview)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
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
				Amount:          preview.FromAmount.Sub(preview.Fee),
				TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
				BillType:        common.BILL_TYPE_TRANSFER_SEND_DEDUCT,
			},
			// 用戶扣手續費
			{
				UserID:          userID,
				WalletID:        preview.FromWalletID,
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
				WalletID:        preview.ToWalletID,
				OrderNO:         orderNO,
				CategoryID:      preview.ToCategoryID,
				Currency:        preview.ToCurrency,
				Amount:          preview.ToAmount,
				TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
				BillType:        common.BILL_TYPE_TRANSFER_RECEIVE_ADD,
			},
		}

		tErr = assetTransactionDaoTX.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_TRANSFER, false)
		if tErr != nil {
			ts.logger.Warn("transaction failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_ASSET_TRANSACTION_FAILED)
			return tErr
		}
		return nil
	})

	if tErr != nil {
		ts.logger.Warn("transaction failed,", tErr)
		err = tErr
		return
	}

	if err != nil {
		ts.logger.Warn("transaction failed,", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
		return
	}

	traceID := ctx.Value(common.CTX_KEY_TRACE_ID)
	msg := &common.Msg{
		OP:       common.MSG_OPCODE_INFUND,
		MsgID:    traceID.(string),
		RecordID: orderNO,
		Subject:  "Get Infund",
		Msg:      fmt.Sprintf("Received %s %s from %s. order no: %s", preview.ToAmount, preview.ToCurrency, user.Email, orderNO),
		UserID:   preview.ToUserID,
	}
	err = ts.mq.Pub(ctx, utils.GetPubsubKey("websocket"), msg)
	if err != nil {
		ts.logger.Warn("push failed,", err)
	}

	ts.logger.Infof("transfered, wallet ID: [%d] -> [%d], order NO: [%s]", preview.FromWalletID, preview.ToWalletID, orderNO)

	return
}

func (ts *TransferService) createWallet(ctx context.Context, currency common.Currency, userID uint64) (walletID uint64, err error) {

	categoryID := uint64(currency)
	if !currency.IsValid() {
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_CURRENCY)
		return
	}

	category, err := ts.categoryDao.GetByID(ctx, categoryID)
	if err != nil {
		ts.logger.Warn("get failed", err)
		err = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	locker := ts.lockers.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)),
		ts.env.LockDuration,
		ts.env.LockWaitDuration,
	); err != nil {
		ts.logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		return 0, err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			ts.logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_WALLET, strconv.FormatUint(userID, 10), strconv.FormatUint(categoryID, 10)), err)
		}
	}()

	var id uint64

	var tErr error
	err = utils.WithTX(ts.db.DB, func(tx *gorm.DB) error {
		cardDaoTX := ts.cardDao.WithTx(ts.db.DB)
		assetDaoTX := ts.assetDao.WithTx(ts.db.DB)
		var wallet *cardDao.Card
		wallet, tErr = cardDaoTX.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(currency))
		if tErr != nil {
			ts.logger.Warn("get failed", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
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
			ts.logger.Warn("save failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
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
			ts.logger.Warn("save failed,", tErr)
			tErr = ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		return 0, tErr
	}

	if err != nil {
		ts.logger.Warn("transaction failed,", err)
		return 0, ts.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return id, nil
}
