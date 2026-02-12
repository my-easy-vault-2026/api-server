package web

import (
	"net/http"

	"github.com/my-easy-vault-2026/api-server/entities"
	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/lib"
	"github.com/my-easy-vault-2026/api-server/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	accountService *services.AccountService
	coinsdoService *services.CoinsdoService
	userService    *services.UserService
	systemService  *services.SystemService
	walletService  *services.WalletService
	logger         lib.Logger
	beBuilder      *lib.BEBuilder
	httpRes        *lib.HttpRes
}

func NewWalletHandler(
	accountService *services.AccountService,
	coinsdoService *services.CoinsdoService,
	userService *services.UserService,
	systemService *services.SystemService,
	walletService *services.WalletService,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
	httpRes *lib.HttpRes) *WalletHandler {
	return &WalletHandler{
		accountService: accountService,
		coinsdoService: coinsdoService,
		userService:    userService,
		systemService:  systemService,
		walletService:  walletService,
		logger:         logger,
		beBuilder:      beBuilder,
		httpRes:        httpRes,
	}
}

// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.ListCardCategoryVO		"data"
// @Router			/web/wallet/category [get]
// @Tags			web/wallet
func (ch *WalletHandler) ListCategory(c *gin.Context) {
	form := &entities.ListWalletCategoryForm{}

	err := c.ShouldBindQuery(form)
	if err != nil {
		ch.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		ch.httpRes.ReError(c, http.StatusBadRequest, ch.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	categories, err := ch.walletService.ListCategory(c)
	if err != nil {
		ch.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	result := &entities.ListCategoryVO{
		Records: make([]*entities.CategoryVO, len(categories)),
	}

	for i, category := range categories {
		categoryCopy := &entities.CategoryVO{}
		if err := copier.Copy(&categoryCopy, &category); err != nil {
			ch.httpRes.ReError(c, http.StatusInternalServerError, err)
			return
		}
		categoryCopy.Type = category.Type.String()
		categoryCopy.Currency = category.Currency.String()
		categoryCopy.CurrencyType = category.CurrencyType.String()
		categoryCopy.CreatedAt = category.CreatedAt.UnixMilli()
		categoryCopy.UpdatedAt = category.UpdatedAt.UnixMilli()

		result.Records[i] = categoryCopy
	}

	ch.httpRes.ReData(
		c,
		result,
	)
}

// @Tags			web/wallet
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Router			/web/wallet [get]
func (wh *WalletHandler) ListWallets(c *gin.Context) {

	form := &entities.ListWalletForm{}

	err := c.ShouldBindQuery(form)

	if err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(form); err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.CTX_KEY_AUTH_UID)
	if !ok {
		wh.logger.Error("no X-Uid")
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		wh.logger.Error("X-Uid parse failed,", userIDAny)
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	wallets, err := wh.walletService.ListWalletsByUserID(c, userID)
	if err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	if len(wallets) == 0 {
		wh.httpRes.ReData(
			c,
			&entities.ListWalletsVO{
				Records: make([]*entities.WalletVO, 0),
			},
		)
		return
	}

	cardIDs := make([]uint64, 0, len(wallets))
	categoryIDs := make([]uint64, 0, len(wallets))
	for _, wallet := range wallets {
		cardIDs = append(cardIDs, wallet.ID)
		categoryIDs = append(categoryIDs, wallet.CategoryID)
	}

	assets, err := wh.accountService.ListAssetsByIDInUserID(c, cardIDs, userID)
	if err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	if len(assets) != len(cardIDs) {
		wh.logger.Error("asset wallet mismatch")
		wh.httpRes.ReError(c, http.StatusInternalServerError, wh.beBuilder.NewBusinessError(c, common.CODE_MALFORMED_DATA))
		return
	}

	assetMap := map[uint64]decimal.Decimal{}
	for _, asset := range assets {
		assetMap[asset.ID] = asset.Amount.Copy()
	}

	result := &entities.ListWalletsVO{
		Records: make([]*entities.WalletVO, len(wallets)),
	}

	for i, wallet := range wallets {

		result.Records[i] = &entities.WalletVO{
			ID:         wallet.ID,
			UserID:     wallet.UserID,
			CategoryID: wallet.CategoryID,
			Amount:     assetMap[wallet.ID].Copy(),
			Nation:     wallet.Nation,
			Currency:   wallet.Currency.String(),
			Status:     wallet.Status,
			CreatedAt:  wallet.CreatedAt.UnixMilli(),
			UpdatedAt:  wallet.UpdatedAt.UnixMilli(),
		}

	}

	wh.httpRes.ReData(
		c,
		result,
	)
}

// @Tags            web/wallet
// @Param           request         body    entities.CreateWalletForm   true    "body"
// @Param           X-Token         header  string                      true    "User token"
// @Param           Accept-Language header  string                      false   "accept language"
// @Router          /web/wallet/apply [post]
func (wh *WalletHandler) CreateWallet(c *gin.Context) {
	form := &entities.CreateWalletForm{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	userIDAny, ok := c.Get(common.CTX_KEY_AUTH_UID)
	if !ok {
		wh.logger.Error("no X-Uid")
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}
	userID, ok := userIDAny.(uint64)
	if !ok {
		wh.logger.Error("X-Uid parse failed,", userIDAny)
		wh.httpRes.ReError(c, http.StatusBadRequest, wh.beBuilder.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	walletID, err := wh.walletService.CreateWallet(c, form.CategoryID, userID)
	if err != nil {
		wh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}
	res := &entities.CreateWalletVO{
		ID: walletID,
	}
	wh.httpRes.ReData(c, res)
}
