package web

import (
	accountDao "api-server/dao/account"
	"api-server/services"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
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

type TransferHandler struct {
	transferService *services.TransferService
	notifyService   *services.NotifyService
	userService     *services.UserService
}

func NewTransferHandler() *TransferHandler {
	return &TransferHandler{
		transferService: services.NewTransferService(),
		notifyService:   services.NewNotifyService(),
		userService:     services.NewUserService(),
	}
}

// @Summary		apply for a new transfer preview. 轉帳(send發送) 自己的數幣轉別人的數幣
// @Description	Apply for a new transfer preview.
// @Param			request			body		entities.TransferPreviewForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ExchangePreviewVO		"data"
// @Router			/web/transfer/preview [post]
// @Tags			web/transfer
func (th *TransferHandler) TransferPreview(c *gin.Context) {
	form := &entities.TransferPreviewForm{}

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

	preview, key, fromPlaces, toPlaces, inverseRate, err := th.transferService.TransferPreview(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	result := &entities.TransferPreviewVO{}

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
	result.ExchangeFee = utils.DecPtrToStr(preview.ExchangeFee, fromPlaces)
	result.Fee = preview.Fee.StringFixed(int32(fromPlaces))
	result.FeeCurrency = preview.FeeCurrency.String()
	result.FromCurrency = preview.FromCurrency.String()
	result.ToCurrency = preview.ToCurrency.String()
	result.Mainnet = preview.Mainnet.String()
	result.Protocol = preview.Protocol.String()
	result.TransferKey = key
	result.ExpiredAt = preview.ExpiredAt.UnixMilli()

	utils.ReData(
		c,
		result,
	)
}

// @Summary		Apply for a new transfer confirmation. 轉帳(send發送) 自己的數幣轉別人的數幣
// @Param			request			body		entities.TransferConfirmForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.TransferConfirmVO		"data"
// @Router			/web/transfer/confirm [post]
// @Description	Apply for a new transfer confirmation.
// @Tags			web/transfer
func (th *TransferHandler) TransferConfirm(c *gin.Context) {
	// 讀取 request body 的資料並保存到一個變數中
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("failed to read request body")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	// 將讀取的資料重新設置回 request body，這樣後續還可以再次使用
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	form := &entities.TransferConfirmForm{}

	// 將 JSON 資料解析到結構體中
	err = json.Unmarshal(bodyBytes, &form)
	if err != nil {
		logger.Error("failed to unmarshal request body")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
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

	fromCategory, toCategory, err := th.transferService.GetCategory(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	orderNO, cardNotify, err := th.getTransferConfirmService(c, fromCategory, toCategory).TransferConfirm(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	// 餘額過低推播通知
	if cardNotify != nil && cardNotify.Card != nil {
		reqID, _ := utils.GetMDCValue("reqId")
		language := common.Language(0).FromString(c.Request.Header.Get("Accept-Language"))

		go func(userID uint64, language common.Language, cardNotify *entities.SendCardMsgData) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			email, err := th.userService.GetEmailByUserID(c, userID)
			if err != nil {
				logger.Warnf("get user email error, userID=%d, error:%v", userID, err)
				return
			}
			utils.SetMDCValue("reqId", reqID)
			cardNo := "••••" + cardNotify.Card.PANNumber[len(cardNotify.Card.PANNumber)-4:]

			notifyVO := &entities.NotifyVO{
				CardNo: cardNo,
			}
			sendVO := &entities.SendVO{
				Email:    email,
				Code:     cardNotify.MsgOPCode,
				UserID:   userID,
				Language: language,
			}

			err = th.notifyService.SendCardAmountLow(ctx, notifyVO, sendVO, cardNotify.Card.ID)
			if err != nil {
				logger.Warnf("send card amount low notify error, cardId=%d userID=%d, error:%v", cardNotify.Card.ID, userID, err)
			}
		}(userID, language, cardNotify)

	}

	utils.ReData(
		c,
		&entities.TransferConfirmVO{OrderNO: *orderNO},
	)
}

func (th *TransferHandler) getTransferConfirmService(ctx context.Context, fromCategory *accountDao.Category, toCategory *accountDao.Category) services.ITransferConfirm {
	return th.transferService
}
