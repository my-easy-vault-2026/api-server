package web

import (
	cardDao "api-server/dao/card"
	"api-server/lib"
	"api-server/services"
	"bytes"
	"encoding/json"
	"io"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderHandler struct {
	orderService   *services.OrderService
	coinsdoService *services.CoinsdoService
	cardService    *services.CardService
	userService    *services.UserService
	logger         lib.Logger
}

func NewOrderHandler(orderService *services.OrderService, coinsdoService *services.CoinsdoService, cardService *services.CardService, userService *services.UserService, logger lib.Logger) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		coinsdoService: coinsdoService,
		cardService:    cardService,
		userService:    userService,
		logger:         logger,
	}
}

// @Summary		Page List Transaction Records
// @Description	Page List transaction records based on provided filters. <br> walletID和cardID和categoryID擇一放入
// @Tags			web/order
// @Accept			json
// @Produce		json
// @Param			request			body	entities.PageTransactionRecordsForm	true	"body"
// @Param			X-Token			header	string								true	"User token"
// @Param			Accept-Language	header	string								false	"accept language"
// @Param			X-Extend		header	string								false	"Extend"
// @Param			X-Convert		header	string								false	"Convert"
// @Router			/web/order/transactions/page [post]
func (oh *OrderHandler) PageTransactionRecords(c *gin.Context) {
	form := &entities.PageTransactionRecordsForm{}

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

	user, err := oh.userService.GetUserRole(c, &entities.GetUserForm{
		ID: userID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	// 取得 merchant id
	merchants, err := oh.userService.ListByEmailRole(c, user.Email, common.ROLE_MERCHANT_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	userIDs := []uint64{userID}
	for _, m := range merchants {
		userIDs = append(userIDs, m.ID)
	}

	records, current, pageSize, total, err := oh.orderService.PageTransactionRecords(c, form, userIDs)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	decimals, err := oh.coinsdoService.ListDisplayDecimals(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	decimals[0] = 10 // 避免幣種為空

	mainnetNames, err := oh.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	cardIDs := make([]uint64, 0, len(records))
	productIDs := make([]uint64, 0, len(records))
	rewardOrderNOs := make([]string, 0, len(records))
	for _, record := range records {
		cardIDs = append(cardIDs, record.FromCardID)
		cardIDs = append(cardIDs, record.ToCardID)
		if record.ToProductID != 0 {
			productIDs = append(productIDs, record.ToProductID)
			rewardOrderNOs = append(rewardOrderNOs, record.TransactionNO)
		}
	}

	cards, err := oh.cardService.ListCard(c, &entities.ListCardForm{
		IDIn: cardIDs,
	}, 0, 0)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	cardMap := make(map[uint64]*cardDao.Card)
	for _, card := range cards {
		cardMap[card.ID] = card
	}

	acceptLang := c.Request.Header.Get("Accept-Language")
	languages, err := utils.MatchLang(acceptLang)
	if err != nil {
		logger.Warn("match failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	if len(languages) == 0 {
		languages = []string{"en"}
	}

	recordsCopy := make([]*entities.TransactionRecordVO, len(records))
	for i, v := range records {
		recordsCopy[i] = oh.orderService.TransactionRecordToVO(
			c,
			v,
			mainnetNames,
			decimals,
			cardMap,
		)
	}

	result := &entities.PageTransactionRecordsVO{
		PageData: utils.PageData[[]*entities.TransactionRecordVO]{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
			Records:  recordsCopy,
		},
	}

	utils.ReData(
		c,
		result,
	)
}

// @Summary		Page List Transaction Records
// @Description	Page List transaction records based on provided filters. <br> walletID和cardID和categoryID擇一放入
// @Tags			web/order
// @Accept			json
// @Produce		json
// @Param			request			body	entities.PageAutoYieldTransactionRecordsForm	true	"body"
// @Param			X-Token			header	string								true	"User token"
// @Param			Accept-Language	header	string								false	"accept language"
// @Param			X-Extend		header	string								false	"Extend"
// @Param			X-Convert		header	string								false	"Convert"
// @Router			/web/order/transactions/autoYield/page [post]
func (oh *OrderHandler) PageAutoYieldTransactionRecords(c *gin.Context) {
	form := &entities.PageAutoYieldTransactionRecordsForm{}

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

	card, err := oh.cardService.GetFinancialCardByUserIDCurrency(c, userID, form.Currency)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	newForm := entities.PageTransactionRecordsForm{
		CardID: card.ID,
		Types:  form.Types,
		Page: utils.Page{
			Current:  form.Page.Current,
			PageSize: form.Page.PageSize,
		},
	}
	newBody, err := json.Marshal(newForm)
	if err != nil {
		logger.Warnf("fail to re compose body: %#v", newForm)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
	oh.PageTransactionRecords(c)

}

// @Summary		Get Transaction Record
// @Description	Get transaction record based on provided filters.
// @Tags			web/order
// @Accept			json
// @Produce		json
// @Param			request			body	entities.GetTransactionRecordForm	true	"body"
// @Param			X-Token			header	string								true	"User token"
// @Param			Accept-Language	header	string								false	"accept language"
// @Router			/web/order/transactions/get [post]
func (oh *OrderHandler) GetTransactionRecord(c *gin.Context) {
	form := &entities.GetTransactionRecordForm{}

	err := c.ShouldBindJSON(form)
	if err != nil {
		utils.ReError(c, err)
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

	user, err := oh.userService.GetUserRole(c, &entities.GetUserForm{
		ID: userID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	// 取得 merchant id
	merchants, err := oh.userService.ListByEmailRole(c, user.Email, common.ROLE_MERCHANT_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	userIDs := []uint64{userID}
	for _, m := range merchants {
		userIDs = append(userIDs, m.ID)
	}

	record, err := oh.orderService.GetTransactionRecord(c, form, userIDs)
	if err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	if record == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_ORDER_USER_HAS_NO_SUCH_ORDER))
		return
	}

	decimals, err := oh.coinsdoService.ListDisplayDecimals(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	decimals[0] = 10 // 避免幣種為空

	mainnetNames, err := oh.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	cardIDs := make([]uint64, 0, 2)
	cardIDs = append(cardIDs, record.FromCardID)
	cardIDs = append(cardIDs, record.ToCardID)

	cards, err := oh.cardService.ListCard(c, &entities.ListCardForm{
		IDIn: cardIDs,
	}, 0, 0)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	cardMap := make(map[uint64]*cardDao.Card)
	for _, card := range cards {
		cardMap[card.ID] = card
	}

	if record.ResponseCode != "" {
		acceptLang := c.Request.Header.Get("Accept-Language")
		record.FailReason = utils.GetErrorMsg(c, utils.GetReapDeclineCode(record.ResponseCode), acceptLang)
	}

	vo := oh.orderService.TransactionRecordToVO(
		c,
		record,
		mainnetNames,
		decimals,
		cardMap)

	utils.ReData(
		c,
		vo,
	)
}
