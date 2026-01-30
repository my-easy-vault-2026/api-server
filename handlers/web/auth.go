package web

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      lib.Logger
}

func NewAuthHandler(authServ *services.AuthService, logger lib.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authServ,
		logger:      logger,
	}
}

// @Param			request		body		entities.LoginOrCreateForm	true	"body"
// @Success		0			{object}	entities.LoginOrCreateVO	"data"
// @Param			X-Token		header		string						false	"回傳User token"
// @Param			X-ExpiredTs	header		string						false	"回傳失效時間"
// @Router			/web/user/loginOrRegister [post]
// @Description	Login or register a user. 如果有pin code，hasPinCode會是true
// @Tags			web/user
func (ah *AuthHandler) LoginOrRegister(c *gin.Context) {

	form := &entities.LoginOrCreateForm{}

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

	form.Platform = c.Request.Header.Get(common.HEADER_X_PLATFORM)
	form.DeviceID = c.Request.Header.Get(common.HEADER_X_DEVICE_ID)
	form.Ip = c.Request.Header.Get(common.HEADER_X_REAL_IP)
	form.AppVersion = c.Request.Header.Get(common.HEADER_X_APP_VERSION)

	user, userToken, expiredAt, _, err := ah.authService.LoginOrCreate(c, form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	hasPinCode := false
	if user.PinCode != "" {
		hasPinCode = true
	}

	res := &entities.LoginOrCreateVO{
		Token:      userToken,
		ExpiredAt:  expiredAt.UnixMilli(),
		HasPinCode: hasPinCode,
	}

	if err := copier.Copy(res, user); err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(
		c,
		res,
		map[string]string{
			common.HEADER_X_TOKEN:      userToken,
			common.HEADER_X_EXPIRED_TS: strconv.FormatInt(expiredAt.UnixMilli(), 10),
		},
	)
}

// @Summary		Logout
// @Description	Logout the user by invalidating their token
// @Tags			web/user
// @Accept			json
// @Produce		json
// @Param			X-Token	header	string	true	"User Token"
// @Router			/web/user/logout [post]
func (ah *AuthHandler) Logout(c *gin.Context) {

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

	token := c.Request.Header.Get("X-Token")
	deviceID := c.Request.Header.Get(common.HEADER_X_DEVICE_ID)

	err = ah.authService.Logout(c, token, deviceID, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	utils.ReData(c, nil)
}
