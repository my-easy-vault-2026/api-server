package test

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	userDao "api-server/dao/user"
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AccountHandler struct {
	accountService *services.AccountService
	assetDao       *accountDao.AssetDao
	userDao        *userDao.UserDao
	cardDao        *cardDao.CardDao
	logger         lib.Logger
	beBuilder      *lib.BEBuilder
	httpRes        *lib.HttpRes
}

func NewAccountHandler(accountService *services.AccountService,
	assetDao *accountDao.AssetDao,
	userDao *userDao.UserDao,
	cardDao *cardDao.CardDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		assetDao:       assetDao,
		userDao:        userDao,
		cardDao:        cardDao,
		logger:         logger,
		beBuilder:      beBuilder,
		httpRes:        httpRes,
	}
}

// @Param			request			body		entities.ListAssetsForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Success		0				{object}	entities.ListAssetsVO	"data"
// @Router			/test/account/addAssets [post]
// @Tags			test/account
func (ah *AccountHandler) AddAssets(c *gin.Context) {
	form := &entities.AddAssetsForm{}

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

	var user *userDao.User
	user, err = ah.userDao.Get(c, &userDao.UserQuery{
		User: userDao.User{
			ID: form.UserID,
		},
	})
	if err != nil {
		ah.logger.Warn("get failed: ", err)
		ah.httpRes.ReError(c, http.StatusInternalServerError, ah.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	if user == nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, ah.beBuilder.NewBusinessError(c, common.CODE_NO_SUCH_USER))
		return
	}

	var asset *accountDao.Asset
	asset, err = ah.accountService.GetAssetByIDUserID(c, form.AssetID, user.ID)
	if err != nil {
		ah.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	err = ah.assetDao.AddAssets(
		c,
		asset.UserID,
		asset.ID,
		asset.CategoryID,
		asset.Currency,
		form.Amount,
	)
	if err != nil {
		ah.logger.Error("add failed,", err)
		ah.httpRes.ReError(c, http.StatusInternalServerError, ah.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	ah.httpRes.ReData(c, nil)
}
