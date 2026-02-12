package services

import (
	"github.com/my-easy-vault-2026/api-server/entities"
	"github.com/my-easy-vault-2026/shared-modules/common"

	systemDao "github.com/my-easy-vault-2026/api-server/dao/system"
	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/gin-gonic/gin"
)

type SystemService struct {
	parameterDao *systemDao.ParameterDao
	logger       lib.Logger
	beBuilder    *lib.BEBuilder
}

func NewSystemService(
	parameterDao *systemDao.ParameterDao,
	logger lib.Logger,
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
