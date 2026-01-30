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
)

type WalletHandler struct {
	accountService        *services.AccountService
	walletService         *services.WalletService
	coinsdoService        *services.CoinsdoService
	cryptoCurrencyService *services.CryptoCurrencyService
	cardService           *services.CardService
}

func NewWalletHandler() *WalletHandler {
	return &WalletHandler{
		accountService:        services.NewAccountService(),
		walletService:         services.NewWalletService(),
		coinsdoService:        services.NewCoinsdoService(),
		cryptoCurrencyService: services.NewCryptoCurrencyService(),
		cardService:           services.NewCardService(),
	}
}

// @Summary		Get all wallets
// @Description	Get all wallets of the user
// @Tags			web/wallet
// @Param			request			body	entities.ListWalletsForm	true	"body"
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Param			X-Extend		header	string						false	"Extend"
// @Param			X-Convert		header	string						false	"Convert"
// @Router			/web/wallet/list [post]
func (wh *WalletHandler) ListWallets(c *gin.Context) {

	form := &entities.ListWalletsForm{}

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
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	wallets, err := wh.walletService.ListWallets(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.ListWalletsVO{
		Records: make([]*entities.WalletVO, len(wallets)),
	}

	for i, w := range wallets {
		res.Records[i] = &entities.WalletVO{}
		err := copier.Copy(res.Records[i], w)
		if err != nil {
			logger.Errorf("copy [%v] error, %v", w, err)
			utils.ReError(c, err)
		}
		res.Records[i].Currency = w.Currency.String()
		res.Records[i].CreatedAt = w.CreatedAt.UnixMilli()
		res.Records[i].UpdatedAt = w.UpdatedAt.UnixMilli()
	}

	utils.ReData(c, res)
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
