package web

import (
	"api-server/services"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type UserHandler struct {
	userService    *services.UserService
	authService    *services.AuthService
	accountService *services.AccountService
	walletService  *services.WalletService
	cardService    *services.CardService
	orderService   *services.OrderService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService:    services.NewUserService(),
		authService:    services.NewAuthService(),
		accountService: services.NewAccountService(),
		walletService:  services.NewWalletService(),
		cardService:    services.NewCardService(),
		orderService:   services.NewOrderService(),
	}
}

// @Param			request			body		entities.SetPinCodeForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Success		0				{object}	entities.SetPinCodeVO	"data"
// @Router			/web/user/setPinCode [post]
// @Description	Set user PIN code.
// @Tags			web/user
func (uh *UserHandler) SetPinCode(c *gin.Context) {

	form := &entities.SetPinCodeForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		vErrs := err.(validator.ValidationErrors)
		for _, vErr := range vErrs {
			if vErr.Field() == "PinCode" {
				utils.ReError(c, utils.NewBusinessError(c, common.CODE_USER_PIN_CODE_INVALID_FORMAT))
				return
			}
		}

		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	token := c.Request.Header.Get(common.HEADER_X_TOKEN)
	if token == "" {
		logger.Error("no X-Token")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	userIDString := c.Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	if err := uh.userService.SetPinCode(c, form, token, userID); err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}

// @Router			/web/user/resetPinCode [post]
// @Param			request			body		entities.ResetPinCodeForm	true	"body"
// @Param			X-Token			header		string						true	"User token"
// @Param			Accept-Language	header		string						false	"accept language"
// @Param			X-Extend		header		string						false	"Extend"
// @Param			X-Convert		header		string						false	"Convert"
// @Success		0				{object}	entities.SetPinCodeVO		"data"
// @Router			/web/user/resetPinCode [post]
// @Description	reset user PIN code.
// @Tags			web/user
func (uh *UserHandler) ResetPinCode(c *gin.Context) {
	// 讀取 request body 的資料並保存到一個變數中
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("failed to read request body")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	// 將讀取的資料重新設置回 request body，這樣後續還可以再次使用
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	form := &entities.ResetPinCodeForm{}

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

	if err := uh.userService.ResetPinCode(c, form); err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}

// @Param			request			body		entities.ForgotPinCodeForm	true	"body"
// @Param			X-Token			header		string						true	"User token"
// @Param			Accept-Language	header		string						false	"accept language"
// @Param			X-Extend		header		string						false	"Extend"
// @Param			X-Convert		header		string						false	"Convert"
// @Success		0				{object}	string		"data"
// @Router			/web/user/forgotPinCode [post]
// @Description	forget user PIN code.
// @Tags			web/user
func (uh *UserHandler) ForgotPinCode(c *gin.Context) {
	form := &entities.ForgotPinCodeForm{}

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

	token := c.Request.Header.Get(common.HEADER_X_TOKEN)
	if token == "" {
		logger.Error("no X-Token")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	userIDString := c.Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	if err := uh.userService.ForgotPinCode(c, form, userID); err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, nil)
}

// @Param		request		body		entities.GetInfoForm	true	"body"
// @Param		X-Token		header		string					true	"X-Token"
// @Param		X-Extend	header		string					false	"X-Extend"
// @Param		X-Convert	header		string					false	"X-Convert"
// @Success	0			{object}	entities.GetInfoVO		"data"
// @Failure	0			{object}	entities.GetInfoVO		"data"
// @Router		/web/user/getInfo [post]
// @Tags		web/user
func (uh *UserHandler) GetInfo(c *gin.Context) {

	form := &entities.GetInfoForm{}

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

	user, groupIDs, err := uh.userService.GetInfo(c, userID, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	ret := &entities.GetInfoVO{}

	if err := copier.Copy(ret, user); err != nil {
		utils.ReError(c, err)
		return
	}
	ret.Gender = user.Gender.String()
	ret.Auto3DS = user.Auto3DS.String()
	ret.AutoTopUp = user.AutoTopUp.String()
	ret.ATMToggle = user.ATMToggle.String()
	ret.CreatedAt = user.CreatedAt.UnixMilli()
	ret.UpdatedAt = user.UpdatedAt.UnixMilli()
	ret.ReferrerCode = user.PromotionCode
	ret.GroupIDs = groupIDs

	// 取得用戶 e卡，如果有 e卡 才返回 promotion資料
	cardForm := &entities.ListCardForm{
		AssetTypeIn: []common.AssetType{common.ASSET_TYPE_CARD_PRODUCT},
	}
	listCard, err := uh.cardService.ListCard(c, cardForm, common.CURRENCY_USD, userID)
	if err != nil {
		logger.Warnf("fail to get main card, ", err)
	}

	if len(listCard) == 0 {
		utils.ReData(c, ret)
		return
	}

	utils.ReData(c, ret)
}

// @Param		request		body		entities.SavePhoneNumberForm	true	"body"
// @Param		X-Token		header		string							true	"X-Token"
// @Param		X-Extend	header		string							false	"X-Extend"
// @Success	0			{object}	entities.SaveIdentityVO			"data"
// @Failure	0			{object}	entities.SaveIdentityVO			"data"
// @Router		/web/user/savePhoneNumber [post]
// @Tags		web/user
func (uh *UserHandler) SavePhoneNumber(c *gin.Context) {
	form := &entities.SavePhoneNumberForm{}

	err := c.ShouldBindJSON(form)

	logger.Info("save phone number form", form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	result, err := uh.userService.SavePhoneNumber(c, *form)
	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_UNKOWN_ERROR, err.Error()))
		return
	}

	utils.ReData(c, result)
}

// @Summary		刪除用戶帳戶
// @Description	刪除指定用戶的帳戶
// @Tags			web/user
// @Accept			json
// @Produce		json
// @Param			X-Token	header	string						true	"Token"
// @Param			body	body	entities.DeleteAccountForm	true	"Delete Account Form"
// @Router			/web/user/deleteAccount [post]
func (uh *UserHandler) DeleteAccount(c *gin.Context) {

	form := &entities.DeleteAccountForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	userIDString := c.Request.Header.Get(common.HEADER_X_UID)
	if userIDString == "" {
		logger.Error("no X-Uid")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	//刪除帳戶
	err = uh.userService.DeleteAccount(c, form, userID)
	if err != nil {
		logger.Error("fail to delete account", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_UNKOWN_ERROR, err.Error()))
		return
	}

	//清除token
	token := c.Request.Header.Get("X-Token")
	deviceID := c.Request.Header.Get(common.HEADER_X_DEVICE_ID)
	err = uh.authService.Logout(c, token, deviceID, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	logger.Infof("user %d deleted", c.GetInt64("uid"))

	utils.ReData(c, nil)
}

// @Param		request		body		entities.SaveLanguageForm	true	"body"
// @Param		X-Token		header		string							true	"X-Token"
// @Param		X-Extend	header		string							false	"X-Extend"
// @Router		/web/user/saveLanguage [post]
// @Tags		web/user
func (uh *UserHandler) SaveLanguage(c *gin.Context) {
	form := &entities.SaveLanguageForm{}

	err := c.ShouldBindJSON(form)

	logger.Info("save user language form", form)

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

	form.UserID = userID

	result, err := uh.userService.SaveLanguage(c, *form)
	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_UNKOWN_ERROR, err.Error()))
		return
	}

	utils.ReData(c, result)
}
