package web

import (
	cardDao "api-server/dao/card"
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
// @Produce		json
// @Param			walletID		query	uint64	false	"Wallet ID"
// @Param			current			query	int		false	"Current page"
// @Param			pageSize		query	int		false	"Page size"
// @Param			X-Token			header	string	true	"User token"
// @Param			Accept-Language	header	string	false	"accept language"
// @Router			/web/order/transactions/page [get]
func (oh *OrderHandler) PageTransactionRecords(c *gin.Context) {
	form := &entities.PageTransactionRecordsForm{}

	// 從 query 參數綁定
	err := c.ShouldBindQuery(form)

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

	user, err := oh.userService.GetUserRole(c, userID, common.ROLE_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if user == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_USER_NO_SUCH_USER))
		return
	}

	records, current, pageSize, total, err := oh.orderService.PageTransactionRecords(c, form.CardID,
		userID,
		form.Current,
		form.PageSize)
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

	cardIDs := make([]uint64, 0, len(records))
	for _, record := range records {
		cardIDs = append(cardIDs, record.FromCardID)
		cardIDs = append(cardIDs, record.ToCardID)
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
		recordsCopy[i] = &entities.TransactionRecordVO{}
		err := copier.Copy(recordsCopy[i], v)
		if err != nil {
			logger.Warnf("copy [%v] error, %v", v, err)
			utils.ReError(c, err)
			return
		}
		recordsCopy[i].Type = v.Type.String()
		recordsCopy[i].IncomeCurrency = v.IncomeCurrency.String()
		recordsCopy[i].AgainstIncomeCurrency = v.AgainstIncomeCurrency.String()
		recordsCopy[i].Side = v.Side.String()
		recordsCopy[i].FromCurrency = v.FromCurrency.String()
		recordsCopy[i].ToCurrency = v.ToCurrency.String()
		recordsCopy[i].ExchangeFeeCurrency = v.ExchangeFeeCurrency.String()
		recordsCopy[i].TransferFeeCurrency = v.TransferFeeCurrency.String()
		recordsCopy[i].Status = v.Status.String()
		recordsCopy[i].CreatedAt = v.CreatedAt.UnixMilli()
		recordsCopy[i].UpdatedAt = v.UpdatedAt.UnixMilli()
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
