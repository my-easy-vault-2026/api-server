package web

import (
	"api-server/lib"
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
	userService *services.UserService
	authService *services.AuthService
	logger      lib.Logger
}

func NewUserHandler(userService *services.UserService,
	authService *services.AuthService,
	logger lib.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}

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

// @Param		request		body		entities.GetInfoForm	true	"body"
// @Param		X-Token		header		string					true	"X-Token"
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

	user, groupIDs, err := uh.userService.GetInfo(c, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	ret := &entities.GetInfoVO{}

	if err := copier.Copy(ret, user); err != nil {
		utils.ReError(c, err)
		return
	}
	ret.CreatedAt = user.CreatedAt.UnixMilli()
	ret.UpdatedAt = user.UpdatedAt.UnixMilli()
	ret.GroupIDs = groupIDs

	utils.ReData(c, ret)
}
