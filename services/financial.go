package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	financialDao "api-server/dao/financial"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"math/rand"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type FinancialService struct {
	assetDao                  *accountDao.AssetDao
	categoryDao               *accountDao.CategoryDao
	userDao                   *userDao.UserDao
	snapshotBalanceDao        *financialDao.SnapshotBalanceDao
	parameterDao              *systemDao.ParameterDao
	financialProductDao       *financialDao.FinancialProductDao
	interestRateDao           *financialDao.InterestRateDao
	cardDao                   *cardDao.CardDao
	interestOrderDao          *orderDao.InterestOrderDao
	transactionRecordDao      *orderDao.TransactionRecordDao
	cryptoCurrencyDao         *coinsdoDao.CryptoCurrencyDao
	assetTransactionDao       *accountDao.AssetTransactionDao
	financialTransferOrderDao *orderDao.FinancialTransferOrderDao
	assetSnapshotDao          *financialDao.AssetSnapshotDao
}

func NewFinancialService() *FinancialService {
	return &FinancialService{
		assetDao:                  accountDao.NewAssetDao(),
		categoryDao:               accountDao.NewCategoryDao(),
		userDao:                   userDao.NewUserDao(),
		snapshotBalanceDao:        financialDao.NewSnapshotBalanceDao(),
		parameterDao:              systemDao.NewParameterDao(),
		financialProductDao:       financialDao.NewFinancialProductDao(),
		interestRateDao:           financialDao.NewInterestRateDao(),
		cardDao:                   cardDao.NewCardDao(),
		interestOrderDao:          orderDao.NewInterestOrderDao(),
		transactionRecordDao:      orderDao.NewTransactionRecordDao(),
		cryptoCurrencyDao:         coinsdoDao.NewCryptoCurrencyDao(),
		assetTransactionDao:       accountDao.NewAssetTransactionDao(),
		financialTransferOrderDao: orderDao.NewFinancialTransferOrderDao(),
		assetSnapshotDao:          financialDao.NewAssetSnapshotDao(),
	}
}

func (fs *FinancialService) CheckBalanceSnapshot(ctx context.Context, code common.FinancialCode, aType common.AssetType, currencies []common.Currency, now time.Time) error {

	lossTimeMap := make(map[int64]time.Time)

	current := 1
	for {
		// 取得所有用��的 asset 資料
		assets, _, _, _, err := fs.assetDao.PageByTypeCurrencyInOrderByID(ctx, aType, currencies, current, utils.Config.Financial.SnapshotBatchSize)
		if err != nil {
			logger.Warn("get all assets failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if len(assets) == 0 {
			break
		}

		userIDs := make([]uint64, 0, len(assets))
		for _, a := range assets {
			userIDs = append(userIDs, a.UserID)
		}

		cctx := context.WithValue(ctx, "db", utils.DB.Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)}))

		users, err := fs.userDao.ListByUserIDIn(cctx, userIDs)
		if err != nil {
			logger.Warn("get users failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		userMap := make(map[uint64]*userDao.User)
		for _, user := range users {
			userMap[user.ID] = user
		}

		for _, a := range assets {
			_, ok := userMap[a.UserID]
			if !ok {
				logger.Warnf("user not found, user id: %d", a.UserID)
				continue
			}

			for i := 0; i < 24; i++ {
				if _, ok := lossTimeMap[now.Add(time.Hour*time.Duration(-i-1)).UnixMilli()]; ok {
					continue
				}

				if oldSnapshot, err := fs.snapshotBalanceDao.Get(ctx, &financialDao.SnapshotBalance{
					CardID:        a.ID,
					FinancialCode: common.FINANCIAL_CODE_AUTO_YIELD,
					TakenAt:       now.Add(time.Hour * time.Duration(-i-1)),
				}); err != nil {
					logger.Warn("get snapshot balance failed,", err)
				} else if oldSnapshot == nil {
					logger.Warnf("snapshot lost: user: [%d], card: [%d], time: [%s]", a.UserID, a.ID, now.Add(time.Hour*time.Duration(-i-1)))
					lossTimeMap[now.Add(time.Hour*time.Duration(-i-1)).UnixMilli()] = now.Add(time.Hour * time.Duration(-i-1))
				}
			}
		}
		current++
	}

	if len(lossTimeMap) == 0 {
		return nil
	}

	logger.Infof("snapshot lost: time: [%#v]", lossTimeMap)
	for _, t := range lossTimeMap {
		err := fs.BalanceSnapshot(ctx, common.FINANCIAL_CODE_AUTO_YIELD, t, aType, currencies)
		if err != nil {
			logger.Warn("redo snapshot failed,", err)
		}
	}

	return nil
}

func (fs *FinancialService) BalanceSnapshot(ctx context.Context, code common.FinancialCode, now time.Time, aType common.AssetType, currencies []common.Currency) error {

	product, err := fs.financialProductDao.GetByCode(ctx, code)
	if err != nil {
		logger.Warn("get financial product failed,", err)
		product = &financialDao.FinancialProduct{
			SnapshotExpiracy: 1560,
		}
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 0
	current := 1
	for {
		// 取得所有用��的 asset 資料
		assets, _, _, _, err := fs.assetDao.PageByTypeCurrencyInOrderByID(ctx, aType, currencies, current, utils.Config.Financial.SnapshotBatchSize)
		if err != nil {
			logger.Warn("get all assets failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if len(assets) == 0 {
			break
		}

		userIDs := make([]uint64, 0, len(assets))
		for _, a := range assets {
			userIDs = append(userIDs, a.UserID)
		}

		cctx := context.WithValue(ctx, "db", utils.DB.Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)}))

		users, err := fs.userDao.ListByUserIDIn(cctx, userIDs)
		if err != nil {
			logger.Warn("get users failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		userMap := make(map[uint64]*userDao.User)
		for _, user := range users {
			userMap[user.ID] = user
		}

		autoYieldCards, err := fs.cardDao.ListByTypeUserIDIn(cctx, common.ASSET_TYPE_AUTO_YIELD, userIDs)
		if err != nil {
			logger.Warn("get auto yield cards failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		ayCardMap := make(map[uint64]map[common.Currency]*cardDao.Card)
		for _, c := range autoYieldCards {
			cardMap, ok := ayCardMap[c.UserID]
			if !ok {
				cardMap = make(map[common.Currency]*cardDao.Card)
				ayCardMap[c.UserID] = cardMap
			}
			cardMap[c.Currency] = c
		}

		for _, a := range assets {
			user, ok := userMap[a.UserID]
			if !ok {
				logger.Warnf("user not found, user id: %d", a.UserID)
				continue
			}

			earningStatus := common.CARD_EARNING_STATUS_DISABLED
			if cardMap, ok := ayCardMap[a.UserID]; ok {
				if c, ok := cardMap[a.Currency]; ok {
					earningStatus = c.EarningStatus
				}
			}

			sb := &financialDao.SnapshotBalance{
				UserID:         a.UserID,
				UserRole:       user.Role,
				UserKYCLevel:   string(user.KycLevel),
				CardType:       a.Type,
				CardID:         a.ID,
				CardCategoryID: a.CategoryID,
				CardCurrency:   a.Currency,
				Balance:        a.Amount,
				FinancialCode:  code,
				EarningStatus:  earningStatus.String(),
				TakenAt:        now,
			}

			if oldSnapshot, err := fs.snapshotBalanceDao.Get(ctx, &financialDao.SnapshotBalance{
				CardID:        a.ID,
				FinancialCode: common.FINANCIAL_CODE_AUTO_YIELD,
				TakenAt:       now,
			}); err != nil {
				logger.Warn("get snapshot balance failed,", err)
			} else if oldSnapshot != nil {
				continue
			}

			err = fs.snapshotBalanceDao.Save(ctx, sb, (product.SnapshotExpiracy+time.Duration(rnd.Int31n(100)))*time.Minute)
			if err != nil {
				logger.Warn(ctx, "save snapshot balance failed,", err)
				continue
			}

			count++
		}

		current++
	}

	logger.Infof("balance snapshot [%d]", count)
	return nil
}

func (fs *FinancialService) AutoYield(ctx context.Context, aType common.AssetType, currencies []common.Currency, now time.Time) error {

	product, err := fs.financialProductDao.GetByCode(ctx, common.FINANCIAL_CODE_AUTO_YIELD)
	if err != nil {
		logger.Warn(ctx, "get financial product failed,", err)
		logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if product == nil || product.Status != common.FINANCIAL_PRODUCT_STATUS_ACTIVE {
		return nil
	}

	inputCurrencies := currencies

	switch product.SupportType {
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_ALL:
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_CURRENCY:
		sups := strings.Split(product.SupportedCurrencies, ",")
		supSet := make(map[common.Currency]bool)
		for _, s := range sups {
			supSet[common.Currency(0).FromString(s)] = true
		}
		newCurrencies := make([]common.Currency, 0, len(currencies))
		for _, c := range currencies {
			if supSet[c] {
				newCurrencies = append(newCurrencies, c)
			}
		}
		currencies = newCurrencies
	}

	if len(currencies) == 0 {
		return nil
	}

	currencyCount := make(map[common.Currency]int)
	for _, c := range currencies {
		currencyCount[c] = 0
	}

	isLeapYear := (time.Now().Year()%4 == 0 && time.Now().Year()%100 != 0) || time.Now().Year()%400 == 0
	hoursOfYear := decimal.NewFromInt(365 * 24)
	if isLeapYear {
		hoursOfYear = decimal.NewFromInt(366 * 24)
	}

	interestRates, err := fs.interestRateDao.List(ctx)
	if err != nil {
		logger.Warn(ctx, "get interest rates failed,", err)
		logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	interestRateMap := make(map[common.Currency]*financialDao.InterestRate)
	for _, ir := range interestRates {
		interestRateMap[ir.Currency] = ir
	}

	// 限定一定要有設置各幣種的門檻
	// param, err := fs.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_AUTO_YIELD_THRESHOLD)
	// if err != nil {
	// 	logger.Warn("get threshold failed,", err)
	// 	logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
	// 	return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	// }
	// if param == nil {
	// 	logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_AUTO_YIELD_THRESHOLD)
	// 	param = &systemDao.Parameter{
	// 		Value: "100",
	// 	}
	// }
	//
	// threshold, err := decimal.NewFromString(param.Value)
	// thresholdCurrency := common.CURRENCY_USD
	// if err != nil {
	// 	logger.Warn("parse threshold failed,", err)
	// 	threshold = decimal.NewFromInt(100)
	// }

	categories, err := fs.categoryDao.List(ctx)
	if err != nil {
		logger.Warn("get failed,", err)
		logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	categoriesMap := make(map[uint64]*accountDao.Category)
	for _, c := range categories {
		categoriesMap[c.ID] = c
	}

	count := 0
	userCurrent := 1
	for {
		users, _, _, _, err := fs.userDao.PageByRole(ctx, common.ROLE_USER, userCurrent, utils.Config.Financial.UserBatchSize)
		if err != nil {
			logger.Warn(ctx, "get users failed,", err)
			logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}

		if len(users) == 0 {
			break
		}

		userIDs := make([]uint64, 0, len(users))
		for _, user := range users {
			// if user.KycLevel != common.KYC_LEVEL_2 && user.KycLevel != common.KYC_LEVEL_3 {
			// 	continue
			// }
			userIDs = append(userIDs, user.ID)
		}

		cctx := context.WithValue(ctx, "db", utils.DB.Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)}))

		assets, err := fs.assetDao.ListByTypeCurrencyInUserIDIn(cctx, aType, currencies, userIDs)
		if err != nil {
			logger.Warn(ctx, "get assets failed,", err)
			logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}

		userCurrencyAssetMap := make(map[uint64]map[common.Currency]map[uint64]*accountDao.Asset) // map[user id]map[currency]map[asset id]asset
		for _, asset := range assets {

			if asset.Type != common.ASSET_TYPE_CRYPTO && asset.Type != common.ASSET_TYPE_FIAT && asset.Type != common.ASSET_TYPE_CARD_PRODUCT {
				continue
			}

			category, ok := categoriesMap[asset.CategoryID]
			if !ok {
				logger.Warnf("category not found, category id: [%d]", asset.CategoryID)
				continue
			}

			if category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD == 0 {
				continue
			}

			if _, ok := userCurrencyAssetMap[asset.UserID]; !ok {
				userCurrencyAssetMap[asset.UserID] = make(map[common.Currency]map[uint64]*accountDao.Asset)
			}
			if _, ok := userCurrencyAssetMap[asset.UserID][asset.Currency]; !ok {
				userCurrencyAssetMap[asset.UserID][asset.Currency] = make(map[uint64]*accountDao.Asset)
			}
			userCurrencyAssetMap[asset.UserID][asset.Currency][asset.ID] = asset
		}

		// 限定一定要有設置各幣種的門檻
		// thresholdRateMap := make(map[common.Currency]map[common.Currency]*utils.ExchangeRate)
		// for _, currencyAssetMap := range userCurrencyAssetMap {
		// 	for currency := range currencyAssetMap {
		// 		if _, ok := thresholdRateMap[currency]; !ok {
		// 			thresholdRateMap[currency] = make(map[common.Currency]*utils.ExchangeRate)
		// 		}
		// 		assetThresholdCurrency := thresholdCurrency
		// 		if r, ok := interestRateMap[currency]; ok && r.Threshold != nil {
		// 			assetThresholdCurrency = currency
		// 		}
		// 		if _, ok := thresholdRateMap[currency][assetThresholdCurrency]; !ok {
		// 			thresholdRate, err := utils.GetExchangeRate(ctx, currency, assetThresholdCurrency)
		// 			if err != nil {
		// 				logger.Warn("get exchange rate failed,", err)
		// 				logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
		// 				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		// 			}
		//
		// 			if thresholdRate == nil {
		// 				logger.Warnf("get exchange rate failed, %s, %s", currency, assetThresholdCurrency)
		// 				logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%v", time.Now().Format(time.RFC3339), currencies)
		// 				return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		// 			}
		//
		// 			thresholdRateMap[currency][assetThresholdCurrency] = thresholdRate
		// 		}
		// 	}
		// }

		for userID, currencyAssetMap := range userCurrencyAssetMap {
			for currency, assetMap := range currencyAssetMap {
				totalAmount := decimal.Zero
				for _, a := range assetMap {
					totalAmount = totalAmount.Add(a.Amount)
				}
				// 限定一定要有設置各幣種的門檻
				// assetThreshold := threshold.Copy()
				// assetThresholdCurrency := thresholdCurrency
				// if r, ok := interestRateMap[currency]; ok && r.Threshold != nil {
				// 	assetThreshold = *r.Threshold
				// 	assetThresholdCurrency = currency
				// }
				//
				// if tr, ok := thresholdRateMap[currency]; !ok || tr == nil {
				// 	logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
				// 	continue
				// }
				//
				// if tr, ok := thresholdRateMap[currency][assetThresholdCurrency]; !ok || tr == nil {
				// 	logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
				// 	continue
				// }
				//
				// thresholdRate := thresholdRateMap[currency][assetThresholdCurrency]
				var assetThreshold decimal.Decimal
				if r, ok := interestRateMap[currency]; ok && r.Threshold != nil {
					assetThreshold = *r.Threshold
				} else {
					logger.Warnf("threshold not set: [%d]", currency)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}
				thresholdRate := &utils.ExchangeRate{
					Rate: decimal.NewFromInt(1),
				}

				allSnapshotInsufficient := true

				interestRate, ok := interestRateMap[currency]
				if !ok {
					if product.DefaultYieldRate == nil {
						logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
						continue
					}
					interestRate = &financialDao.InterestRate{
						Rate: *product.DefaultYieldRate,
					}
				}

				crypto, err := fs.cryptoCurrencyDao.GetCryptoCurrencyByCurrencyType(ctx, currency)
				if err != nil {
					logger.Warn("get crypto currency failed,", err)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}
				if crypto == nil {
					logger.Warnf("no crypto currency: [%s]", currency)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				toCard, err := fs.cardDao.GetByUserIDCurrencyType(ctx, userID, currency, common.ASSET_TYPE_AUTO_YIELD)
				if err != nil {
					logger.Warn("get to card failed,", err)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				if toCard == nil {
					toCard, err = fs.CreateAutoYieldCard(ctx, userID, currency)
					if err != nil {
						logger.Warn("create card failed,", err)
						logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
						continue
					}
				}

				parentOrderNO := "IRS_" + strconv.FormatUint(toCard.ID, 10) + "_" + string(common.FINANCIAL_CODE_AUTO_YIELD) + "_" + now.Format("20060102")

				currencyIntererst := decimal.Zero
				interestOrderMap := make(map[uint64]*orderDao.InterestOrder)
				interestOrders := make([]*orderDao.InterestOrder, 0)
				assetSnapshots := make([]*financialDao.AssetSnapshot, 0)
				assetSnapshotMap := make(map[uint64]map[int64]*financialDao.SnapshotBalance) // 卡ID 時間
				for _, asset := range assetMap {

					interest := decimal.Zero

					snapshots := make([]*financialDao.SnapshotBalance, 24)
					snapshots[0] = &financialDao.SnapshotBalance{
						UserID:         userID,
						UserRole:       common.ROLE_USER,
						UserKYCLevel:   string(common.KYC_LEVEL_2),
						CardType:       asset.Type,
						CardID:         asset.ID,
						CardCategoryID: asset.CategoryID,
						CardCurrency:   asset.Currency,
						Balance:        decimal.Zero,
						FinancialCode:  common.FINANCIAL_CODE_AUTO_YIELD,
						EarningStatus:  common.CARD_EARNING_STATUS_ENABLED.String(),
						TakenAt:        now.Add(time.Hour * time.Duration(-1)),
					}

					orderSnapshots := make([]*financialDao.AssetSnapshot, 0)
					for i := 0; i < 24; i++ {
						snapshot, err := fs.snapshotBalanceDao.Get(ctx, &financialDao.SnapshotBalance{
							CardID:        asset.ID,
							FinancialCode: common.FINANCIAL_CODE_AUTO_YIELD,
							TakenAt:       now.Add(time.Hour * time.Duration(-i-1)),
						})
						if err != nil {
							logger.Warn("get snapshot balance failed,", err)
							logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s", time.Now().Format(time.RFC3339), currency.String())
							for _, c := range inputCurrencies {
								if c == currency {
									continue
								}
								if count, ok := currencyCount[c]; !ok || count == 0 {
									logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s", time.Now().Format(time.RFC3339), c.String())
									continue
								} else {
									logger.Infof("%s INTEREST_DISTRIBUTE:currency=%s,count=%d", time.Now().Format(time.RFC3339), c.String(), count)
								}
							}
							return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
						}

						as := &financialDao.AssetSnapshot{
							UserID:         userID,
							UserRole:       common.ROLE_USER,
							FinancialCode:  common.FINANCIAL_CODE_AUTO_YIELD,
							ParentOrderNO:  parentOrderNO,
							CardType:       asset.Type,
							CardID:         asset.ID,
							CardCategoryID: asset.CategoryID,
							CardCurrency:   asset.Currency,
							Balance:        utils.Ptr(decimal.Zero),
							Interest:       utils.Ptr(decimal.Zero),
							Missing:        common.SNAPSHOT_MISSING_STATUS_MISSING,
							EarningStatus:  common.CARD_EARNING_STATUS_ENABLED,
							TakenAt:        now.Add(time.Hour * time.Duration(-i-1)),
						}

						if snapshot == nil && i == 0 {
							orderSnapshots = append(orderSnapshots, as)
							continue
						}

						if snapshot == nil {
							snapshot = &financialDao.SnapshotBalance{}
							err := copier.Copy(snapshot, snapshots[i-1])
							if err != nil {
								logger.Warnf("copy snapshot balance [%d][%s] failed,", asset.ID, now.Add(time.Hour*time.Duration(-i-1)), err)
								logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
								return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
							}
							snapshot.TakenAt = now.Add(time.Hour * time.Duration(-i-1))
							as.Balance = &snapshot.Balance
							as.Missing = common.SNAPSHOT_MISSING_STATUS_MISSING
							as.EarningStatus = common.CardEarningStatus(0).FromString(snapshot.EarningStatus)
						} else {
							as.Balance = &snapshot.Balance
							as.Missing = common.SNAPSHOT_MISSING_STATUS_NOT_MISSING
							as.EarningStatus = common.CardEarningStatus(0).FromString(snapshot.EarningStatus)
						}
						snapshots[i] = snapshot
						orderSnapshots = append(orderSnapshots, as)

						if common.CardEarningStatus(0).FromString(snapshot.EarningStatus) != common.CARD_EARNING_STATUS_ENABLED {
							continue
						}

						if snapshots[i].Balance.LessThanOrEqual(decimal.Zero) {
							continue
						}

						as.Interest = utils.Ptr(snapshots[i].Balance.Mul(interestRate.Rate.Div(hoursOfYear)))
					}

					fromCard, err := fs.cardDao.GetByID(ctx, asset.ID)
					if err != nil {
						logger.Warn("get from card failed,", err)
						continue
					}

					if fromCard == nil {
						logger.Warnf("no from card: [%d]", asset.ID)
						continue
					}

					orderNO := "IRS_" + strconv.FormatUint(asset.ID, 10) + "_" + strconv.FormatUint(toCard.ID, 10) + "_" + string(common.FINANCIAL_CODE_AUTO_YIELD) + "_" + now.Format("20060102")
					order := &orderDao.InterestOrder{
						Code:            common.FINANCIAL_CODE_AUTO_YIELD,
						ParentOrderNO:   parentOrderNO,
						OrderNO:         orderNO,
						UserID:          asset.UserID,
						FromCardType:    asset.Type,
						FromCardID:      asset.ID,
						FromCategoryID:  asset.CategoryID,
						FromCurrency:    asset.Currency,
						ToCardType:      toCard.Type,
						ToCardID:        toCard.ID,
						ToCategoryID:    toCard.CategoryID,
						ToCurrency:      toCard.Currency,
						PrincipalAmount: &asset.Amount,
						InterestAmount:  interest,
						InterestRate:    interestRate.Rate,
						CalculateCount:  0,
						Status:          common.INTEREST_ORDER_STATUS_FAILED,
						CalculatedAt:    now,
					}
					interestOrderMap[asset.ID] = order
					currencyIntererst = currencyIntererst.Add(interest)

					for _, s := range orderSnapshots {
						s.OrderNO = orderNO
						assetSnapshots = append(assetSnapshots, s)
					}

					assetSnapshotMap[asset.ID] = make(map[int64]*financialDao.SnapshotBalance)
					for _, s := range snapshots {
						assetSnapshotMap[asset.ID][s.TakenAt.UnixMilli()] = s
					}

				}

				for i := 0; i < 24; i++ {
					totalAmount := decimal.Zero
					for _, a := range assetMap {
						as, ok := assetSnapshotMap[a.ID][now.Add(time.Hour*time.Duration(-i-1)).UnixMilli()]
						if !ok {
							continue
						}

						if common.CardEarningStatus(0).FromString(as.EarningStatus) != common.CARD_EARNING_STATUS_ENABLED {
							continue
						}

						if as.Balance.LessThanOrEqual(decimal.Zero) {
							continue
						}

						totalAmount = totalAmount.Add(as.Balance)
					}

					if totalAmount.Div(thresholdRate.Rate).LessThan(assetThreshold) {
						continue
					}

					for _, a := range assetMap {
						as, ok := assetSnapshotMap[a.ID][now.Add(time.Hour*time.Duration(-i-1)).UnixMilli()]
						if !ok {
							continue
						}

						if common.CardEarningStatus(0).FromString(as.EarningStatus) != common.CARD_EARNING_STATUS_ENABLED {
							continue
						}

						if as.Balance.LessThanOrEqual(decimal.Zero) {
							continue
						}
						currencyIntererst = currencyIntererst.Add(as.Balance.Mul(interestRate.Rate.Div(hoursOfYear)))
						interestOrderMap[a.ID].CalculateCount++
						interestOrderMap[a.ID].InterestAmount = interestOrderMap[a.ID].InterestAmount.Add(as.Balance.Mul(interestRate.Rate.Div(hoursOfYear)))
						allSnapshotInsufficient = false
					}

				}

				var truncateAmount *decimal.Decimal
				if ta := currencyIntererst.Sub(currencyIntererst.RoundFloor(int32(crypto.InterestDecimals))); !ta.IsZero() {
					truncateAmount = &ta
				}
				if allSnapshotInsufficient {
					truncateAmount = utils.Ptr(currencyIntererst.Copy())
					currencyIntererst = decimal.Zero
					cardIDs := make([]string, 0, len(assetMap))
					for _, a := range assetMap {
						cardIDs = append(cardIDs, strconv.FormatUint(a.ID, 10))
					}
					logger.Infof("asset insufficient [%s]", strings.Join(cardIDs, ","))
				}
				currencyIntererst = currencyIntererst.RoundFloor(int32(crypto.InterestDecimals))

				if currencyIntererst.IsZero() {

					for _, o := range interestOrderMap {
						logger.Infof("skipped [%s][%d]", o.OrderNO, o.FromCardID)
					}

					func() {
						locker := utils.NewLocker()
						if err = locker.WaitLock(
							ctx,
							utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)),
							utils.Config.System.LockMicroseconds*time.Microsecond,
							utils.Config.System.LockWaitMicroseconds*time.Microsecond,
						); err != nil {
							logger.Warnf("lock failed: [%s], %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)), err)
							return
						}
						defer func() {
							if err := locker.UnLock(ctx); err != nil {
								logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)), err)
							}
						}()

						parentOrder, err := fs.interestOrderDao.GetByToCardIDCodeCalculatedAt(ctx, toCard.ID, common.FINANCIAL_CODE_AUTO_YIELD, now)
						if err != nil {
							return
						}
						if parentOrder != nil {
							cardIDs := make([]string, len(interestOrderMap))
							for _, o := range interestOrderMap {
								cardIDs = append(cardIDs, strconv.FormatUint(o.FromCardID, 10))
							}
							logger.Infof("interest order already exists, card ID: [%s] -> [%d], order NO: [%s]", strings.Join(cardIDs, ","), toCard.ID, parentOrderNO)
							return
						}

						parentOrder = &orderDao.InterestOrder{
							Code:            common.FINANCIAL_CODE_AUTO_YIELD,
							OrderNO:         parentOrderNO,
							UserID:          userID,
							FromCurrency:    currency,
							ToCardType:      toCard.Type,
							ToCardID:        toCard.ID,
							ToCategoryID:    toCard.CategoryID,
							ToCurrency:      toCard.Currency,
							PrincipalAmount: &totalAmount,
							InterestAmount:  currencyIntererst,
							TruncateAmount:  truncateAmount,
							InterestRate:    interestRate.Rate,
							Status:          common.INTEREST_ORDER_STATUS_FAILED,
							CalculatedAt:    now,
						}
						interestOrders = make([]*orderDao.InterestOrder, 0, len(interestOrderMap)+1)
						for _, o := range interestOrderMap {
							interestOrders = append(interestOrders, o)
						}
						interestOrders = append(interestOrders, parentOrder)

						var rowsAffected int64
						rowsAffected, err = fs.interestOrderDao.Saves(ctx, interestOrders)
						if err != nil {
							logger.Warn("save failed, ", err)
							return
						}
						if len(interestOrders) != int(rowsAffected) {
							logger.Warnf("save interest orders not complete, parent order: [%s], count:[%d/%d]", parentOrderNO, rowsAffected, len(interestOrders))
						}

						rowsAffected, err = fs.assetSnapshotDao.Saves(ctx, assetSnapshots)
						if err != nil {
							logger.Warn("save failed, ", err)
							return
						}
						if len(assetSnapshots) != int(rowsAffected) {
							logger.Warnf("save snapshots not complete, parent order: [%s], count:[%d/%d]", parentOrderNO, rowsAffected, len(assetSnapshots))
						}
					}()

					continue
				}

				for _, o := range interestOrderMap {
					o.Status = common.INTEREST_ORDER_STATUS_SUCCESS
					logger.Infof("start interest, card ID: [%d] -> [%d], order NO: [%s]", o.FromCardID, toCard.ID, o.OrderNO)
				}

				financialEnableCategoryIDs := make([]uint64, 0, len(categories))
				for _, c := range categories {
					if c.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD > 0 {
						financialEnableCategoryIDs = append(financialEnableCategoryIDs, c.ID)
					}
				}

				transferToAsset, err := fs.assetDao.GetByUserIDTypeInCurrencyCategoryIDInOrderByAmount(ctx, userID, []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_CARD_PRODUCT}, currency, financialEnableCategoryIDs)
				if err != nil {
					logger.Warn("get failed,", err)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}
				if transferToAsset == nil {
					logger.Errorf("no available transfer asset, card ID: [%d]", toCard.ID)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				transferToCard, err := fs.cardDao.GetByID(ctx, transferToAsset.ID)
				if err != nil {
					logger.Warn("get failed,", err)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}
				if transferToCard == nil {
					logger.Errorf("card asset mismatch: [%d]", transferToAsset.ID)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				lastTransfer, err := fs.financialTransferOrderDao.GetByFromCardIDOrderByCreatedAtDESC(ctx, toCard.ID)
				if err != nil {
					logger.Warn("get failed,", err)
				} else if lastTransfer == nil {
					logger.Warn("last transfer not found: %d", toCard.ID)
				} else {
					lastTransferToCard, err := fs.cardDao.GetByID(ctx, lastTransfer.ToCardID)
					if err != nil {
						logger.Warn("get failed,", err)
					} else if lastTransferToCard == nil {
						logger.Warn("last transfer to card not found: %d", lastTransfer.ToCardID)
					} else {
						lastTransferToAsset, err := fs.assetDao.GetByID(ctx, lastTransferToCard.ID)
						if err != nil {
							logger.Warn("get failed,", err)
						} else if lastTransferToAsset == nil {
							logger.Warn("last transfer to asset not found: %d", lastTransferToCard.ID)
						} else {
							transferToAsset = lastTransferToAsset
							transferToCard = lastTransferToCard
						}
					}
				}

				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				transferOrderNO := "FTS_" + strconv.FormatUint(toCard.ID, 10) + "_" + strconv.FormatUint(transferToAsset.ID, 10) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.Itoa(int(rnd.Int31n(10000)))

				var tErr error
				func() {
					locker := utils.NewLocker()
					if err = locker.WaitLock(
						ctx,
						utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)),
						utils.Config.System.LockMicroseconds*time.Microsecond,
						utils.Config.System.LockWaitMicroseconds*time.Microsecond,
					); err != nil {
						logger.Warnf("lock failed: [%s], %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)), err)
						return
					}
					defer func() {
						if err := locker.UnLock(ctx); err != nil {
							logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_DISTRIBUTE_AUTO_YIELD, strconv.FormatUint(toCard.ID, 10), now.Format("20060102"), string(common.FINANCIAL_CODE_AUTO_YIELD)), err)
						}
					}()

					parentOrder, err := fs.interestOrderDao.GetByToCardIDCodeCalculatedAt(ctx, toCard.ID, common.FINANCIAL_CODE_AUTO_YIELD, now)
					if err != nil {
						return
					}
					if parentOrder != nil {
						cardIDs := make([]string, len(interestOrders))
						for i, o := range interestOrders {
							cardIDs[i] = strconv.FormatUint(o.FromCardID, 10)
						}
						logger.Infof("interest order already exists, card ID: [%s] -> [%d], order NO: [%s]", strings.Join(cardIDs, ","), toCard.ID, parentOrderNO)
						return
					}

					parentOrder = &orderDao.InterestOrder{
						Code:            common.FINANCIAL_CODE_AUTO_YIELD,
						OrderNO:         parentOrderNO,
						UserID:          userID,
						FromCurrency:    currency,
						ToCardType:      toCard.Type,
						ToCardID:        toCard.ID,
						ToCategoryID:    toCard.CategoryID,
						ToCurrency:      toCard.Currency,
						PrincipalAmount: &totalAmount,
						InterestAmount:  currencyIntererst,
						TruncateAmount:  truncateAmount,
						InterestRate:    interestRate.Rate,
						Status:          common.INTEREST_ORDER_STATUS_SUCCESS,
						CalculatedAt:    now,
					}

					interestOrders = make([]*orderDao.InterestOrder, 0, len(interestOrderMap)+1)
					for _, o := range interestOrderMap {
						interestOrders = append(interestOrders, o)
					}
					interestOrders = append(interestOrders, parentOrder)

					err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
						var ctx = context.WithValue(ctx, "db", tx)

						var rowsAffected int64
						rowsAffected, tErr = fs.interestOrderDao.Saves(ctx, interestOrders)
						if tErr != nil {
							logger.Warn("save failed", tErr)
							return tErr
						}
						if len(interestOrders) != int(rowsAffected) {
							logger.Warnf("save interest orders not complete, parent order: [%s], count:[%d/%d]", parentOrderNO, rowsAffected, len(interestOrders))
						}

						rowsAffected, tErr = fs.assetSnapshotDao.Saves(ctx, assetSnapshots)
						if tErr != nil {
							logger.Warn("save failed", tErr)
							return tErr
						}
						if len(assetSnapshots) != int(rowsAffected) {
							logger.Warnf("save snapshots not complete, parent order: [%s], count:[%d/%d]", parentOrderNO, rowsAffected, len(assetSnapshots))
						}

						rowsAffected, err := fs.transactionRecordDao.Saves(ctx, []*orderDao.TransactionRecord{
							{
								Type:                  common.TRANSACTION_RECORD_TYPE_INTEREST,
								TransactionNO:         parentOrderNO,
								TransferTransactionNO: transferOrderNO,
								UserID:                userID,
								CardID:                toCard.ID,
								Income:                decimal.NewNullDecimal(currencyIntererst),
								IncomeCategoryID:      toCard.CategoryID,
								IncomeCurrency:        toCard.Currency,
								Side:                  common.TRANSACTION_SIDE_TO,
								FromCurrency:          currency,
								ToCardID:              toCard.ID,
								ToCategoryID:          toCard.CategoryID,
								ToCurrency:            toCard.Currency,
								ToAmount:              decimal.NewNullDecimal(currencyIntererst),
								TransferToType:        transferToCard.Type,
								TransferToCardID:      transferToCard.ID,
								TransferToCategoryID:  transferToCard.CategoryID,
								TransferToCurrency:    transferToCard.Currency,
								TransferToUserID:      transferToCard.UserID,
								TransferToAlias:       transferToCard.Alias,
								TransferToPANNumber:   transferToCard.PANNumber,
								FinancialCode:         product.Code,
								YieldRate:             &interestRate.Rate,
								ExecutorRole:          common.ROLE_SYSTEM,
								Status:                common.TRANSACTION_STATUS_INTEREST_SUCCESS,
							},
						})
						if err != nil {
							logger.Warn("saves failed,", err)
							return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
						}
						if rowsAffected != 1 {
							logger.Warnf("duplicated save: [%+v]", parentOrder)
							return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
						}

						transactions := []*accountDao.AssetTransaction{
							// 利息帳戶扣利息
							{
								UserID:          common.SYSTEM_ACCOUNT_PLATFORM,
								OrderNO:         parentOrderNO,
								Currency:        currency,
								Amount:          currencyIntererst,
								TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
								BillType:        common.BILL_TYPE_INTEREST_DEDUCT,
							},
							// 用戶加利息
							{
								UserID:          userID,
								CardID:          toCard.ID,
								OrderNO:         parentOrderNO,
								CategoryID:      toCard.CategoryID,
								Currency:        toCard.Currency,
								Amount:          currencyIntererst,
								TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
								BillType:        common.BILL_TYPE_INTEREST_ADD,
							},
						}

						tErr = fs.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_INTEREST, false)
						if tErr != nil {
							logger.Warn("transaction failed,", tErr)
							tErr = utils.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
							return tErr
						}

						return nil
					})
				}()

				if tErr != nil {
					logger.Errorf("auto yield transaction failed, card: [%d], interest: [%s], %v", toCard.ID, currencyIntererst, tErr)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				if err != nil {
					logger.Errorf("auto yield transaction failed, card: [%d], interest: [%s], %v", toCard.ID, currencyIntererst, err)
					logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s,user_id=%d", time.Now().Format(time.RFC3339), currency, userID)
					continue
				}

				count++
				currencyCount[currency] = currencyCount[currency] + 1

				fromCardIDs := make([]string, len(interestOrders))
				for i, o := range interestOrders {
					if o.ParentOrderNO == "" {
						continue
					}
					fromCardIDs[i] = strconv.FormatUint(o.FromCardID, 10)
				}
				logger.Infof("[AUTO YIELD] #%d user: [%d], card: [%s], auto yield card: [%d], intererst: [%s]", count, userID, strings.Join(fromCardIDs, ","), toCard.ID, currencyIntererst)

				toAsset, err := fs.assetDao.GetByID(ctx, toCard.ID)
				if err != nil {
					logger.Warn("get failed,", err)
					continue
				}

				transferAmount := toAsset.Amount.RoundFloor(int32(crypto.DisplayDecimals))
				if transferAmount.IsZero() {
					logger.Infof("no need to transfer, card ID: [%d]", toCard.ID)
					continue
				}

				logger.Infof("start financial transfer, card ID: [%d] -> [%d], order NO: [%s]", toCard.ID, transferToAsset.ID, transferOrderNO)

				err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
					var ctx = context.WithValue(ctx, "db", tx)

					var toAsset *accountDao.Asset
					toAsset, tErr = fs.assetDao.GetByIDForUpdate(ctx, toCard.ID)
					if tErr != nil {
						logger.Warn("get to asset failed,", tErr)
						return tErr
					}

					transferAmount = toAsset.Amount.RoundFloor(int32(crypto.DisplayDecimals))
					if transferAmount.IsZero() {
						logger.Infof("no need to transfer, card ID: [%d]", toCard.ID)
						return nil
					}

					order := &orderDao.FinancialTransferOrder{
						OrderNO:        transferOrderNO,
						FromType:       toCard.Type,
						FromAmount:     transferAmount,
						FromUserID:     toCard.UserID,
						FromCardID:     toCard.ID,
						FromCategoryID: toAsset.CategoryID,
						FromCurrency:   toCard.Currency,
						ToType:         transferToAsset.Type,
						ToAmount:       transferAmount,
						ToUserID:       transferToAsset.UserID,
						ToCardID:       transferToAsset.ID,
						ToCategoryID:   transferToAsset.CategoryID,
						ToCurrency:     transferToAsset.Currency,
						Direction:      common.FINANCIAL_TRANSFER_DIRECTION_OUT,
						TriggerMethod:  common.FINANCIAL_TRANSFER_TRIGGER_METHOD_INSTANT,
						Status:         common.FINANCIAL_TRANSFER_STATUS_SUCCESS,
					}

					_, tErr = fs.financialTransferOrderDao.Save(ctx, order)
					if tErr != nil {
						logger.Warn("save failed", tErr)
						return tErr
					}

					rowsAffected, err := fs.transactionRecordDao.Saves(ctx, []*orderDao.TransactionRecord{
						{
							Type:             common.TRANSACTION_RECORD_TYPE_FINANCIAL_TRANSFER,
							TransactionNO:    transferOrderNO,
							UserID:           transferToAsset.UserID,
							CardID:           toCard.ID,
							Income:           decimal.NewNullDecimal(transferAmount.Neg()),
							IncomeCategoryID: toCard.CategoryID,
							IncomeCurrency:   toCard.Currency,
							Side:             common.TRANSACTION_SIDE_FROM,
							FromCardID:       toCard.ID,
							FromCategoryID:   toCard.CategoryID,
							FromCurrency:     toCard.Currency,
							FromAmount:       decimal.NewNullDecimal(transferAmount),
							FromAlias:        toCard.Alias,
							ToCardID:         transferToAsset.ID,
							ToCategoryID:     transferToAsset.CategoryID,
							ToCurrency:       transferToAsset.Currency,
							ToAlias:          transferToCard.Alias,
							ToAmount:         decimal.NewNullDecimal(transferAmount),
							ExecutorRole:     common.ROLE_SYSTEM,
							Status:           common.TRANSACTION_STATUS_FINANCIAL_TRANSFER_SUCCESS,
						},
						{
							Type:             common.TRANSACTION_RECORD_TYPE_FINANCIAL_TRANSFER,
							TransactionNO:    transferOrderNO,
							UserID:           transferToAsset.UserID,
							CardID:           transferToAsset.ID,
							Income:           decimal.NewNullDecimal(transferAmount),
							IncomeCategoryID: transferToAsset.CategoryID,
							IncomeCurrency:   transferToAsset.Currency,
							Side:             common.TRANSACTION_SIDE_TO,
							FromCardID:       toCard.ID,
							FromCategoryID:   toCard.CategoryID,
							FromCurrency:     toCard.Currency,
							FromAmount:       decimal.NewNullDecimal(transferAmount),
							FromAlias:        toCard.Alias,
							ToCardID:         transferToAsset.ID,
							ToCategoryID:     transferToAsset.CategoryID,
							ToCurrency:       transferToAsset.Currency,
							ToAlias:          transferToCard.Alias,
							ToAmount:         decimal.NewNullDecimal(transferAmount),
							ExecutorRole:     common.ROLE_SYSTEM,
							Status:           common.TRANSACTION_STATUS_FINANCIAL_TRANSFER_SUCCESS,
						},
					})
					if err != nil {
						logger.Warn("saves failed,", err)
						return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
					}
					if rowsAffected != 2 {
						logger.Warnf("duplicated save: [%+v]", order)
						return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
					}

					transactions := []*accountDao.AssetTransaction{
						// 發方卡片扣款
						{
							UserID:          toCard.UserID,
							CardID:          toCard.ID,
							OrderNO:         transferOrderNO,
							CategoryID:      toCard.CategoryID,
							Currency:        toCard.Currency,
							Amount:          transferAmount,
							TransactionType: common.TRANSACTION_TYPE_ASSET_DEDUCT,
							BillType:        common.BILL_TYPE_FINANCIAL_TRANSFER_SEND_DEDUCT,
						},
						// 收方卡片入款
						{
							UserID:          transferToAsset.UserID,
							CardID:          transferToAsset.ID,
							OrderNO:         transferOrderNO,
							CategoryID:      transferToAsset.CategoryID,
							Currency:        transferToAsset.Currency,
							Amount:          transferAmount,
							TransactionType: common.TRANSACTION_TYPE_ASSET_ADD,
							BillType:        common.BILL_TYPE_FINANCIAL_TRANSFER_RECEIVE_ADD,
						},
					}

					tErr = fs.assetTransactionDao.BatchTransaction(ctx, transactions, common.TRANSACTION_RECORD_TYPE_FINANCIAL_TRANSFER, false)
					if tErr != nil {
						logger.Warn("transaction failed,", tErr)
						tErr = utils.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
						return tErr
					}

					return nil
				})

				if tErr != nil {
					logger.Errorf("auto yield transaction failed, card: [%d], interest: [%s], transfer: [%s], %v", transferToAsset.ID, currencyIntererst, transferAmount, tErr)
					continue
				}

				if err != nil {
					logger.Errorf("auto yield transaction failed, card: [%d], interest: [%s], transfer: [%s], %v", transferToAsset.ID, currencyIntererst, transferAmount, err)
					continue
				}

				logger.Infof("financial transfered, card ID: [%d] -> [%d], order NO: [%s]", toCard.ID, transferToAsset.ID, transferOrderNO)
			}
		}
		userCurrent++
	}

	for _, c := range inputCurrencies {
		if count, ok := currencyCount[c]; !ok {
			logger.Errorf("%s INTEREST_DISTRIBUTE_FAIL:currency=%s", time.Now().Format(time.RFC3339), c.String())
			continue
		} else {
			logger.Infof("%s INTEREST_DISTRIBUTE:currency=%s,count=%d", time.Now().Format(time.RFC3339), c.String(), count)
		}
	}

	return nil
}

func (fs *FinancialService) CreateAutoYieldCard(ctx context.Context, userID uint64, currency common.Currency) (card *cardDao.Card, err error) {

	category, err := fs.categoryDao.GetByID(ctx, uint64(2000+currency))
	if err != nil {
		logger.Warn("get failed", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if category == nil {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_NO_SUCH_CATEGORY)
		return
	}
	if category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD == 0 {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CATEGORY)
		return
	}

	locker := utils.NewLocker()
	if err = locker.WaitLock(
		ctx,
		utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_AUTO_YIELD, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)),
		utils.Config.System.LockMicroseconds*time.Microsecond,
		utils.Config.System.LockWaitMicroseconds*time.Microsecond,
	); err != nil {
		logger.Warnf("lock failed: [%s], #v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_AUTO_YIELD, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)), err)
		return
	}
	defer func() {
		if err := locker.UnLock(ctx); err != nil {
			logger.Warnf("unlock %s failed, %v", utils.GetGlobalLockKey(common.LOCK_PURPOSE_CREATE_AUTO_YIELD, strconv.FormatUint(userID, 10), strconv.FormatUint(category.ID, 10)), err)
		}
	}()

	var id uint64

	var tErr error
	err = utils.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ctx = context.WithValue(ctx, "db", tx)

		card, tErr = fs.cardDao.GetByUserIDCategoryIDForUpdate(ctx, userID, uint64(2000+currency))
		if tErr != nil {
			logger.Warn("get failed", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}
		if card != nil {
			return nil
		}

		card = &cardDao.Card{
			UserID:        userID,
			Type:          common.ASSET_TYPE_AUTO_YIELD,
			CategoryID:    category.ID,
			ProductName:   currency.String(),
			Alias:         currency.String() + " investment",
			Currency:      currency,
			CurrencyType:  currency.Type(),
			Status:        common.CARD_STATUS_ACTIVATED,
			FreezeStatus:  common.CARD_FREEZE_STATUS_UNFROZEN,
			EarningStatus: common.CARD_EARNING_STATUS_DISABLED,
		}

		id, tErr = fs.cardDao.Save(ctx, card)
		if tErr != nil {
			logger.Warn("save failed,", tErr)
			tErr = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return tErr
		}

		if _, tErr = fs.assetDao.Save(ctx, &accountDao.Asset{
			ID:           id,
			UserID:       userID,
			Type:         common.ASSET_TYPE_AUTO_YIELD,
			CategoryID:   category.ID,
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
		logger.Warn("transaction failed,", tErr)
		err = utils.NewBusinessError(ctx, common.CODE_TRANSACTION_FAILED)
		return
	}

	if err != nil {
		logger.Warn("transaction failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	return

}

func (fs *FinancialService) AutoYieldInfo(ctx context.Context, form *entities.AutoYieldInfoForm, userID uint64) (
	dailyInterest decimal.Decimal,
	principalAmount decimal.Decimal,
	annualYieldRate decimal.Decimal,
	oneMonthAccumulatedInterest decimal.Decimal,
	twoMonthsAccumulatedInterest decimal.Decimal,
	threeMonthsAccumulatedInterest decimal.Decimal,
	earningStatus common.CardEarningStatus,
	cardID uint64,
	threshold decimal.Decimal,
	thresholdCurrency common.Currency,
	err error) {

	currency := common.Currency(0).FromString(form.Currency)
	if currency == 0 {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CURRENCY)
		return
	}

	category, err := fs.categoryDao.GetByID(ctx, uint64(2000+currency))
	if err != nil {
		logger.Warn("get failed", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if category == nil {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CURRENCY)
		return
	}
	if category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD == 0 {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CURRENCY)
		return
	}

	product, err := fs.financialProductDao.GetByCode(ctx, common.FINANCIAL_CODE_AUTO_YIELD)
	if err != nil {
		logger.Warn("get financial product failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if product == nil || product.Status != common.FINANCIAL_PRODUCT_STATUS_ACTIVE {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_PRODUCT_DISABLED)
		return
	}

	isSupported := false
	switch product.SupportType {
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_ALL:
		isSupported = true
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_CURRENCY:
		sups := strings.Split(product.SupportedCurrencies, ",")
		for _, s := range sups {
			if common.Currency(0).FromString(s) == currency {
				isSupported = true
				break
			}
		}
	}

	if !isSupported {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_PRODUCT_DISABLED)
		return
	}

	interestRates, err := fs.interestRateDao.List(ctx)
	if err != nil {
		logger.Warn("get interest rates failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	interestRateMap := make(map[common.Currency]*financialDao.InterestRate)
	for _, ir := range interestRates {
		interestRateMap[ir.Currency] = ir
	}

	interestRate, ok := interestRateMap[currency]
	if !ok {
		if product.DefaultYieldRate == nil {
			logger.Warn("get interest rates failed [%#v]", product)
			err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}
		interestRate = &financialDao.InterestRate{
			Rate: *product.DefaultYieldRate,
		}
	}

	annualYieldRate = interestRate.Rate

	param, err := fs.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_AUTO_YIELD_THRESHOLD)
	if err != nil {
		logger.Warn("get threshold failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if param == nil {
		logger.Warnf("no parameter: [%s]", common.PARAMETER_KEY_AUTO_YIELD_THRESHOLD)
		param = &systemDao.Parameter{
			Value: "100",
		}
	}

	threshold, err = decimal.NewFromString(param.Value)
	if err != nil {
		logger.Warn("parse threshold failed,", err)
		threshold = decimal.NewFromInt(100)
	}

	thresholdCurrency = common.CURRENCY_USD
	if ok && interestRate.Threshold != nil {
		threshold = *interestRate.Threshold
		thresholdCurrency = interestRate.Currency
	}

	thresholdRate, err := utils.GetExchangeRate(ctx, currency, thresholdCurrency)
	if err != nil {
		logger.Warn("get exchange rate failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if thresholdRate == nil {
		logger.Warnf("get exchange rate failed, %s, %s", currency, thresholdCurrency)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	isLeapYear := (time.Now().Year()%4 == 0 && time.Now().Year()%100 != 0) || time.Now().Year()%400 == 0
	hoursOfYear := decimal.NewFromInt(365 * 24)
	if isLeapYear {
		hoursOfYear = decimal.NewFromInt(366 * 24)
	}

	card, err := fs.cardDao.GetByUserIDCurrencyType(ctx, userID, currency, common.ASSET_TYPE_AUTO_YIELD)
	if err != nil {
		logger.Warn("get card failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if card == nil {
		card, err = fs.CreateAutoYieldCard(ctx, userID, currency)
		if err != nil {
			return
		}
	}

	cardID = card.ID

	earningStatus = card.EarningStatus

	user, err := fs.userDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get user failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if user == nil {
		err = utils.NewBusinessError(ctx, common.CODE_MALFORMED_DATA)
		return
	}

	lastTakenAt := time.Now().Truncate(time.Hour)
	firstTakenAt := time.Date(lastTakenAt.Year(), lastTakenAt.Month(), lastTakenAt.Day(), 0, 0, 0, 0, time.Local)

	dailyInterest = decimal.Zero

	fromCards, err := fs.cardDao.ListByUserIDCurrencyType(ctx, userID, currency, []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_CARD_PRODUCT, common.ASSET_TYPE_FIAT})
	if err != nil {
		logger.Warn("get failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	fromCardCategoryIDs := make([]uint64, 0, len(fromCards))
	for _, c := range fromCards {
		fromCardCategoryIDs = append(fromCardCategoryIDs, c.CategoryID)
	}
	categories, err := fs.categoryDao.ListByIDs(ctx, fromCardCategoryIDs)
	if err != nil {
		logger.Warn("get categories failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	fromCardCategoryMap := make(map[uint64]*accountDao.Category, len(categories))
	for _, c := range categories {
		fromCardCategoryMap[c.ID] = c
	}

	fromCardIDs := make([]uint64, 0, len(fromCards))
	assetSnapshotMap := make(map[uint64]map[int64]*financialDao.SnapshotBalance) // 卡ID 時間
	for _, c := range fromCards {
		if category, ok := fromCardCategoryMap[c.CategoryID]; !ok || category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD == 0 {
			continue
		}
		assetSnapshotMap[c.ID] = make(map[int64]*financialDao.SnapshotBalance)
		fromCardIDs = append(fromCardIDs, c.ID)
		for t := firstTakenAt; t.Before(lastTakenAt) || t.Equal(lastTakenAt); t = t.Add(time.Hour) {
			snapshot, err := fs.snapshotBalanceDao.Get(ctx, &financialDao.SnapshotBalance{
				CardID:        c.ID,
				FinancialCode: common.FINANCIAL_CODE_AUTO_YIELD,
				TakenAt:       t,
			})
			if err != nil {
				logger.Warn("get snapshot balance failed,", err)
				continue
			}

			if snapshot == nil {
				continue
			}

			assetSnapshotMap[c.ID][t.UnixMilli()] = snapshot
		}
	}

	for t := firstTakenAt; t.Before(lastTakenAt) || t.Equal(lastTakenAt); t = t.Add(time.Hour) {
		totalAmount := decimal.Zero
		for _, c := range fromCards {
			if category, ok := fromCardCategoryMap[c.CategoryID]; !ok || category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD == 0 {
				continue
			}
			s, ok := assetSnapshotMap[c.ID][t.UnixMilli()]
			if !ok {
				continue
			}
			if common.CardEarningStatus(0).FromString(s.EarningStatus) != common.CARD_EARNING_STATUS_ENABLED {
				continue
			}

			if s.Balance.LessThanOrEqual(decimal.Zero) {
				continue
			}
			totalAmount = totalAmount.Add(s.Balance)
		}
		if thresholdRate.Rate.IsZero() {
			logger.Errorf("threshold rate is zero: %v %v %v", thresholdRate.BaseCurrency, thresholdRate.QuoteCurrency, thresholdRate)
			err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}
		if totalAmount.Div(thresholdRate.Rate).LessThan(threshold) {
			continue
		}
		dailyInterest = dailyInterest.Add(totalAmount.Mul(interestRate.Rate.Div(hoursOfYear)))
	}

	crypto, err := fs.cryptoCurrencyDao.GetCryptoCurrencyByCurrencyType(ctx, currency)
	if err != nil {
		logger.Warn("get crypto currency failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if crypto == nil {
		logger.Warnf("no crypto currency: [%s]", currency)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	assets, err := fs.assetDao.ListByIDIn(ctx, fromCardIDs)
	if err != nil {
		logger.Warn("get failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	principalAmount = decimal.Zero
	for _, a := range assets {
		snapshot, err := fs.snapshotBalanceDao.Get(ctx, &financialDao.SnapshotBalance{
			CardID:        a.ID,
			FinancialCode: common.FINANCIAL_CODE_AUTO_YIELD,
			TakenAt:       lastTakenAt,
		})
		if err != nil {
			logger.Warn("get snapshot balance failed,", err)
			continue
		}

		if snapshot == nil {
			continue
		}

		// if snapshot.UserKYCLevel != string(common.KYC_LEVEL_2) && snapshot.UserKYCLevel != string(common.KYC_LEVEL_3) {
		// 	continue
		// }

		if common.CardEarningStatus(0).FromString(snapshot.EarningStatus) != common.CARD_EARNING_STATUS_ENABLED {
			continue
		}

		if snapshot.Balance.LessThanOrEqual(decimal.Zero) {
			continue
		}

		principalAmount = principalAmount.Add(snapshot.Balance)
	}

	if principalAmount.Div(thresholdRate.Rate).LessThan(threshold) {
		principalAmount = decimal.Zero
	}

	if dailyInterest.IsZero() {
		principalAmount = decimal.Zero
	}

	// if user.KycLevel == common.KYC_LEVEL_0 {
	// 	dailyInterest = decimal.Zero
	// 	principalAmount = decimal.Zero
	// }

	orders, err := fs.interestOrderDao.ListByCodeToCardIDCalculatedAt(ctx, common.FINANCIAL_CODE_AUTO_YIELD, card.ID, time.Now().AddDate(0, 0, -90))
	if err != nil {
		logger.Warn("get failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	oneMonthAccumulatedInterest, twoMonthsAccumulatedInterest, threeMonthsAccumulatedInterest = decimal.Zero, decimal.Zero, decimal.Zero
	for _, o := range orders {
		if o.CalculatedAt.After(time.Now().AddDate(0, 0, -30)) {
			oneMonthAccumulatedInterest = oneMonthAccumulatedInterest.Add(o.InterestAmount)
		}
		if o.CalculatedAt.After(time.Now().AddDate(0, 0, -60)) {
			twoMonthsAccumulatedInterest = twoMonthsAccumulatedInterest.Add(o.InterestAmount)
		}
		if o.CalculatedAt.After(time.Now().AddDate(0, 0, -90)) {
			threeMonthsAccumulatedInterest = threeMonthsAccumulatedInterest.Add(o.InterestAmount)
		}
	}
	oneMonthAccumulatedInterest = oneMonthAccumulatedInterest.RoundFloor(int32(crypto.DisplayDecimals))
	twoMonthsAccumulatedInterest = twoMonthsAccumulatedInterest.RoundFloor(int32(crypto.DisplayDecimals))
	threeMonthsAccumulatedInterest = threeMonthsAccumulatedInterest.RoundFloor(int32(crypto.DisplayDecimals))
	return

}

func (fs *FinancialService) AutoYieldHistory(ctx context.Context, form *entities.AutoYieldHistoryForm, userID uint64) (
	records []*entities.AutoYieldHistoryData,
	err error) {

	currency := common.Currency(0).FromString(form.Currency)
	if currency == 0 {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CURRENCY)
		return
	}

	card, err := fs.cardDao.GetByUserIDCurrencyType(ctx, userID, currency, common.ASSET_TYPE_AUTO_YIELD)
	if err != nil {
		logger.Warn("get card failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if card == nil {
		return
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	orders, err := fs.interestOrderDao.ListByCodeToCardIDCalculatedAtOrderByCalculatedAt(ctx, common.FINANCIAL_CODE_AUTO_YIELD, card.ID, today.AddDate(0, 0, -form.Period))
	if err != nil {
		logger.Warn("get failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	records = make([]*entities.AutoYieldHistoryData, form.Period)
	for i := range records {
		records[i] = &entities.AutoYieldHistoryData{
			Timestamp:       today.AddDate(0, 0, i-form.Period+1).UnixMilli(),
			Interest:        decimal.Zero,
			PrincipalAmount: decimal.Zero,
		}
	}
	for _, o := range orders {
		i := form.Period - 1 - int(today.Add(time.Hour*24-1).Sub(o.CalculatedAt)/time.Hour/24) // last moment of this day sub calculated at
		if i < 0 || i >= form.Period {
			logger.Warnf("wrong calculated at: %s", o.CalculatedAt)
			continue
		}
		records[i].Interest = records[i].Interest.Add(o.InterestAmount)
		if o.PrincipalAmount != nil {
			records[i].PrincipalAmount = records[i].PrincipalAmount.Add(*o.PrincipalAmount)
		}
	}

	crypto, err := fs.cryptoCurrencyDao.GetCryptoCurrencyByCurrencyType(ctx, currency)
	if err != nil {
		logger.Warn("get crypto currency failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if crypto == nil {
		logger.Warnf("no crypto currency: [%s]", currency)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	for i := range records {
		records[i].Interest = records[i].Interest.RoundFloor(int32(crypto.DisplayDecimals))
	}

	return
}

func (fs *FinancialService) AutoYieldEnable(ctx context.Context, form *entities.AutoYieldEnableForm, userID uint64) (err error) {

	card, err := fs.cardDao.GetByIDUserID(ctx, form.CardID, userID)
	if err != nil {
		logger.Warn("get card failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if card == nil {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_NO_SUCH_CARD)
		return
	}

	if card.Type != common.ASSET_TYPE_AUTO_YIELD {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_TARGET)
		return
	}

	usageEnable := false
	cards, err := fs.cardDao.ListByUserIDCurrencyType(ctx, userID, card.Currency, []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_CARD_PRODUCT, common.ASSET_TYPE_FIAT})
	if err != nil {
		logger.Warn("get failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	cardCategoryIDs := make([]uint64, 0, len(cards))
	for _, c := range cards {
		cardCategoryIDs = append(cardCategoryIDs, c.CategoryID)
	}
	categories, err := fs.categoryDao.ListByIDs(ctx, cardCategoryIDs)
	if err != nil {
		logger.Warn("get categories failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	cardCategoryMap := make(map[uint64]*accountDao.Category, len(categories))
	for _, c := range categories {
		cardCategoryMap[c.ID] = c
	}

	for _, c := range cards {
		category, ok := cardCategoryMap[c.CategoryID]
		if !ok {
			err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
			return
		}

		if category.Usage&common.CATEGORY_USAGE_USER_AUTO_YIELD > 0 {
			usageEnable = true
			break
		}
	}

	if !usageEnable {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_PRODUCT_DISABLED)
		return
	}

	product, err := fs.financialProductDao.GetByCode(ctx, common.FINANCIAL_CODE_AUTO_YIELD)
	if err != nil {
		logger.Warn("get financial product failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}
	if product == nil || product.Status != common.FINANCIAL_PRODUCT_STATUS_ACTIVE {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_PRODUCT_DISABLED)
		return
	}

	isSupported := false
	switch product.SupportType {
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_ALL:
		isSupported = true
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_CURRENCY:
		sups := strings.Split(product.SupportedCurrencies, ",")
		for _, s := range sups {
			if common.Currency(0).FromString(s) == card.Currency {
				isSupported = true
				break
			}
		}
	}

	if !isSupported {
		err = utils.NewBusinessError(ctx, common.CODE_FINANCIAL_PRODUCT_DISABLED)
		return
	}

	earningStatus := common.CardEarningStatus(0)
	if (card.EarningStatus == common.CARD_EARNING_STATUS_ENABLED || card.EarningStatus == 0) && *form.Enable == false {
		earningStatus = common.CARD_EARNING_STATUS_DISABLED
	} else if (card.EarningStatus == common.CARD_EARNING_STATUS_DISABLED || card.EarningStatus == 0) && *form.Enable == true {
		earningStatus = common.CARD_EARNING_STATUS_ENABLED
	} else {
		return
	}

	var rowsAffected int64
	rowsAffected, err = fs.cardDao.Update(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			ID: form.CardID,
		},
		Attrs: cardDao.Card{
			EarningStatus: earningStatus,
		},
	})
	if err != nil {
		logger.Warn("update card failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	if rowsAffected == 0 {
		logger.Warn("duplicate update card [%d]", form.CardID)
		return
	}
	return
}

func (fs *FinancialService) GetProduct(ctx context.Context, form *entities.FinancialGetProductForm) (
	product *financialDao.FinancialProduct,
	err error) {

	product, err = fs.financialProductDao.GetByCode(ctx, form.Code)
	if err != nil {
		logger.Warn("get financial product failed,", err)
		err = utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		return
	}

	return
}

func (fs *FinancialService) AutoYieldInterestList(ctx context.Context, userID uint64) ([]*entities.AutoYieldInterestVO, error) {
	product, err := fs.financialProductDao.GetByCode(ctx, common.FINANCIAL_CODE_AUTO_YIELD)
	if err != nil {
		logger.Warn(ctx, "get financial product failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if product == nil || product.Status != common.FINANCIAL_PRODUCT_STATUS_ACTIVE {
		return nil, nil
	}

	var supportedCurrencies []string

	switch product.SupportType {
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_ALL:
	case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_CURRENCY:
		supportedCurrencies = strings.Split(product.SupportedCurrencies, ",")
	}
	if len(supportedCurrencies) == 0 {
		return nil, nil
	}

	var resp []*entities.AutoYieldInterestVO
	for _, currency := range supportedCurrencies {
		crypto, err := fs.getCryptoCurrency(ctx, "", currency)
		if err != nil {
			return nil, err
		}
		if crypto == nil {
			return nil, utils.NewBusinessError(ctx, common.CODE_FINANCIAL_INVALID_CURRENCY)
		}
		vo := &entities.AutoYieldInterestVO{Currency: currency}
		dailyInterest, principalAmount, annualYieldRate, _, _, _, earningStatus, _, _, _, err := fs.AutoYieldInfo(ctx, &entities.AutoYieldInfoForm{Currency: currency}, userID)
		if err != nil {
			return nil, err
		}
		vo.EarningStatus = earningStatus.String()
		vo.AnnualYieldRate = annualYieldRate
		vo.DailyInterest = dailyInterest.String()
		vo.PrincipalAmount = principalAmount.StringFixed(int32(crypto.DisplayDecimals))
		resp = append(resp, vo)
	}

	return resp, nil
}

func (fs *FinancialService) getCryptoCurrency(ctx context.Context, mainnet string, currency string) (*coinsdoDao.CryptoCurrency, error) {
	cryptoCurrency, err := fs.cryptoCurrencyDao.GetCryptoCurrency(ctx, common.Mainnet(0).FromString(mainnet), currency)
	if err != nil {
		logger.Warn(ctx, "cryptoCurrency get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return cryptoCurrency, nil
}
