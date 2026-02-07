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

type UserHandler struct {
	userService *services.UserService
	authService *services.AuthService
	logger      lib.Logger
	beBuilder   *lib.BEBuilder
	httpRes     *lib.HttpRes
}

func NewUserHandler(userService *services.UserService,
	authService *services.AuthService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
		beBuilder:   beBuilder,
		httpRes:     httpRes,
	}
}

// @Param		request		body		entities.GetInfoForm	true	"body"
// @Param		X-Token		header		string					true	"X-Token"
// @Success	0			{object}	entities.GetInfoVO		"data"
// @Router		/web/user/:id [get]
// @Tags		web/user
func (uh *UserHandler) GetInfo(c *gin.Context) {

	id := c.Param("id")

	form := &entities.GetInfoForm{}

	err := c.ShouldBindQuery(form)

	if err != nil {
		uh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		uh.httpRes.ReError(c, http.StatusBadRequest, uh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.CTX_KEY_AUTH_UID)
	if !ok {
		uh.logger.Error("no X-Uid")
		uh.httpRes.ReError(c, http.StatusBadRequest, uh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		uh.logger.Error("X-Uid parse failed,", userIDAny)
		uh.httpRes.ReError(c, http.StatusBadRequest, uh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	if id != strconv.FormatUint(userID, 10) {
		uh.logger.Error("user id mismatch")
		uh.httpRes.ReError(c, http.StatusBadRequest, uh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	user, groupIDs, err := uh.userService.GetInfo(c, userID)
	if err != nil {
		uh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	ret := &entities.GetInfoVO{}

	if err := copier.Copy(ret, user); err != nil {
		uh.httpRes.ReError(c, http.StatusInternalServerError, err)
		return
	}
	ret.CreatedAt = user.CreatedAt.UnixMilli()
	ret.UpdatedAt = user.UpdatedAt.UnixMilli()
	ret.GroupIDs = groupIDs

	uh.httpRes.ReData(c, ret)
}
