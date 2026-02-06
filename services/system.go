package services

import (
	systemDao "api-server/dao/system"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
)

type SystemService struct {
	parameterDao *systemDao.ParameterDao
	logger       logger.Logger
	beBuilder    *lib.BEBuilder
}

func NewSystemService(
	parameterDao *systemDao.ParameterDao,
	logger logger.Logger,
	beBuilder *lib.BEBuilder,
) *SystemService {
	return &SystemService{
		parameterDao: parameterDao,
		logger:       logger,
		beBuilder:    beBuilder,
	}
}

// ListSystemParameters retrieves all system parameters.
func (ss *SystemService) ListSystemParameters(c *gin.Context, form *entities.ListSystemParametersForm) ([]*systemDao.Parameter, error) {

	parameters, err := ss.parameterDao.Gets(c, &systemDao.ParameterQuery{
		Parameter: systemDao.Parameter{
			Key: form.Key,
		},
	})
	if err != nil {
		ss.logger.Warn("get error,", err)
		return nil, ss.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
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

func (ss *SystemService) GetSystemParameterByKey(c *gin.Context, key common.ParameterKey) (*systemDao.Parameter, error) {

	parameter, err := ss.parameterDao.GetByKey(c, key)
	if err != nil {
		logger.Error("get error,", err)
		return nil, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR)
	}

	return parameter, nil
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
