package web

import (
	cardDao "api-server/dao/card"
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type OrderHandler struct {
	orderService   *services.OrderService
	coinsdoService *services.CoinsdoService
	walletService  *services.WalletService
	userService    *services.UserService
	logger         lib.Logger
	beBuilder      *lib.BEBuilder
	httpRes        *lib.HttpRes
}

func NewOrderHandler(orderService *services.OrderService,
	coinsdoService *services.CoinsdoService,
	walletService *services.WalletService,
	userService *services.UserService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		coinsdoService: coinsdoService,
		walletService:  walletService,
		userService:    userService,
		logger:         logger,
		beBuilder:      beBuilder,
		httpRes:        httpRes,
	}
}

// @Tags			web/order
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
		oh.httpRes.ReError(c, http.StatusBadRequest, oh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.HEADER_X_UID)
	if !ok {
		oh.logger.Error("no X-Uid")
		oh.httpRes.ReError(c, http.StatusBadRequest, oh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		oh.logger.Error("X-Uid parse failed,", userIDAny)
		oh.httpRes.ReError(c, http.StatusBadRequest, oh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	user, err := oh.userService.GetUserRole(c, userID, common.ROLE_USER)
	if err != nil {
		oh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	if user == nil {
		oh.httpRes.ReError(c, http.StatusBadRequest, oh.beBuilder.NewBusinessError(c, common.CODE_USER_NO_SUCH_USER))
		return
	}

	records, current, pageSize, total, err := oh.orderService.PageTransactionRecords(c, form.CardID,
		userID,
		form.Current,
		form.PageSize)
	if err != nil {
		oh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	decimals, err := oh.coinsdoService.ListDisplayDecimals(c)
	if err != nil {
		oh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}
	decimals[0] = 10 // 避免幣種為空

	cardIDs := make([]uint64, 0, len(records))
	for _, record := range records {
		cardIDs = append(cardIDs, record.FromCardID)
		cardIDs = append(cardIDs, record.ToCardID)
	}

	cards, err := oh.walletService.ListWalletByUserIDWalletID(c, userID, cardIDs)
	if err != nil {
		oh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}
	cardMap := make(map[uint64]*cardDao.Card)
	for _, card := range cards {
		cardMap[card.ID] = card
	}

	recordsCopy := make([]*entities.TransactionRecordVO, len(records))
	for i, v := range records {
		recordsCopy[i] = &entities.TransactionRecordVO{}
		err := copier.Copy(recordsCopy[i], v)
		if err != nil {
			logger.Warnf("copy [%v] error, %v", v, err)
			oh.httpRes.ReError(c, http.StatusInternalServerError, err)
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
		PageData: common.PageData[[]*entities.TransactionRecordVO]{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
			Records:  recordsCopy,
		},
	}

	oh.httpRes.ReData(
		c,
		result,
	)
}
