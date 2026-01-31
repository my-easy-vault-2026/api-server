package web

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type ExchangeHandler struct {
	exchangeService *services.ExchangeService
	logger          lib.Logger
}

func NewExchangeHandler(exchangeService *services.ExchangeService, logger lib.Logger) *ExchangeHandler {
	return &ExchangeHandler{
		exchangeService: exchangeService,
		logger:          logger,
	}
}

// @Summary		apply for a new ex preview. 兌換(deposit充值) 自己的數幣轉數幣
// @Description	Apply for a new ex preview.
// @Param			request			body		entities.ExchangePreviewForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ExchangePreviewVO		"data"
// @Router			/web/exchange/preview [post]
// @Tags			web/exchange
func (eh *ExchangeHandler) ExchangePreview(c *gin.Context) {
	form := &entities.ExchangePreviewForm{}

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

	preview, key, fromPlaces, toPlaces, inverseRate, err := eh.exchangeService.ExchangePreview(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	result := &entities.ExchangePreviewVO{}

	if err := copier.Copy(result, preview); err != nil {
		logger.Error("copy failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	result.Rate = make([]*entities.ExchangeRateVO, 0, len(preview.Rate))
	for _, rate := range preview.Rate {
		rateVO := &entities.ExchangeRateVO{
			BaseCurrency:  rate.BaseCurrency.String(),
			QuoteCurrency: rate.QuoteCurrency.String(),
			Rate:          rate.Rate,
			Timestamp:     rate.Timestamp.UnixMilli(),
			Purpose:       rate.Purpose.String(),
		}
		result.Rate = append(result.Rate, rateVO)
	}

	result.DisplayRate = entities.ExchangeRateVO{
		BaseCurrency:  preview.FromCurrency.String(),
		QuoteCurrency: preview.ToCurrency.String(),
		Rate:          inverseRate,
		Timestamp:     time.Now().UnixMilli(),
		Purpose:       common.RATE_PURPOSE_DISPLAY.String(),
	}

	switch common.GetAssetType(preview.FromCategoryID) {
	case common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT:
		result.FromCategory = preview.FromCurrency.String()
	}
	switch common.GetAssetType(preview.ToCategoryID) {
	case common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT:
		result.ToCategory = preview.ToCurrency.String()
	}

	result.FromAmount = preview.FromAmount.StringFixed(int32(fromPlaces))
	result.ToAmount = preview.ToAmount.StringFixed(int32(toPlaces))
	result.ExchangeFee = preview.ExchangeFee.StringFixed(int32(fromPlaces))
	result.FromCurrency = preview.FromCurrency.String()
	result.ToCurrency = preview.ToCurrency.String()
	result.Key = key
	result.ExpiredAt = preview.ExpiredAt.UnixMilli()

	utils.ReData(
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
