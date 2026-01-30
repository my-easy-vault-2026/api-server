package services

import (
	coinsdoDao "api-server/dao/coinsdo"
	systemDao "api-server/dao/system"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type SystemService struct {
	parameterDao         *systemDao.ParameterDao
	structuredContentDao *systemDao.StructuredContentDao
}

func NewSystemService() *SystemService {
	return &SystemService{
		parameterDao:         systemDao.NewParameterDao(),
		structuredContentDao: systemDao.NewStructuredContentDao(),
	}
}

// ListSystemParameters retrieves all system parameters.
func (ss *SystemService) ListSystemParameters(c *gin.Context, form *entities.ListSystemParametersForm) ([]*systemDao.Parameter, error) {

	parameters, err := ss.parameterDao.Gets(c, &systemDao.ParameterQuery{
		Parameter: systemDao.Parameter{
			Key:        form.Key,
			CategoryID: form.CategoryID,
		},
	})
	if err != nil {
		logger.Warn("get error,", err)
		return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
	}

	return parameters, nil
}

// ListCurrencies retrieves all currencies.
func (ss *SystemService) ListCurrencies(c *gin.Context, form *entities.ListCurrenciesForm) ([]common.Currency, []common.CurrencyType, error) {
	currencies := make([]common.Currency, 0)
	currencyTypes := make([]common.CurrencyType, 0)
	return currencies, currencyTypes, nil
}

func (ss *SystemService) CurrencyRatePingAndSwitch(ctx *gin.Context) error {
	param, err := ss.parameterDao.Get(ctx, &systemDao.ParameterQuery{
		Parameter: systemDao.Parameter{Key: common.PARAMETER_KEY_EXCHANGE_SOURCE},
	})
	if err != nil || param == nil {
		logger.Warnf("failed to get exchange rate source: %v", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	rateSource := common.ExchangeRateSource(param.Value)
	logger.Infof("current exchange rate source is %s", rateSource)

	// 當前是 bitop，失敗才切 binance
	if rateSource == common.EXCHANGE_RATE_SOURCE_BITOP {
		logger.Infof("bitop failed, switching to binance")
		_, updateErr := ss.parameterDao.Update(ctx, &systemDao.ParameterQuery{
			Parameter: systemDao.Parameter{ID: param.ID},
			Attrs:     systemDao.Parameter{Value: string(common.EXCHANGE_RATE_SOURCE_BINANCE)},
		})
		if updateErr != nil {
			return updateErr
		}
	}

	return nil
}

func retryUntilSuccess(fn func() error, retries int, delay time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil // 成功則提早返回
		}
		time.Sleep(delay)
	}
	return err // 最後一次錯誤返回
}

func (ss *SystemService) CurrencyRateProcess(c *gin.Context, currencies []*coinsdoDao.CryptoCurrency) {
	// 取得匯差設定
	param, err := ss.parameterDao.GetByKey(c, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)

	}
	if param == nil {
		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	}

	if param.ValueType == common.PARAMETER_VALUE_TYPE_AMOUNT {
		logger.Error("exchange value type cannot be PARAMETER_VALUE_TYPE_AMOUNT")
		return
	}

	exchangeFee, err := decimal.NewFromString(param.Value)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
		return
	}

	var currencyNames []string
	for _, v := range currencies {
		currencyNames = append(currencyNames, v.CurrencyName)
	}
	currencyNames = utils.SliceRemoveDuplicated(currencyNames, func(e1 string, e2 string) bool {
		return e1 == e2
	})

	param, err = ss.parameterDao.Get(c, &systemDao.ParameterQuery{Parameter: systemDao.Parameter{Key: common.PARAMETER_KEY_EXCHANGE_SOURCE}})
	if err != nil {
		logger.Warn("exchange rate source get failed:", err)
		err = utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
		return
	}
	if param == nil {
		logger.Warnf("exchange rate source no parameter: %s", common.PARAMETER_KEY_EXCHANGE_SOURCE)
		err = utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
		return
	}

	logger.Infof("fetch exchange rate matrix from source: %v", param.Value)
	rateMatrix := getExchangeRateMatrix(c, currencyNames, common.ExchangeRateSource(param.Value))

	for index, unit := range rateMatrix {
		rate := unit.Value.Mul(decimal.NewFromFloat(1).Add(exchangeFee))
		actualRate := unit.Value
		if common.CurrencyIsSameGroup(common.Currency(0).FromString(unit.Source), common.Currency(0).FromString(strings.ToLower(unit.Dest))) {
			rate = decimal.NewFromInt(1).Mul(decimal.NewFromFloat(1).Add(exchangeFee))
			actualRate = decimal.NewFromInt(1)
			rateMatrix[index].Value = decimal.NewFromInt(1)
		}

		exchangeRate := utils.ExchangeRate{
			BaseCurrency:  common.Currency(0).FromString(unit.Dest),
			QuoteCurrency: common.Currency(0).FromString(unit.Source),
			Rate:          rate,
			ActualRate:    actualRate,
			Timestamp:     time.Now(),
		}

		rateData, err := json.Marshal(exchangeRate)
		if err != nil {
			logger.Error("CurrencyRateProcess marshal ExchangeRate error,", err)
			continue
		}

		redisKey := utils.GetRateRedisKey(strings.ToUpper(unit.Dest))
		err = utils.RDB.HSet(c, redisKey, unit.Source, string(rateData)).Err()
		if err != nil {
			logger.Warn("CurrencyRateProcess 存储到 Redis 失败: %v", err)
		}
	}

	// 以 usd 為基礎的所有匯率
	for _, unit := range rateMatrix.GetBySource(strings.ToUpper(common.CURRENCY_USD.String())) {
		rate := unit.Value
		if common.CurrencyIsSameGroup(common.CURRENCY_USD, common.Currency(0).FromString(unit.Dest)) {
			rate = decimal.NewFromInt(1)
		}

		exchangeRate := utils.ExchangeRate{
			BaseCurrency:  common.CURRENCY_USD,
			QuoteCurrency: common.Currency(0).FromString(strings.ToLower(unit.Dest)),
			Rate: func() decimal.Decimal {
				if rate.Equal(decimal.Zero) {
					return decimal.Zero
				}
				return decimal.NewFromFloat(1).Div(rate)
			}(),
			Timestamp: time.Now(),
		}
		rateData, err := json.Marshal(exchangeRate)
		if err != nil {
			logger.Error("CurrencyRateProcess marshal ExchangeRate error,", err)
			continue
		}

		redisKey := utils.GetUsdBaseRateRedisKey(strings.ToUpper(common.CURRENCY_USD.String()))
		err = utils.RDB.HSet(c, redisKey, strings.ToUpper(unit.Dest), string(rateData)).Err()
		if err != nil {
			logger.Warn("CurrencyRateProcess 存储到 Redis 失败: %v", err)
		}

		// 反轉幣別
		reverseRate := utils.ExchangeRate{
			BaseCurrency:  common.Currency(0).FromString(strings.ToLower(unit.Dest)),
			QuoteCurrency: common.CURRENCY_USD,
			Rate:          rate,
			Timestamp:     time.Now(),
		}
		rateData, err = json.Marshal(reverseRate)
		if err != nil {
			logger.Error("CurrencyRateProcess marshal reverseRate error,", err)
			continue
		}
		redisKey = utils.GetUsdBaseRateRedisKey(strings.ToUpper(unit.Dest))
		err = utils.RDB.HSet(c, redisKey, strings.ToUpper(common.CURRENCY_USD.String()), string(rateData)).Err()
		if err != nil {
			logger.Warnf("CurrencyRateProcess 存储到 Redis 失败: %v", err)
		}
	}

	// 取出 usd base 的所有匯率，並計算出所有幣兌的匯率
	// 儲存已計算過的匯率對（避免重複計算）
	storeRates := make(map[string]bool)
	// 遍歷所有幣種，計算彼此之間的匯率
	for _, baseUnit := range rateMatrix.GetBySource(strings.ToUpper(common.CURRENCY_USD.String())) {
		for _, quoteUnit := range rateMatrix.GetBySource(strings.ToUpper(common.CURRENCY_USD.String())) {
			// 計算匯率
			key := baseUnit.Dest + "_" + quoteUnit.Dest
			revertKey := quoteUnit.Dest + "_" + baseUnit.Dest

			// 避免重複計算
			if _, exists := storeRates[key]; exists {
				continue
			}
			if _, exists := storeRates[revertKey]; exists {
				continue
			}
			//排除已有的匯率
			if baseUnit.Dest == common.CURRENCY_USD.String() || quoteUnit.Dest == common.CURRENCY_USD.String() {
				continue
			}

			baseUsdRate := rateMatrix.GetBySourceAndDest(baseUnit.Dest, strings.ToUpper(common.CURRENCY_USD.String()))
			if baseUsdRate == nil {
				continue
			}
			quoteUsdRate := rateMatrix.GetBySourceAndDest(quoteUnit.Dest, strings.ToUpper(common.CURRENCY_USD.String()))
			if quoteUsdRate == nil {
				continue
			}

			// 計算不同匯率
			rate := decimal.Zero
			if !baseUsdRate.Value.Equal(decimal.Zero) {
				rate = quoteUsdRate.Value.Div(baseUsdRate.Value)
			}
			computeRate := &utils.ExchangeRate{
				BaseCurrency:  common.Currency(0).FromString(baseUnit.Dest),
				QuoteCurrency: common.Currency(0).FromString(quoteUnit.Dest),
				Rate:          rate,
				Timestamp:     time.Now(),
			}
			storeRates[key] = true

			rateData, err := json.Marshal(computeRate)
			if err != nil {
				logger.Errorf("CurrencyRateProcess marshal reverseRate error. %v", err)
				continue
			}
			redisKey := utils.GetUsdBaseRateRedisKey(strings.ToUpper(baseUnit.Dest))
			err = utils.RDB.HSet(c, redisKey, strings.ToUpper(quoteUnit.Dest), string(rateData)).Err()
			if err != nil {
				logger.Warnf("CurrencyRateProcess 存储到 Redis 失败: %v", err)
			}

			convertRate := &utils.ExchangeRate{
				BaseCurrency:  common.Currency(0).FromString(quoteUnit.Dest),
				QuoteCurrency: common.Currency(0).FromString(baseUnit.Dest),
				Rate: func() decimal.Decimal {
					if rate.Equal(decimal.Zero) {
						return decimal.Zero
					}
					return decimal.NewFromFloat(1).Div(rate)
				}(),
				Timestamp: time.Now(),
			}
			storeRates[revertKey] = true

			rateData, err = json.Marshal(convertRate)
			if err != nil {
				logger.Errorf("CurrencyRateProcess marshal reverseRate error. %v", err)
				continue
			}
			redisKey = utils.GetUsdBaseRateRedisKey(strings.ToUpper(quoteUnit.Dest))
			err = utils.RDB.HSet(c, redisKey, strings.ToUpper(baseUnit.Dest), string(rateData)).Err()
			if err != nil {
				logger.Warnf("CurrencyRateProcess 存储到 Redis 失败: %v", err)
			}
		}
	}
}

func (ss *SystemService) GetSystemParameterByKey(c *gin.Context, key common.ParameterKey) (*systemDao.Parameter, error) {

	parameter, err := ss.parameterDao.GetByKey(c, key)
	if err != nil {
		logger.Error("get error,", err)
		return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
	}

	return parameter, nil
}

func (ss *SystemService) CheckPlatformVersion(c *gin.Context, form *entities.QueryVersionLimitForm) (isLimit bool) {
	key := common.ParameterKey(fmt.Sprintf("%s_version_limit", strings.ToLower(form.Platform)))
	par, err := ss.GetSystemParameterByKey(c, key)
	if err != nil || par == nil {
		return false
	}
	version := par.Value
	return version == form.Version
}

func (ss *SystemService) GetClientParams(c *gin.Context, form *entities.QueryVersionLimitForm) (paramsMap map[common.ParameterKey]string) {

	parArr, err := ss.parameterDao.ListByKeys(c, form.Keys)

	if err != nil {
		logger.Error("GetClientParams get error,", err)
		return
	}

	paramsMap = make(map[common.ParameterKey]string)
	for _, par := range parArr {
		// SecurityLevel < 90 代表才是给客户端的参数， >= 90 不能返回给用户
		if par.SecurityLevel < 90 {
			paramsMap[par.Key] = par.Value
		}
	}
	return
}

func (ss *SystemService) GetParams(c *gin.Context, keys []common.ParameterKey) (map[common.ParameterKey]string, error) {

	parArr, err := ss.parameterDao.ListByKeys(c, keys)

	if err != nil {
		logger.Error("GetClientParams get error,", err)
		return nil, err
	}

	paramsMap := make(map[common.ParameterKey]string)
	for _, par := range parArr {
		paramsMap[par.Key] = par.Value
	}
	return paramsMap, nil
}

func (ss *SystemService) ListStructuredContentBySceneCustomIDsLanguage(ctx context.Context, scene common.ContentScene, customIDs []string, language string) ([]*systemDao.StructuredContent, error) {

	if len(customIDs) == 0 {
		return make([]*systemDao.StructuredContent, 0), nil
	}

	cs, err := ss.structuredContentDao.ListBySceneCustomIDsLanguage(ctx, scene, customIDs, language)
	if err != nil {
		logger.Error("get error,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cs, nil
}

func (ss *SystemService) GetStructuredContentBySceneCustomIDLanguage(ctx context.Context, scene common.ContentScene, customID string, language string) (*systemDao.StructuredContent, error) {

	c, err := ss.structuredContentDao.GetBySceneCustomIDLanguage(ctx, scene, customID, language)
	if err != nil {
		logger.Error("get error,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return c, nil
}

func (ss *SystemService) GetStructuredContentBySceneCustomIDs(ctx context.Context, scene common.ContentScene, customIDs []string) (*systemDao.StructuredContent, error) {

	c, err := ss.structuredContentDao.GetBySceneCustomIDs(ctx, scene, customIDs)
	if err != nil {
		logger.Error("get error,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return c, nil
}

func getExchangeRateMatrix(ctx context.Context, currencies []string, source common.ExchangeRateSource) entities.RateResponseMatrix {
	// convert currency name.
	for i, v := range currencies {
		currencies[i] = utils.ToBittopCurrency(v)
	}

	matrix := entities.RateResponseMatrix{}
	matrix = getBitopExchangeRateMatrix(ctx, currencies)

	for i, v := range matrix {
		matrix[i].Source = utils.FromBittopCurrency(v.Source)
		matrix[i].Dest = utils.FromBittopCurrency(v.Dest)
	}

	currs := []common.Currency{common.CURRENCY_USD, common.CURRENCY_USDT, common.CURRENCY_USDC, common.CURRENCY_DAI}

	// Add fixed rate pair
	var findUsdUsdt, findUsdUsdc, findUsdDai, findUsdtUsd bool
	for i, v := range matrix {
		if v.Source == strings.ToUpper(common.CURRENCY_USD.String()) && v.Dest == strings.ToUpper(common.CURRENCY_USDT.String()) {
			matrix[i].Value = decimal.NewFromFloat(1.0)
			findUsdUsdt = true
		}
		if v.Source == strings.ToUpper(common.CURRENCY_USD.String()) && v.Dest == strings.ToUpper(common.CURRENCY_USDC.String()) {
			matrix[i].Value = decimal.NewFromFloat(1.0)
			findUsdUsdc = true
		}
		if v.Source == strings.ToUpper(common.CURRENCY_USD.String()) && v.Dest == strings.ToUpper(common.CURRENCY_DAI.String()) {
			matrix[i].Value = decimal.NewFromFloat(1.0)
			findUsdDai = true
		}
		if v.Source == strings.ToUpper(common.CURRENCY_USDT.String()) && v.Dest == strings.ToUpper(common.CURRENCY_USD.String()) {
			matrix[i].Value = decimal.NewFromFloat(1.0)
			findUsdtUsd = true
		}
	}
	if !findUsdUsdt {
		matrix = append(matrix, entities.RateResponseUnit{Source: strings.ToUpper(common.CURRENCY_USD.String()), Dest: strings.ToUpper(common.CURRENCY_USDT.String()), Value: decimal.NewFromFloat(1.0)})
	}
	if !findUsdUsdc {
		matrix = append(matrix, entities.RateResponseUnit{Source: strings.ToUpper(common.CURRENCY_USD.String()), Dest: strings.ToUpper(common.CURRENCY_USDC.String()), Value: decimal.NewFromFloat(1.0)})
	}
	if !findUsdDai {
		matrix = append(matrix, entities.RateResponseUnit{Source: strings.ToUpper(common.CURRENCY_USD.String()), Dest: strings.ToUpper(common.CURRENCY_DAI.String()), Value: decimal.NewFromFloat(1.0)})
	}
	if !findUsdtUsd {
		matrix = append(matrix, entities.RateResponseUnit{Source: strings.ToUpper(common.CURRENCY_USDT.String()), Dest: strings.ToUpper(common.CURRENCY_USD.String()), Value: decimal.NewFromFloat(1.0)})
	}

	// Add USD exchange rate
	matrix = func(m entities.RateResponseMatrix) entities.RateResponseMatrix {
		var ret entities.RateResponseMatrix
		for _, e := range matrix {
			if utils.SliceContain(currs, func(c common.Currency) bool { return strings.ToUpper(c.String()) == e.Dest }) {
				if !matrix.Exist(e.Source, strings.ToUpper(common.CURRENCY_USD.String())) {
					ret = append(ret, entities.RateResponseUnit{
						Source: e.Source,
						Dest:   strings.ToUpper(common.CURRENCY_USD.String()),
						Value:  e.Value,
					})
				}
			}
		}
		ret = append(ret, m...)
		return ret
	}(matrix)

	// add reverse pair.
	// add self pair.
	matrix = func(m entities.RateResponseMatrix) entities.RateResponseMatrix {
		var ret entities.RateResponseMatrix
		for _, e := range matrix {
			if !matrix.Exist(e.Dest, e.Source) {
				ret = append(ret, entities.RateResponseUnit{
					Source: e.Dest,
					Dest:   e.Source,
					Value: func() decimal.Decimal {
						if e.Value.Equal(decimal.Zero) {
							return decimal.Zero
						}
						return decimal.NewFromFloat(1).Div(e.Value)
					}(),
				})
			}
			if !matrix.Exist(e.Source, e.Source) && !ret.Exist(e.Source, e.Source) {
				ret = append(ret, entities.RateResponseUnit{
					Source: e.Source,
					Dest:   e.Source,
					Value:  decimal.NewFromFloat(1),
				})
			}
			if !matrix.Exist(e.Dest, e.Dest) && !ret.Exist(e.Dest, e.Dest) {
				ret = append(ret, entities.RateResponseUnit{
					Source: e.Dest,
					Dest:   e.Dest,
					Value:  decimal.NewFromFloat(1),
				})
			}
		}
		ret = append(ret, m...)
		return ret
	}(matrix)

	logMatrix := entities.RateResponseMatrix{}
	for _, v := range matrix {
		if v.Source == strings.ToUpper(common.CURRENCY_USD.String()) || v.Dest == strings.ToUpper(common.CURRENCY_USD.String()) ||
			v.Source == strings.ToUpper(common.CURRENCY_USDT.String()) || v.Dest == strings.ToUpper(common.CURRENCY_USDT.String()) {
			logMatrix = append(logMatrix, v)
		}
	}
	b, _ := json.Marshal(logMatrix)
	logger.Infof("exchange rate matrix: %s", string(b))

	return matrix
}

func getBitopExchangeRateMatrix(ctx context.Context, currencies []string) entities.RateResponseMatrix {
	srcCurrency := strings.Join(currencies, ",")
	var matrix entities.RateResponseMatrix

	for _, destCurrency := range currencies {
		func() {
			params := url.Values{}
			params.Add("src", srcCurrency)
			//params.Add("des", strings.ToUpper(utils.ToBittopCurrency(destCurrency)))
			params.Add("des", strings.ToUpper(destCurrency))
			params.Add("type", "1")
			finalURL := fmt.Sprintf("%s?%s", utils.Config.System.RateUrl, params.Encode())

			headers := http.Header{}
			headers.Set("Content-Type", "application/json")

			req, _ := http.NewRequest("GET", finalURL, nil)
			req.Header = headers
			req.WithContext(ctx)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Warnf("get bitop exchange rates error. src: %v, dest: %v %v\n", srcCurrency, destCurrency, err)
				return
			}
			defer resp.Body.Close()

			vo := entities.BitopRateResponseVO{}
			err = json.NewDecoder(resp.Body).Decode(&vo)
			if err != nil {
				logger.Warnf("get bitop exchange rates error. src: %v, dest: %v %v\n", srcCurrency, destCurrency, err)
				return
			}

			for _, v := range vo.List {
				matrix = append(matrix, entities.RateResponseUnit{
					Source: v.Name,
					Dest:   destCurrency,
					Value:  decimal.NewFromFloat(v.Value),
				})
			}
		}()
	}

	return matrix
}
