package test

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	userDao "api-server/dao/user"
	"api-server/services"
	"encoding/json"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	accountService *services.AccountService
	assetDao       *accountDao.AssetDao
	userDao        *userDao.UserDao
	cardDao        *cardDao.CardDao
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountService: services.NewAccountService(),
		assetDao:       accountDao.NewAssetDao(),
		userDao:        userDao.NewUserDao(),
		cardDao:        cardDao.NewCardDao(),
	}
}

// @Param			request			body		entities.ListAssetsForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Success		0				{object}	entities.ListAssetsVO	"data"
// @Router			/test/account/addAssets [post]
// @Description	add assets
// @Tags			test/account
func (ah *AccountHandler) AddAssets(c *gin.Context) {
	form := &entities.AddAssetsForm{}

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

	if form.Forwarding {
		url := "https://www.eusddev.com/api/test/account/addAssets"

		header := make(http.Header)
		header.Add("accept", "application/json")
		header.Add("content-type", "application/json")

		form.Forwarding = false

		data, err := json.Marshal(form)
		if err != nil {
			logger.Warn("marshal failed: ", err)
			return
		}

		resBody, resHeader, resCode, err := utils.HttpPost(c, url, string(data), header)
		if err != nil {
			logger.Warn("forawrd failed: ", err)
			return
		}

		logger.Infof("forawrd res: [%s][%#v][%d]", string(resBody), resHeader, resCode)
		utils.ReRaw(c, resCode, resBody)
		return
	}

	var user *userDao.User
	user, err = ah.userDao.Get(c, &userDao.UserQuery{
		User: userDao.User{
			Email: form.Email,
		},
	})
	if err != nil {
		logger.Warn("get failed: ", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	if user == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_USER_NO_SUCH_USER))
		return
	}

	var assets []*accountDao.Asset
	assets, err = ah.accountService.ListAssets(c, &entities.ListAssetsForm{}, user.ID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	for _, a := range assets {
		err = ah.assetDao.AddAssets(
			c,
			a.UserID,
			a.ID,
			a.CategoryID,
			a.Currency,
			decimal.NewFromInt(500000),
		)
		if err != nil {
			logger.Error("add failed,", err)
			utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
			return
		}
	}

	var cards []*cardDao.Card
	cards, err = ah.cardDao.Gets(c, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID: user.ID,
			Type:   common.ASSET_TYPE_CARD_PRODUCT,
		},
	})
	if err != nil {
		logger.Error("get failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	if len(cards) == 0 {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_CARD_INSUFFICIENT_FUNDS, "請先開卡後執行"))
	}

	res := make(map[string]interface{})
	for _, c := range cards {
		res[c.PreferredName] = c.IssueID
	}

	res["defaultCard"] = cards[0].IssueID

	utils.ReData(c, res)
}
