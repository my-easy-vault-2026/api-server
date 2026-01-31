package web

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type SystemHandler struct {
	systemService *services.SystemService
	logger        lib.Logger
}

func NewSystemHandler(systemService *services.SystemService, logger lib.Logger) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
		logger:        logger,
	}
}

// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ListSystemParametersVO	"data"
// @Router			/web/system/parameters [post]
// @Description	List all system parameters.
// @Tags			web/system
func (sh *SystemHandler) ListSystemParameters(c *gin.Context) {
	form := &entities.ListSystemParametersForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}
	parameters, err := sh.systemService.ListSystemParameters(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.ListSystemParametersVO{
		Records: make([]*entities.SystemParameterVO, len(parameters)),
	}
	for i, parameter := range parameters {
		res.Records[i] = &entities.SystemParameterVO{}
		err := copier.Copy(res.Records[i], parameter)
		if err != nil {
			logger.Errorf("copy [%v] error, %v", parameter, err)
			utils.ReError(c, err)
			return
		}
	}

	utils.ReData(c, res)
}

// @Param			X-Token			header		string				true	"User token"
// @Param			Accept-Language	header		string				false	"accept language"
// @Param			X-Extend		header		string				false	"Extend"
// @Param			X-Convert		header		string				false	"Convert"
// @Success		0				{object}	entities.CurrencyVO	"data"
// @Router			/web/system/currencies [post]
// @Description	List all currencies.
// @Tags			web/system
func (sh *SystemHandler) ListCurrencies(c *gin.Context) {
	form := &entities.ListCurrenciesForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}
	currencies, _, err := sh.systemService.ListCurrencies(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.ListCurrenciesVO{
		Records: make([]*entities.CurrencyVO, len(currencies)),
	}
	for i, currency := range currencies {
		res.Records[i] = &entities.CurrencyVO{}
		err := copier.Copy(res.Records[i], currency)
		if err != nil {
			logger.Errorf("copy [%v] error, %v", currency, err)
			utils.ReError(c, err)
			return
		}
	}

	utils.ReData(c, res)
}
