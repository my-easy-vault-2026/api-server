package web

import (
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"

	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	accountService *services.AccountService
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountService: services.NewAccountService(),
	}
}

// @Param			request			body		entities.ListAssetsForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Success		0				{object}	entities.ListAssetsVO	"data"
// @Router			/web/account/assets/list [post]
// @Description	list assets
// @Tags			web/account
func (ah *AccountHandler) ListAssets(c *gin.Context) {
	form := &entities.ListAssetsForm{}

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

	assets, err := ah.accountService.ListAssets(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.ListAssetsVO{
		Records: make([]*entities.AssetVO, len(assets)),
	}
	for i, asset := range assets {
		res.Records[i] = &entities.AssetVO{}
		err := copier.Copy(res.Records[i], asset)
		if err != nil {
			logger.Errorf("copy [%v] error, %v", asset, err)
			utils.ReError(c, err)
			return
		}
		res.Records[i].Type = asset.Type.String()
		res.Records[i].Currency = asset.Currency.String()
		res.Records[i].CurrencyType = asset.CurrencyType.String()
		res.Records[i].CreatedAt = asset.CreatedAt.UnixMilli()
		res.Records[i].UpdatedAt = asset.UpdatedAt.UnixMilli()
	}

	utils.ReData(c, res)
}

// @Param			request			body		entities.GetEquivalentValueForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Success		0				{object}	entities.GetEquivalentValueVO	"data"
// @Router			/web/account/assets/getEquivalent [post]
// @Description	list assets
// @Tags			web/account
func (ah *AccountHandler) GetEquivalentAsset(c *gin.Context) {
	form := &entities.GetEquivalentValueForm{}

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

	total, frozenTotal, fees, rates, err := ah.accountService.GetEquivalentValue(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.GetEquivalentValueVO{
		Amount:       total,
		FrozenAmount: frozenTotal,
		Fees:         make(map[string]decimal.Decimal),
		Rates:        make([]*entities.ExchangeRateVO, 0, len(rates)),
		ActualRates:  make([]*entities.ExchangeRateVO, 0, len(rates)),
	}
	for k, v := range fees {
		res.Fees[k.String()] = v
	}
	for _, rate := range rates {
		rateVO := &entities.ExchangeRateVO{
			BaseCurrency:  rate.BaseCurrency.String(),
			QuoteCurrency: rate.QuoteCurrency.String(),
			Rate:          rate.Rate,
			Timestamp:     rate.Timestamp.UnixMilli(),
		}
		res.Rates = append(res.Rates, rateVO)
	}
	for _, rate := range rates {
		rateVO := &entities.ExchangeRateVO{
			BaseCurrency:  rate.BaseCurrency.String(),
			QuoteCurrency: rate.QuoteCurrency.String(),
			Rate:          rate.ActualRate,
			Timestamp:     rate.Timestamp.UnixMilli(),
		}
		res.ActualRates = append(res.ActualRates, rateVO)
	}

	utils.ReData(c, res)
}
