package web

import (
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type ExchangeHandler struct {
	exchangeService *services.ExchangeService
	logger          lib.Logger
	beBuilder       *lib.BEBuilder
	httpRes         *lib.HttpRes
}

func NewExchangeHandler(exchangeService *services.ExchangeService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes) *ExchangeHandler {
	return &ExchangeHandler{
		exchangeService: exchangeService,
		logger:          logger,
		beBuilder:       beBuilder,
		httpRes:         httpRes,
	}
}

// @Param			request			body		entities.ExchangePreviewForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.ExchangePreviewVO		"data"
// @Router			/web/exchange/preview [get]
// @Tags			web/exchange
func (eh *ExchangeHandler) ExchangePreview(c *gin.Context) {
	form := &entities.ExchangePreviewForm{}

	err := c.ShouldBindJSON(form)
	if err != nil {
		eh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		eh.httpRes.ReError(c, http.StatusBadRequest, eh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.HEADER_X_UID)
	if !ok {
		eh.logger.Error("no X-Uid")
		eh.httpRes.ReError(c, http.StatusBadRequest, eh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		eh.logger.Error("X-Uid parse failed,", userIDAny)
		eh.httpRes.ReError(c, http.StatusBadRequest, eh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	preview, key, fromPlaces, toPlaces, err := eh.exchangeService.ExchangePreview(c, 
		form.FromWalletID, 
		, userID)
	if err != nil {
		eh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	result := &entities.ExchangePreviewVO{}

	if err := copier.Copy(result, preview); err != nil {
		eh.logger.Error("copy failed,", err)
		eh.httpRes.ReError(c, http.StatusInternalServerError, eh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	result.Rate = make([]*entities.ExchangeRateVO, 0, len(preview.Rate))
	for _, rate := range preview.Rate {
		rateVO := &entities.ExchangeRateVO{
			Base:      rate.BaseCurrency.String(),
			Quote:     rate.QuoteCurrency.String(),
			Rate:      rate.Rate,
			Timestamp: rate.Timestamp.UnixMilli(),
			Purpose:   rate.Purpose.String(),
		}
		result.Rate = append(result.Rate, rateVO)
	}

	result.FromCategory = preview.FromCurrency.String()
	result.ToCategory = preview.ToCurrency.String()
	result.FromAmount = preview.FromAmount.StringFixed(int32(fromPlaces))
	result.ToAmount = preview.ToAmount.StringFixed(int32(toPlaces))
	result.ExchangeFee = preview.ExchangeFee.StringFixed(int32(fromPlaces))
	result.FromCurrency = preview.FromCurrency.String()
	result.ToCurrency = preview.ToCurrency.String()
	result.Key = key
	result.ExpiredAt = preview.ExpiredAt.UnixMilli()

	eh.httpRes.ReData(
		c,
		result,
	)
}

// @Summary		Apply for a new exchange confirmation.
// @Param			request			body		entities.ExchangeConfirmForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ExchangeConfirmVO		"data"
// @Router			/web/exchange/confirm [post]
// @Description	Apply for a new exchange confirmation.
// @Tags			web/exchange
func (eh *ExchangeHandler) ExchangeConfirm(c *gin.Context) {
	form := &entities.ExchangeConfirmForm{}

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

	userIDString := c.Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	orderNO, err := eh.exchangeService.ExchangeConfirm(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(
		c,
		entities.TransferConfirmVO{*orderNO},
	)
}
