package services

import (
	accountDao "api-server/dao/account"
	coinsdoDao "api-server/dao/coinsdo"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/jinzhu/copier"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type AccountService struct {
	assetDao              *accountDao.AssetDao
	assetDailySnapshotDao *accountDao.AssetDailySnapshotDao
	categoryDao           *accountDao.CategoryDao
	manualOrderDao        *orderDao.ManualOrderDao
	transactionRecordDao  *orderDao.TransactionRecordDao
	assetTransactionDao   *accountDao.AssetTransactionDao
	userDao               *userDao.UserDao
	parameterDao          *systemDao.ParameterDao
	cryptoCurrencyDao     *coinsdoDao.CryptoCurrencyDao
	logger                lib.Logger
}

func NewAccountService(assetDao *accountDao.AssetDao,
	assetDailySnapshotDao *accountDao.AssetDailySnapshotDao,
	categoryDao *accountDao.CategoryDao,
	manualOrderDao *orderDao.ManualOrderDao,
	transactionRecordDao *orderDao.TransactionRecordDao,
	assetTransactionDao *accountDao.AssetTransactionDao,
	userDao *userDao.UserDao,
	parameterDao *systemDao.ParameterDao,
	cryptoCurrencyDao *coinsdoDao.CryptoCurrencyDao,
	logger lib.Logger) *AccountService {

	return &AccountService{
		assetDao:              assetDao,
		assetDailySnapshotDao: assetDailySnapshotDao,
		categoryDao:           categoryDao,
		manualOrderDao:        manualOrderDao,
		transactionRecordDao:  transactionRecordDao,
		assetTransactionDao:   assetTransactionDao,
		userDao:               userDao,
		parameterDao:          parameterDao,
		cryptoCurrencyDao:     cryptoCurrencyDao,
		logger:                logger,
	}
}

func (as *AccountService) GetAsset(ctx context.Context, id uint64) (*accountDao.Asset, error) {

	asset, err := as.assetDao.GetByID(ctx, id)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return asset, nil
}

func (as *AccountService) ListAssetsByIDInUserID(ctx context.Context, ids []uint64, userID uint64) ([]*accountDao.Asset, error) {

	assets, err := as.assetDao.Gets(ctx, &accountDao.AssetQuery{
		Asset: accountDao.Asset{
			UserID: userID,
		},
		IDIn: ids,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) GetAssetByCategoryID(ctx context.Context, categoryID uint64, userID uint64) (*accountDao.Asset, error) {

	asset, err := as.assetDao.GetByUserIDCategoryID(ctx, userID, categoryID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return asset, nil
}

func (as *AccountService) ListAssets(ctx context.Context, form *entities.ListAssetsForm, userID uint64) ([]*accountDao.Asset, error) {

	assets, err := as.assetDao.Gets(ctx, &accountDao.AssetQuery{
		Asset: accountDao.Asset{
			UserID:     userID,
			CategoryID: form.CategoryID,
		},
		IDIn:         form.IDIn,
		CategoryIDIn: form.CategoryIDIn,
		TypeIn:       form.TypeIn,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) ListAssetsByUserIDIn(ctx context.Context, form *entities.ListAssetsForm, userIDs []uint64) ([]*accountDao.Asset, error) {

	assets, err := as.assetDao.Gets(ctx, &accountDao.AssetQuery{
		Asset: accountDao.Asset{
			CategoryID: form.CategoryID,
		},
		IDIn:         form.IDIn,
		CategoryIDIn: form.CategoryIDIn,
		TypeIn:       form.TypeIn,
		UserIDIn:     userIDs,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) ListAssetsByType(ctx context.Context, form *entities.ListAssetsByTypeForm, userID uint64) ([]*accountDao.Asset, error) {

	var err error
	var assets []*accountDao.Asset
	if form.Type == 0 {
		assets, err = as.assetDao.ListByUserID(ctx, userID)
	} else {
		assets, err = as.assetDao.ListByUserIDType(ctx, userID, form.Type)
	}

	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) GetCategory(ctx context.Context, form *entities.GetCategoryForm) (*accountDao.Category, error) {

	category, err := as.categoryDao.GetByID(ctx, form.ID)

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return category, nil
}

func (as *AccountService) ListCategories(ctx context.Context, form *entities.ListCategoriesForm) ([]*accountDao.Category, error) {

	categories, err := as.categoryDao.ListByIDs(ctx, form.IDIn)

	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Category{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return categories, nil
}

func (as *AccountService) AddCategory(ctx context.Context, form *entities.AddCategoryForm) (uint64, error) {

	c := &accountDao.Category{}
	if err := copier.Copy(c, form); err != nil {
		return 0, err
	}
	id, err := as.categoryDao.Save(ctx, c)
	if err != nil {
		logger.Warn("save failed,", err)
		return 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return id, nil
}

func (as *AccountService) GetByCurrency(ctx context.Context, currency common.Currency) ([]*accountDao.Asset, error) {

	assets, err := as.assetDao.GetByCurrency(ctx, currency)

	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Asset{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assets, nil
}

func (as *AccountService) GetTotalAmountByCurrency(ctx context.Context, currency common.Currency) (decimal.Decimal, error) {

	totalAmount, err := as.assetDao.GetTotalAmountByCurrency(ctx, currency)

	if err != nil {
		logger.Warn("get failed,", err)
		return decimal.Zero, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return totalAmount, nil
}

func (as *AccountService) ManualTransfer(ctx context.Context, form *entities.ManualTransferForm) (string, error) {

	j, err := json.Marshal(form)
	if err != nil {
		logger.Warn("marshal failed", err)
		return "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	key := utils.Md5String(string(j))

	locker := utils.NewLocker()
	if err := locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_MANUAL, key),
		utils.Config.Manual.LockSeconds*time.Second,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_MANUAL, key), err)
		return "", err
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_MANUAL, key), err)
		}
	}()

	fromAsset, err := as.assetDao.GetByID(ctx, form.AssetID)
	if err != nil {
		logger.Warn("get failed,", err)
		return "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if fromAsset == nil {
		return "", utils.NewBusinessError(ctx, common.CODE_ACCOUNT_NO_SUCH_ASSET)
	}

	toAsset := fromAsset
	if form.AgainstAssetID != 0 {
		toAsset, err = as.assetDao.GetByID(ctx, form.AgainstAssetID)
		if err != nil {
			logger.Warn("get failed,", err)
			return "", utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if toAsset == nil {
			return "", utils.NewBusinessError(ctx, common.CODE_ACCOUNT_NO_SUCH_ASSET)
		}
		if fromAsset.Currency != toAsset.Currency {
			return "", utils.NewBusinessError(ctx, common.CODE_ACCOUNT_INVALID_TARGET)
		}
	} else {
		switch form.Type {
		case common.TRANSACTION_TYPE_ASSET_ADD,
			common.TRANSACTION_TYPE_ASSET_DEDUCT,
			common.TRANSACTION_TYPE_FROZEN_ASSET_ADD,
			common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT:
			return "", utils.NewBusinessError(ctx, common.CODE_ACCOUNT_INVALID_TRANSACTION_TYPE)
		}
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNO := "MUA_" + strconv.FormatUint(fromAsset.ID, 10) + "_" + strconv.FormatUint(toAsset.ID, 10) +
		"_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

	logger.Infof("start manual transfering, card ID: [%d] -> [%d], order NO: [%s]", fromAsset.ID, toAsset.ID, orderNO)

	var tErr error
	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)

		order := &orderDao.ManualOrder{
			BusinessNO:      form.BusinessNO,
			OrderNO:         orderNO,
			UserID:          fromAsset.UserID,
			AssetID:         fromAsset.ID,
			Currency:        fromAsset.Currency,
			Amount:          form.Amount,
			AgainstUserID:   toAsset.UserID,
			AgainstAssetID:  toAsset.ID,
			TransactionType: form.Type,
			CreatedBy:       form.By,
			Memo:            form.Memo,
			Status:          common.MANUAL_STATUS_SUCCESS,
		}

		_, tErr = as.manualOrderDao.Save(ctx, order)
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		income := form.Amount

		switch form.Type {
		case common.TRANSACTION_TYPE_ASSET_ADD, common.TRANSACTION_TYPE_FROZEN_ASSET_ADD:
		//no-op
		case common.TRANSACTION_TYPE_ASSET_FREEZE, common.TRANSACTION_TYPE_ASSET_UNFREEZE:
			income = decimal.Zero
		case common.TRANSACTION_TYPE_ASSET_DEDUCT, common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT:
			income = income.Neg()
		}

		var rowsAffected int64
		rowsAffected, tErr = as.transactionRecordDao.Saves(ctx, []*orderDao.TransactionRecord{
			{
				Type:             common.TRANSACTION_RECORD_TYPE_MANUAL,
				TransactionNO:    orderNO,
				UserID:           fromAsset.UserID,
				CardID:           fromAsset.ID,
				Income:           decimal.NewNullDecimal(income),
				IncomeCategoryID: fromAsset.CategoryID,
				IncomeCurrency:   fromAsset.Currency,
				Side:             common.TRANSACTION_SIDE_FROM,
				FromCardID:       fromAsset.ID,
				FromCategoryID:   fromAsset.CategoryID,
				FromCurrency:     fromAsset.Currency,
				FromAmount:       decimal.NewNullDecimal(form.Amount),
				ToCardID:         toAsset.ID,
				ToCategoryID:     toAsset.CategoryID,
				ToCurrency:       toAsset.Currency,
				ToAmount:         decimal.NewNullDecimal(form.Amount),
				Status:           common.TRANSACTION_STATUS_MANUAL_SUCCESS,
			},
			{
				Type:             common.TRANSACTION_RECORD_TYPE_MANUAL,
				TransactionNO:    orderNO,
				UserID:           toAsset.UserID,
				CardID:           toAsset.ID,
				Income:           decimal.NewNullDecimal(income.Neg()),
				IncomeCategoryID: toAsset.CategoryID,
				IncomeCurrency:   toAsset.Currency,
				Side:             common.TRANSACTION_SIDE_TO,
				FromCardID:       fromAsset.ID,
				FromCategoryID:   fromAsset.CategoryID,
				FromCurrency:     fromAsset.Currency,
				FromAmount:       decimal.NewNullDecimal(form.Amount),
				ToCardID:         toAsset.ID,
				ToCategoryID:     toAsset.CategoryID,
				ToCurrency:       toAsset.Currency,
				ToAmount:         decimal.NewNullDecimal(form.Amount),
				Status:           common.TRANSACTION_STATUS_MANUAL_SUCCESS,
			},
		})
		if err != nil {
			logger.Warn("saves failed,", err)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if rowsAffected != 2 {
			logger.Warnf("save failed: [%#v]", order)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		transactions := make([]*accountDao.AssetTransaction, 0, 2)

		switch form.Type {
		case common.TRANSACTION_TYPE_ASSET_ADD:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶加錢
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
					BillType:        common.BILL_TYPE_MANUAL_ADD_ADD,
				},
				&accountDao.AssetTransaction{ // 對方扣錢
					UserID:          toAsset.UserID,
					CardID:          toAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      toAsset.CategoryID,
					Currency:        toAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
					BillType:        common.BILL_TYPE_MANUAL_ADD_DEDUCT,
				},
			)
		case common.TRANSACTION_TYPE_ASSET_DEDUCT:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶扣錢
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
					BillType:        common.BILL_TYPE_MANUAL_DEDUCT_DEDUCT,
				},
				&accountDao.AssetTransaction{ // 對方加錢
					UserID:          toAsset.UserID,
					CardID:          toAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      toAsset.CategoryID,
					Currency:        toAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
					BillType:        common.BILL_TYPE_MANUAL_DEDUCT_ADD,
				},
			)
		case common.TRANSACTION_TYPE_ASSET_FREEZE:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶凍結
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_FREEZE,
					BillType:        common.BILL_TYPE_MANUAL_FREEZE_FREEZE,
				},
			)
		case common.TRANSACTION_TYPE_ASSET_UNFREEZE:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶解凍
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_UNFREEZE,
					BillType:        common.BILL_TYPE_MANUAL_UNFREEZE_UNFREEZE,
				},
			)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_ADD:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶加凍結
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_FROZEN_ASSET_ADD,
					BillType:        common.BILL_TYPE_MANUAL_ADD_FREEZE_ADD_FREEZE,
				},
				&accountDao.AssetTransaction{ // 對方扣錢
					UserID:          toAsset.UserID,
					CardID:          toAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      toAsset.CategoryID,
					Currency:        toAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
					BillType:        common.BILL_TYPE_MANUAL_ADD_FREEZE_DEDUCT,
				},
			)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT:
			transactions = append(transactions,
				&accountDao.AssetTransaction{ // 用戶扣凍結
					UserID:          fromAsset.UserID,
					CardID:          fromAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      fromAsset.CategoryID,
					Currency:        fromAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT,
					BillType:        common.BILL_TYPE_MANUAL_DEDUCT_FREEZE_DEDUCT_FREEZE,
				},
				&accountDao.AssetTransaction{ // 對方加錢
					UserID:          toAsset.UserID,
					CardID:          toAsset.ID,
					OrderNO:         orderNO,
					CategoryID:      toAsset.CategoryID,
					Currency:        toAsset.Currency,
					Amount:          form.Amount,
					TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
					BillType:        common.BILL_TYPE_MANUAL_DEDUCT_FREEZE_ADD,
				},
			)
		}

		tErr = as.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_MANUAL, true)
		if tErr != nil {
			logger.Warn("transaction failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
			return tErr
		}

		return nil
	})

	if tErr != nil {
		logger.Warn("transaction failed,", tErr)
		return "", tErr
	}

	if err != nil {
		logger.Warn("transaction failed,", tErr)
		return "", err
	}

	logger.Infof("manual transfered, card ID: [%d] -> [%d], order NO: [%s]", fromAsset.ID, toAsset.ID, orderNO)

	return orderNO, nil
}

func (as *AccountService) GetAssetCategories(ctx context.Context) ([]*accountDao.AssetCategory, error) {

	assetCatogries, err := as.assetCategoryDao.Gets(ctx, &accountDao.AssetCategoryQuery{
		AssetCategory: accountDao.AssetCategory{},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return assetCatogries, nil
}

func (as *AccountService) ListCategoryByUsage(ctx context.Context, form *entities.ListCategoryByUsageForm) ([]*accountDao.Category, error) {

	categories, err := as.categoryDao.ListByTypeUsage(ctx, 0, form.Usages)
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Category{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return categories, nil
}

func (as *AccountService) GetEquivalentValue(ctx context.Context, form *entities.GetEquivalentValueForm, userID uint64) (total decimal.Decimal, totalFrozen decimal.Decimal, fee map[common.Currency]decimal.Decimal, rates []*utils.ExchangeRate, err error) {

	currency := common.Currency(0).FromString(form.Currency)
	if currency == 0 {
		logger.Warn("invalid currency")
		return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	assets, err := as.ListAssetsByType(ctx, &entities.ListAssetsByTypeForm{
		Type: form.Type,
	}, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	quote := make([]common.Currency, 0, len(assets))
	for _, asset := range assets {
		quote = append(quote, asset.Currency)
	}
	rs, err := utils.ListExchangeRate(ctx, currency, quote)
	if err != nil {
		logger.Warn("get exchange rate failed,", err)
		return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	feeRate := decimal.Zero
	fee = make(map[common.Currency]decimal.Decimal)
	rateMap, oriRateMap := make(map[common.Currency]decimal.Decimal), make(map[common.Currency]decimal.Decimal)
	switch form.Purpose {
	case common.RATE_PURPOSE_TOP_UP:
		keys := make([]common.ParameterKey, 0, 1)
		keys = append(keys, common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE)

		params, err := as.parameterDao.ListByKeys(ctx, keys)
		if err != nil {
			logger.Warn("get failed,", err)
			return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if len(params) < 1 {
			logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE)
			return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}

		for _, param := range params {
			if param.Key == common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE {
				v, err := decimal.NewFromString(param.Value)
				if err != nil {
					logger.Warn("parse failed,", err)
					return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
				}
				switch param.ValueType {
				case common.PARAMETER_VALUE_TYPE_AMOUNT:
					// no-op
				case common.PARAMETER_VALUE_TYPE_PERCENTAGE:
					feeRate = v
				}
			}
		}

		for _, r := range rs {
			temp := decimal.NewFromInt(1)
			if !r.Rate.Equal(decimal.Zero) {
				temp = r.Rate
			}
			oriRateMap[r.QuoteCurrency] = temp.Copy()
			temp = temp.Mul(decimal.NewFromInt(1).Sub(feeRate))
			rateMap[r.QuoteCurrency] = temp
		}
	default:
		for _, r := range rs {
			temp := decimal.NewFromInt(1)
			if !r.Rate.Equal(decimal.Zero) {
				temp = r.Rate
			}
			oriRateMap[r.QuoteCurrency] = temp.Copy()
			rateMap[r.QuoteCurrency] = temp
		}

	}

	crypto, err := as.cryptoCurrencyDao.GetCryptoCurrencyByCurrencyType(ctx, currency)
	if err != nil {
		logger.Warn("get failed,", err)
		return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if crypto == nil {
		logger.Warnf("no crypto currency config: [%s]", currency.String())
		return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	total, totalFrozen = decimal.Zero, decimal.Zero
	for _, asset := range assets {
		var r decimal.Decimal
		var ok bool
		if r, ok = rateMap[asset.Currency]; !ok {
			logger.Errorf("missing rate [%s]", asset.Currency.String())
			return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		logger.Debugf("add: currency [%s][%s]", asset.Currency.String(), asset.Amount.Mul(r).RoundFloor(int32(crypto.DisplayDecimals)).String())
		total = total.Add(asset.Amount.Mul(r).RoundFloor(int32(crypto.DisplayDecimals)))
		fee[asset.Currency] = asset.Amount.Mul(feeRate)

		if r, ok = oriRateMap[asset.Currency]; !ok {
			logger.Errorf("missing rate [%s]", asset.Currency.String())
			return decimal.Decimal{}, decimal.Decimal{}, nil, nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		totalFrozen = totalFrozen.Add(asset.FreezedAmount.Mul(r).RoundFloor(int32(crypto.DisplayDecimals)))
	}

	rates = make([]*utils.ExchangeRate, 0, len(rateMap))
	for k, v := range rateMap {
		rates = append(rates, &utils.ExchangeRate{
			BaseCurrency:  currency,
			QuoteCurrency: k,
			Rate:          v,
			ActualRate:    oriRateMap[k],
		})
	}

	return
}

func (as *AccountService) DailyAssetSnapshot(ctx context.Context) error {

	date := time.Now().Format("2006-01-02")
	current, pagesize := 1, 3000
	lastTotal, totalAsset := 0, 0

	assetMap := make(map[uint64]bool)

	cctx := context.WithValue(ctx, "db", utils.DB.Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)}))

	for {
		assets, _, _, total, err := as.assetDao.Page(ctx, current, pagesize)
		if err != nil {
			return err
		}
		if len(assets) == 0 {
			break
		}
		current++
		lastTotal = total
		assetDailySnapshots := make([]*accountDao.AssetDailySnapshot, 0, len(assets))
		for _, asset := range assets {
			if _, ok := assetMap[asset.ID]; ok {
				continue
			}
			assetMap[asset.ID] = true
			assetDailySnapshots = append(assetDailySnapshots, &accountDao.AssetDailySnapshot{
				SnapshotDate:    date,
				ID:              asset.ID,
				CategoryID:      asset.CategoryID,
				Type:            asset.Type,
				UserID:          asset.UserID,
				Currency:        asset.Currency,
				CurrencyType:    asset.CurrencyType,
				Amount:          asset.Amount,
				FreezedAmount:   asset.FreezedAmount,
				Signature:       asset.Signature,
				DeletedAt:       asset.DeletedAt,
				SourceCreatedAt: asset.CreatedAt,
				SourceUpdatedAt: asset.UpdatedAt,
			})
		}
		if totalAsset == 0 {
			totalAsset = total
		}

		if len(assetDailySnapshots) > 0 {
			rowsAffected, err := as.assetDailySnapshotDao.Saves(cctx, assetDailySnapshots)
			if err != nil {
				ids := make([]uint64, 0, len(assetDailySnapshots))
				for _, assetDailySnapshot := range assetDailySnapshots {
					ids = append(ids, assetDailySnapshot.ID)
				}
				logger.Warn("snapshot save failed [%v], ids: [%v]", err, ids)
				continue
			}

			if rowsAffected != int64(len(assetDailySnapshots)) {
				ids := make([]uint64, 0, len(assetDailySnapshots))
				for _, assetDailySnapshot := range assetDailySnapshots {
					ids = append(ids, assetDailySnapshot.ID)
				}
				logger.Warnf("duplicated save: [%v]", ids)
				continue
			}
		}
	}

	if lastTotal != totalAsset {
		logger.Warnf("the amount of assets changed")
	}

	return nil

}
