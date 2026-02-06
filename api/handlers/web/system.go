package web

import (
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type SystemHandler struct {
	systemService *services.SystemService
	logger        lib.Logger
	beBuilder     *lib.BEBuilder
	httpRes       *lib.HttpRes
}

func NewSystemHandler(systemService *services.SystemService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes,
) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
		logger:        logger,
		beBuilder:     beBuilder,
		httpRes:       httpRes,
	}
}

// @Param   key   query   string   false   "Parameter key"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.ListSystemParametersVO	"data"
// @Router			/web/system/parameters [get]
// @Description	List all system parameters.
// @Tags			web/system
func (sh *SystemHandler) ListSystemParameters(c *gin.Context) {
	form := &entities.ListSystemParametersForm{}

	err := c.ShouldBindQuery(form)

	if err != nil {
		sh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		sh.httpRes.ReError(c, http.StatusBadRequest, sh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}
	parameters, err := sh.systemService.ListSystemParameters(c, form)
	if err != nil {
		sh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	res := &entities.ListSystemParametersVO{
		Records: make([]*entities.SystemParameterVO, len(parameters)),
	}
	for i, parameter := range parameters {
		res.Records[i] = &entities.SystemParameterVO{}
		err := copier.Copy(res.Records[i], parameter)
		if err != nil {
			sh.logger.Errorf("copy [%v] error, %v", parameter, err)
			sh.httpRes.ReError(c, http.StatusInternalServerError, err)
			return
		}
	}

	sh.httpRes.ReData(c, res)
}
