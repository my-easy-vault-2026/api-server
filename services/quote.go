package services

import (
	systemDao "api-server/dao/system"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/shopspring/decimal"
)

type QuoteService struct {
	parameterDao *systemDao.ParameterDao
	rateGetters  map[common.RatePurpose]IRateGetter
}

func NewQuoteService() *QuoteService {
	return &QuoteService{
		parameterDao: systemDao.NewParameterDao(),
		rateGetters: map[common.RatePurpose]IRateGetter{
			0: &defaultRateGetter{
				parameterDao: systemDao.NewParameterDao(),
			},
			common.RATE_PURPOSE_CHIPPAY_EXPRESS_BUY: &chippayExpressBuyRateGetter{
				parameterDao: systemDao.NewParameterDao(),
			},
		},
	}
}

type IRateGetter interface {
	GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error)
}

type defaultRateGetter struct {
	parameterDao *systemDao.ParameterDao
}
type chippayExpressBuyRateGetter struct {
	parameterDao *systemDao.ParameterDao
}

// ListExchangeRate retrieves the real-time exchange rate for a given currency pair.
func (qs *QuoteService) ListExchangeRate(ctx context.Context, form *entities.ListExchangeRateForm) ([]*utils.ExchangeRate, error) {

	quoteCurrencies := make([]common.Currency, len(form.QuoteCurrencies))
	for i := range form.QuoteCurrencies {
		quoteCurrencies[i] = common.Currency(0).FromString(form.QuoteCurrencies[i])
	}

	rates, err := utils.ListExchangeRate(ctx, common.Currency(0).FromString(form.BaseCurrency), quoteCurrencies)
	if err != nil {
		logger.Warn("get exchange rate failed,", err)
		return nil, err
	}

	if form.Purpose == common.RATE_PURPOSE_TOP_UP || form.Purpose == common.RATE_PURPOSE_TOP_DOWN {

		// 取得匯差設定
		param, err := qs.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		if err != nil {
			logger.Warn("CurrencyRateProcess exchange setting get failed,", err)

		}
		if param == nil {
			logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
		}

		exchangeFee, err := decimal.NewFromString(param.Value)
		if err != nil {
			logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
			return nil, err
		}

		for _, rate := range rates {
			if common.CurrencyIsSameGroup(rate.BaseCurrency, rate.QuoteCurrency) {
				rate.Rate = decimal.NewFromInt(1).Mul(decimal.NewFromFloat(1).Add(exchangeFee))
			}
		}
	}

	return rates, nil
}

// ListExchangeRate retrieves the real-time exchange rate for a given currency pair.
func (qs *QuoteService) GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error) {
	return qs.rateGetters[form.Purpose].GetExchangeRates(ctx, form)
}

func (rg *defaultRateGetter) GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error) {

	baseCurrencies := make([]common.Currency, len(form.BaseCurrencies))
	for i := range form.BaseCurrencies {
		baseCurrencies[i] = common.Currency(0).FromString(form.BaseCurrencies[i])
	}

	quoteCurrencies := make([]common.Currency, len(form.QuoteCurrencies))
	for i := range form.QuoteCurrencies {
		quoteCurrencies[i] = common.Currency(0).FromString(form.QuoteCurrencies[i])
	}

	// 臨時改動 之後請刪掉
	paramTopUpExchange, err := rg.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)
	}
	if paramTopUpExchange == nil {
		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE)
	}
	topUpExchangeFee, err := decimal.NewFromString(paramTopUpExchange.Value)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
		return nil, err
	}
	paramTopUp, err := rg.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_TOP_UP_TOP_UP_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)
	}
	if paramTopUp == nil {
		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_TOP_UP_TOP_UP_FEE)
	}
	topUpFee, err := decimal.NewFromString(paramTopUp.Value)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
		return nil, err
	}
	paramTopDownExchange, err := rg.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)
	}
	if paramTopDownExchange == nil {
		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_EXCHANGE_FEE)
	}
	topDownExchangeFee, err := decimal.NewFromString(paramTopDownExchange.Value)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
		return nil, err
	}
	paramTopDown, err := rg.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)
	}
	if paramTopDown == nil {
		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_FEE)
	}
	topDownFee, err := decimal.NewFromString(paramTopDown.Value)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
		return nil, err
	}
	// 臨時改動 之後請刪掉

	res := &entities.GetExchangeRateVO{}
	resData := make(map[string]*entities.ExchangeRateVO)
	// 遍歷所有幣種，計算彼此之間的匯率
	for _, base := range baseCurrencies {
		for _, quote := range quoteCurrencies {
			rate, err := utils.GetUsdBaseExchangeRate(ctx, base, quote)
			if err != nil {
				logger.Warn("get exchange rate failed,", err)
				continue
			}
			key := rate.BaseCurrency.String() + "_" + rate.QuoteCurrency.String()

			// 臨時改動 之後請刪掉
			otherFee := decimal.Zero
			if common.GetAssetType(uint64(base)) == common.ASSET_TYPE_CRYPTO && common.GetAssetType(uint64(quote)) == common.ASSET_TYPE_FIAT {
				otherFee = topUpFee.Add(topUpExchangeFee)
			}
			if common.GetAssetType(uint64(base)) == common.ASSET_TYPE_FIAT && common.GetAssetType(uint64(quote)) == common.ASSET_TYPE_CRYPTO {
				otherFee = topDownFee.Add(topDownExchangeFee)
			}
			// 臨時改動 之後請刪掉

			if rate.Rate.IsZero() {
				logger.Errorf("exchange rate is zero: %v %v %v", rate.BaseCurrency, rate.QuoteCurrency, rate.Rate)
				return nil, utils.NewBusinessError(ctx, common.CODE_QUOTE_NO_SUCH_RATE)
			}

			resData[key] = &entities.ExchangeRateVO{
				// 這裡其實是寫反了，因為rate應該是來源幣種購買詢價幣種的買價，不過前端要求我們這樣寫，所以就先給前端他們想要的數字
				QuoteCurrency: rate.BaseCurrency.String(),
				BaseCurrency:  rate.QuoteCurrency.String(),
				Rate:          decimal.NewFromInt(1).Div(rate.Rate).Mul(decimal.NewFromFloat(1).Sub(otherFee)),
				Timestamp:     rate.Timestamp.UnixMilli(),
			}

			// 臨時改動 之後請刪掉
			if common.GetAssetType(uint64(base)) == common.ASSET_TYPE_FIAT && common.GetAssetType(uint64(quote)) == common.ASSET_TYPE_CRYPTO {
				resData[key].Rate = decimal.NewFromInt(1).Div(rate.Rate)
			}
			// 臨時改動 之後請刪掉
		}
	}

	// TODO: fix me
	// etherfi exchange rate
	paramTopUpExchange, err = rg.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_TOP_UP_ETHERFI_TOP_UP_EXCHANGE_FEE)
	if err != nil {
		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)
		return nil, err
	}
	// 有設定才會計算
	if paramTopUpExchange != nil {
		topUpExchangeFee, err = decimal.NewFromString(paramTopUpExchange.Value)
		if err != nil {
			logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
			return nil, err
		}

		// 遍歷所有幣種，計算彼此之間的匯率
		for _, base := range baseCurrencies {
			for _, quote := range quoteCurrencies {
				rate, err := utils.GetUsdBaseExchangeRate(ctx, base, quote)
				if err != nil {
					logger.Warn("get exchange rate failed,", err)
					continue
				}
				key := rate.BaseCurrency.String() + "_" + rate.QuoteCurrency.String() + "ForEtherfi" // 前端定好的後綴

				// 臨時改動 之後請刪掉
				otherFee := decimal.Zero
				if common.GetAssetType(uint64(base)) == common.ASSET_TYPE_CRYPTO && common.GetAssetType(uint64(quote)) == common.ASSET_TYPE_FIAT {
					otherFee = topUpFee.Add(topUpExchangeFee)
				} else { // 不需要其他類的匯率
					continue
				}
				// 臨時改動 之後請刪掉

				if rate.Rate.IsZero() {
					logger.Errorf("exchange rate is zero: %v %v %v", rate.BaseCurrency, rate.QuoteCurrency, rate.Rate)
					return nil, utils.NewBusinessError(ctx, common.CODE_QUOTE_NO_SUCH_RATE)
				}

				resData[key] = &entities.ExchangeRateVO{
					// 這裡其實是寫反了，因為rate應該是來源幣種購買詢價幣種的買價，不過前端要求我們這樣寫，所以就先給前端他們想要的數字
					QuoteCurrency: rate.BaseCurrency.String(),
					BaseCurrency:  rate.QuoteCurrency.String(),
					Rate:          decimal.NewFromInt(1).Div(rate.Rate).Mul(decimal.NewFromFloat(1).Sub(otherFee)),
					Timestamp:     rate.Timestamp.UnixMilli(),
				}

				// 臨時改動 之後請刪掉
				if common.GetAssetType(uint64(base)) == common.ASSET_TYPE_FIAT && common.GetAssetType(uint64(quote)) == common.ASSET_TYPE_CRYPTO {
					resData[key].Rate = decimal.NewFromInt(1).Div(rate.Rate)
				}
				// 臨時改動 之後請刪掉
			}
		}
	}

	res.Records = resData

	// 如果是 top up 或 top down ，美元 usdt 匯兌需做調整
	// if form.Purpose == common.RATE_PURPOSE_TOP_UP || form.Purpose == common.RATE_PURPOSE_TOP_DOWN {

	// 	// 取得匯差設定
	// 	param, err := qs.parameterDao.GetByKey(ctx, common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	// 	if err != nil {
	// 		logger.Warn("CurrencyRateProcess exchange setting get failed,", err)

	// 	}
	// 	if param == nil {
	// 		logger.Warnf("CurrencyRateProcess exchange no parameter: [%s]", common.PARAMETER_KEY_EXCHANGE_EXCHANGE_FEE)
	// 	}

	// 	exchangeFee, err := decimal.NewFromString(param.Value)
	// 	if err != nil {
	// 		logger.Warn("CurrencyRateProcess exchange parameter parse failed,", err)
	// 		return nil, err
	// 	}

	// 	for _, rate := range rates {
	// 		if common.CurrencyIsSameGroup(rate.BaseCurrency, rate.QuoteCurrency) {
	// 			rate.Rate = decimal.NewFromInt(1).Mul(decimal.NewFromFloat(1).Add(exchangeFee))
	// 		}
	// 	}
	// }

	return res, nil
}

func (rg *chippayExpressBuyRateGetter) GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error) {

	baseCurrencies := make([]common.Currency, len(form.BaseCurrencies))
	for i := range form.BaseCurrencies {
		baseCurrencies[i] = common.Currency(0).FromString(form.BaseCurrencies[i])
	}

	quoteCurrencies := make([]common.Currency, len(form.QuoteCurrencies))
	for i := range form.QuoteCurrencies {
		quoteCurrencies[i] = common.Currency(0).FromString(form.QuoteCurrencies[i])
	}

	res := &entities.GetExchangeRateVO{}
	resData := make(map[string]*entities.ExchangeRateVO)
	// 遍歷所有幣種，計算彼此之間的匯率
	for _, base := range baseCurrencies {
		for _, quote := range quoteCurrencies {
			rate, err := utils.GetUsdBaseExchangeRate(ctx, base, quote)
			if err != nil {
				logger.Warn("get exchange rate failed,", err)
				continue
			}
			key := rate.BaseCurrency.String() + "_" + rate.QuoteCurrency.String()

			resData[key] = &entities.ExchangeRateVO{
				BaseCurrency:  rate.BaseCurrency.String(),
				QuoteCurrency: rate.QuoteCurrency.String(),
				Rate:          rate.Rate,
				Purpose:       common.RATE_PURPOSE_CHIPPAY_EXPRESS_BUY.String(),
				Timestamp:     rate.Timestamp.UnixMilli(),
			}
		}
	}
	res.Records = resData

	return res, nil
}
