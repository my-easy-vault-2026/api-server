package services

import (
	systemDao "api-server/dao/system"
	"api-server/infra"
	"api-server/lib"
	"context"
	"encoding/json"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/redis/go-redis/v9"
)

type DefaultRateGetter struct {
	parameterDao *systemDao.ParameterDao
	logger       lib.Logger
	beBuilder    *lib.BEBuilder
	redis        infra.Redis
}

func NewDefaultRateGetter(parameterDao *systemDao.ParameterDao, logger lib.Logger, beBuilder *lib.BEBuilder, redis infra.Redis) *DefaultRateGetter {
	return &DefaultRateGetter{
		parameterDao: parameterDao,
		logger:       logger,
		beBuilder:    beBuilder,
		redis:        redis,
	}
}

func (rg *DefaultRateGetter) GetExchangeRate(ctx context.Context, quote common.Currency, base common.Currency) (*common.ExchangeRate, error) {

	redisKey := utils.GetRateRedisKey(quote.String())

	rate, err := rg.redis.HGet(ctx, redisKey, base.String()).Result()
	if err != nil {
		if err == redis.Nil {
			rg.logger.Errorf("GetExchangeRate: key %s not found", redisKey)
			return nil, rg.beBuilder.NewBusinessError(ctx, common.CODE_QUOTE_NO_SUCH_RATE)
		} else {
			rg.logger.Errorf("GetExchangeRate: key %s error: %v", redisKey, err)
			return nil, rg.beBuilder.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}

	var res *common.ExchangeRate
	err = json.Unmarshal([]byte(rate), &res)
	if err != nil {
		logger.Error("GetExchangeRate JSON Unmarshal error: ", err)
		return nil, err
	}

	return res, nil
}
