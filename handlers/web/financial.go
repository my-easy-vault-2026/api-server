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
)

type FinancialHandler struct {
	financialService *services.FinancialService
	systemService    *services.SystemService
	coinsdoService   *services.CoinsdoService
}

func NewFinancialHandler() *FinancialHandler {
	return &FinancialHandler{
		financialService: services.NewFinancialService(),
		systemService:    services.NewSystemService(),
		coinsdoService:   services.NewCoinsdoService(),
	}
}

// @Param			request			body		entities.AutoYieldInfoForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.AutoYieldInfoVO		"data"
// @Router			/web/financial/autoYield/info [post]
// @Tags			web/financial
func (eh *FinancialHandler) AutoYieldInfo(c *gin.Context) {
	form := &entities.AutoYieldInfoForm{}

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

	dailyInterest,
		principalAmount,
		annualYieldRate,
		oneMonthAccumulatedInterest,
		twoMonthsAccumulatedInterest,
		threeMonthsAccumulatedInterest,
		earningStatus,
		cardID,
		threshold,
		thresholdCurrency,
		err := eh.financialService.AutoYieldInfo(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	crypto, err := eh.coinsdoService.GetCryptoCurrency(c, "", form.Currency)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if crypto == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_FINANCIAL_INVALID_CURRENCY))
		return
	}

	vo := &entities.AutoYieldInfoVO{
		DailyInterest:                  dailyInterest.String(),
		PrincipalAmount:                principalAmount.StringFixed(int32(crypto.DisplayDecimals)),
		AnnualYieldRate:                annualYieldRate,
		OneMonthAccumulatedInterest:    oneMonthAccumulatedInterest.StringFixed(int32(crypto.DisplayDecimals)),
		TwoMonthsAccumulatedInterest:   twoMonthsAccumulatedInterest.StringFixed(int32(crypto.DisplayDecimals)),
		ThreeMonthsAccumulatedInterest: threeMonthsAccumulatedInterest.StringFixed(int32(crypto.DisplayDecimals)),
		EarningStatus:                  earningStatus.String(),
		CardID:                         cardID,
		Threshold:                      threshold,
		ThresholdCurrency:              thresholdCurrency.String(),
	}

	utils.ReData(
		c,
		vo,
	)
}

// @Param			request			body		entities.AutoYieldHistoryForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.AutoYieldHistoryVO		"data"
// @Router			/web/financial/autoYield/history [post]
// @Tags			web/financial
func (eh *FinancialHandler) AutoYieldHistory(c *gin.Context) {
	form := &entities.AutoYieldHistoryForm{}

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

	records, err := eh.financialService.AutoYieldHistory(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	crypto, err := eh.coinsdoService.GetCryptoCurrency(c, "", form.Currency)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if crypto == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_FINANCIAL_INVALID_CURRENCY))
		return
	}

	res := &entities.AutoYieldHistoryVO{
		Records: make([]*entities.AutoYieldHistoryDataVO, len(records)),
	}

	for i, r := range records {
		res.Records[i] = &entities.AutoYieldHistoryDataVO{
			Timestamp:       r.Timestamp,
			PrincipalAmount: r.PrincipalAmount.StringFixed(int32(crypto.DisplayDecimals)),
			Interest:        r.Interest.StringFixed(int32(crypto.DisplayDecimals)),
		}
	}

	utils.ReData(
		c,
		res,
	)
}

// @Param			request			body		entities.AutoYieldEnableForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Router			/web/financial/autoYield/enable [post]
// @Tags			web/financial
func (eh *FinancialHandler) AutoYieldEnable(c *gin.Context) {
	form := &entities.AutoYieldEnableForm{}

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

	err = eh.financialService.AutoYieldEnable(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(
		c,
		nil,
	)
}

// AutoYieldList lists all auto yield card information
//
//	@Param		X-Token			header		string							true	"User token"
//	@Param		Accept-Language	header		string							false	"accept language"
//	@Success	0				{object}	entities.AutoYieldInterestList	"data"
//	@Router		/web/financial/autoYield/list [post]
//	@Tags		web/financial
func (eh *FinancialHandler) AutoYieldList(c *gin.Context) {
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

	data, err := eh.financialService.AutoYieldInterestList(c, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	isActivated := false
	for _, v := range data {
		if common.CardEarningStatus(0).FromString(v.EarningStatus) == 1 {
			isActivated = true
		}
	}

	resp := entities.AutoYieldInterestList{
		IsEarning: isActivated,
		Records:   data,
	}

	utils.ReData(c, &resp)
}
