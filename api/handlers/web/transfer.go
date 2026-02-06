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

type TransferHandler struct {
	transferService *services.TransferService
	logger          lib.Logger
	beBuilder       *lib.BEBuilder
	httpRes         *lib.HttpRes
}

func NewTransferHandler(transferService *services.TransferService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes,
) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
		logger:          logger,
		beBuilder:       beBuilder,
		httpRes:         httpRes,
	}
}

// @Param fromWalletId     query   string  true  "From wallet id"
// @Param toUserId     query   string  false  "to User id"
// @Param toEmail       query   string  false  "to email"
// @Param fromAmount       query   string  false "From amount (decimal)"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.ExchangePreviewVO		"data"
// @Router			/web/transfer/preview [get]
// @Tags			web/transfer
func (th *TransferHandler) TransferPreview(c *gin.Context) {
	form := &entities.TransferPreviewForm{}

	err := c.ShouldBindQuery(form)
	if err != nil {
		th.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.HEADER_X_UID)
	if !ok {
		th.logger.Error("no X-Uid")
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		th.logger.Error("X-Uid parse failed,", userIDAny)
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	preview, key, places, err := th.transferService.TransferPreview(c, form.FromAmount,
		form.FromWalletID,
		form.ToEmail,
		form.ToUserID,
		userID)
	if err != nil {
		th.httpRes.ReError(c, http.StatusInternalServerError, err)
		return
	}

	result := &entities.TransferPreviewVO{}

	if err := copier.Copy(result, preview); err != nil {
		th.logger.Error("copy failed,", err)
		th.httpRes.ReError(c, http.StatusInternalServerError, th.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	result.FromAmount = preview.FromAmount.StringFixed(int32(places))
	result.ToAmount = preview.ToAmount.StringFixed(int32(places))
	result.Fee = preview.Fee.StringFixed(int32(places))
	result.FeeCurrency = preview.FeeCurrency.String()
	result.FromCurrency = preview.FromCurrency.String()
	result.ToCurrency = preview.ToCurrency.String()
	result.Key = key
	result.ExpiredAt = preview.ExpiredAt.UnixMilli()

	th.httpRes.ReData(
		c,
		result,
	)
}

// @Param			request			body		entities.TransferConfirmForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.TransferConfirmVO		"data"
// @Router			/web/transfer/confirm [post]
// @Tags			web/transfer
func (th *TransferHandler) TransferConfirm(c *gin.Context) {
	form := &entities.TransferConfirmForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		th.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.HEADER_X_UID)
	if !ok {
		th.logger.Error("no X-Uid")
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		th.logger.Error("X-Uid parse failed,", userIDAny)
		th.httpRes.ReError(c, http.StatusBadRequest, th.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	orderNO, err := th.transferService.TransferConfirm(c, form.Key, form.PinCode, userID)
	if err != nil {
		th.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	th.httpRes.ReData(
		c,
		&entities.TransferConfirmVO{OrderNO: orderNO},
	)
}
