package web

import (
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      lib.Logger
	beBuilder   *lib.BEBuilder
	httpRes     *lib.HttpRes
}

func NewAuthHandler(authServ *services.AuthService, logger lib.Logger, beBuilder *lib.BEBuilder, httpRes *lib.HttpRes) *AuthHandler {
	return &AuthHandler{
		authService: authServ,
		logger:      logger,
		beBuilder:   beBuilder,
		httpRes:     httpRes,
	}
}

// @Param			request		body		entities.LoginOrCreateForm	true	"body"
// @Success		0			{object}	entities.LoginOrCreateVO	"data"
// @Param			X-Token		header		string						false	"回傳User token"
// @Param			X-ExpiredTs	header		string						false	"回傳失效時間"
// @Router			/web/user/loginOrRegister [post]
// @Tags			web/user
func (ah *AuthHandler) LoginOrRegister(c *gin.Context) {

	form := &entities.LoginOrCreateForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, ah.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	user, userToken, expiredAt, err := ah.authService.LoginOrCreate(c, form.Email, form.PINCode)

	if err != nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	res := &entities.LoginOrCreateVO{}

	if err := copier.Copy(res, user); err != nil {
		ah.httpRes.ReError(c, http.StatusInternalServerError, err)
		return
	}

	ah.httpRes.ReData(
		c,
		res,
		map[string]string{
			common.HEADER_X_TOKEN:      userToken,
			common.HEADER_X_EXPIRED_TS: strconv.FormatInt(expiredAt.UnixMilli(), 10),
		},
	)
}

// @Tags			web/user
// @Param			X-Token	header	string	true	"User Token"
// @Router			/web/user/logout [post]
func (ah *AuthHandler) Logout(c *gin.Context) {

	userIDAny, ok := c.Get(common.HEADER_X_UID)
	if !ok {
		ah.logger.Error("no X-Uid")
		ah.httpRes.ReError(c, http.StatusBadRequest, ah.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		ah.logger.Error("X-Uid parse failed,", userIDAny)
		ah.httpRes.ReError(c, http.StatusBadRequest, ah.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	token := c.Request.Header.Get("X-Token")

	err := ah.authService.Logout(c, token, userID)
	if err != nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}
	ah.httpRes.ReData(c, nil)
}
