package web

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	cardService    *services.CardService
	accountService *services.AccountService
	coinsdoService *services.CoinsdoService
	userService    *services.UserService
	systemService  *services.SystemService
	walletService  *services.WalletService
	logger         lib.Logger
}

func NewWalletHandler(cardService *services.CardService,
	accountService *services.AccountService,
	coinsdoService *services.CoinsdoService,
	userService *services.UserService,
	systemService *services.SystemService,
	walletService *services.WalletService,
	logger lib.Logger) *WalletHandler {
	return &WalletHandler{
		cardService:    cardService,
		accountService: accountService,
		coinsdoService: coinsdoService,
		userService:    userService,
		systemService:  systemService,
		walletService:  walletService,
		logger:         logger,
	}
}

// @Summary		Get all wallets
// @Description	Get all wallets of the user
// @Tags			web/wallet
// @Param			request			body	entities.ListWalletsForm	true	"body"
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Router			/web/wallet/list [post]
func (wh *WalletHandler) ListWallets(c *gin.Context) {

	form := &entities.ListCardForm{}

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

	user, err := wh.userService.GetUserRole(c, userID, common.ROLE_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	if user == nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_USER_NO_SUCH_USER))
		return
	}

	cards, err := wh.cardService.ListWalletsByUserID(c, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if len(cards) == 0 {
		utils.ReData(
			c,
			&entities.ListCardVO{
				Records: make([]*entities.CardVO, 0),
			},
		)
		return
	}

	cardIDs := make([]uint64, 0, len(cards))
	categoryIDs := make([]uint64, 0, len(cards))
	for _, card := range cards {
		cardIDs = append(cardIDs, card.ID)
		categoryIDs = append(categoryIDs, card.CategoryID)

	}

	assets, err := wh.accountService.ListAssetsByIDInUserID(c, cardIDs, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if len(assets) != len(cardIDs) {
		logger.Error("asset wallet mismatch")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_CARD_WALLET_ASSET_MISMATCH))
		return
	}

	assetMap := map[uint64]decimal.Decimal{}
	for _, asset := range assets {
		assetMap[asset.ID] = asset.Amount.Copy()
	}

	result := &entities.ListCardVO{
		Records: make([]*entities.CardVO, len(cards)),
	}

	for i, card := range cards {

		result.Records[i] = &entities.CardVO{
			ID:            card.ID,
			UserID:        card.UserID,
			MerchantID:    card.MerchantID,
			CategoryID:    card.CategoryID,
			PreferredName: card.PreferredName,
			Type:          common.ASSET_TYPE_CARD_PRODUCT.String(),
			Amount:        assetMap[card.ID].Copy(),
			Issuer:        card.Issuer,
			Currency:      card.Currency.String(),
			Status:        card.Status.String(),
			CreatedAt:     card.CreatedAt.UnixMilli(),
			UpdatedAt:     card.UpdatedAt.UnixMilli(),
		}

	}

	utils.ReData(
		c,
		result,
	)
}

// @Summary     Create a new wallet
// @Description Create a new wallet for the user <br> 幣種列表： BTC ETH  USDT USDC DAI  WBTC  TRX  ADA  BCH  DOGE  LTC  XRP  SOL  BNB  ETC  MATIC
// @Tags            web/wallet
// @Param           request         body    entities.CreateWalletForm   true    "body"
// @Param           X-Token         header  string                      true    "User token"
// @Param           Accept-Language header  string                      false   "accept language"
// @Param           X-Extend        header  string                      false   "Extend"
// @Param           X-Convert       header  string                      false   "Convert"
// @Router          /web/wallet/apply [post]
func (wh *WalletHandler) CreateWallet(c *gin.Context) {
	form := &entities.CreateWalletForm{}
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
	walletID, err := wh.walletService.CreateWallet(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	res := &entities.CreateWalletVO{
		ID: walletID,
	}
	utils.ReData(c, res)
}
