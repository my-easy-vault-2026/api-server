package web

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	systemDao "api-server/dao/system"
	"api-server/services"
	"encoding/json"
	"fmt"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

type CardHandler struct {
	cardService      *services.CardService
	accountService   *services.AccountService
	walletService    *services.WalletService
	coinsdoService   *services.CoinsdoService
	userService      *services.UserService
	systemService    *services.SystemService
	notifyService    *services.NotifyService
	financialService *services.FinancialService
	orderService     *services.OrderService
}

func NewCardHandler() *CardHandler {
	return &CardHandler{
		cardService:      services.NewCardService(),
		accountService:   services.NewAccountService(),
		walletService:    services.NewWalletService(),
		coinsdoService:   services.NewCoinsdoService(),
		userService:      services.NewUserService(),
		systemService:    services.NewSystemService(),
		notifyService:    services.NewNotifyService(),
		financialService: services.NewFinancialService(),
		orderService:     services.NewOrderService(),
	}
}

// @Param			request			body		entities.GetCardForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Param			X-App-Version		header		string					false	"app version"
// @Success		0				{object}	entities.CardVO		"data"
// @Router			/web/card/get [post]
// @Description	List user cards.
// @Tags			web/card
func (ch *CardHandler) GetCard(c *gin.Context) {

	form := &entities.GetCardForm{}

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

	user, err := ch.userService.GetUserRole(c, &entities.GetUserForm{
		ID: userID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	// 取得 merchant id
	merchants, err := ch.userService.ListByEmailRole(c, user.Email, common.ROLE_MERCHANT_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	userIDs := []uint64{userID}
	for _, m := range merchants {
		userIDs = append(userIDs, m.ID)
	}

	var card *cardDao.Card
	if form.ID != 0 {
		card, err = ch.cardService.GetCardByIDUserIDIn(c, form.ID, userIDs)
		if err != nil {
			utils.ReError(c, err)
			return
		}
	} else if form.CategoryID != 0 {
		card, err = ch.cardService.GetCardByUserIDInCategory(c, form.Category, form.CategoryID, userIDs)
		if err != nil {
			utils.ReError(c, err)
			return
		}
	}
	if card == nil {
		utils.ReData(
			c,
			nil,
		)
		return
	}

	var asset *accountDao.Asset
	asset, err = ch.accountService.GetAsset(c, card.ID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if asset == nil {
		logger.Error("asset wallet mismatch")
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_CARD_WALLET_ASSET_MISMATCH))
		return
	}

	category, err := ch.accountService.GetCategory(c, &entities.GetCategoryForm{
		ID: card.CategoryID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if category == nil {
		logger.Error("category not found: [%d]", card.CategoryID)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
		return
	}

	cardSwitch := make(map[string]interface{}, 32)
	for i := common.CategoryFrontendUsage(1); i < common.CATEGORY_FRONT_USAGE_NONE; i *= 2 {
		if i.String() == "" {
			break
		}
		if category.FrontendUsage&i != 0 {
			cardSwitch[i.String()] = true
			continue
		}
		cardSwitch[i.String()] = false
	}

	mainnetNames, err := ch.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainnetNames[0] = ""

	displayStatus := ch.getDisplayStatus(card)

	var result *entities.CardVO
	switch card.Type {
	case common.ASSET_TYPE_CARD_PRODUCT:
		panNumber := ""
		if len(card.PANNumber) > 4 {
			panNumber = "************" + card.PANNumber[len(card.PANNumber)-4:]
		}

		amount := asset.Amount.Copy()
		if card.Vendor == common.CARD_PRODUCT_VENDOR_WHALE || card.Vendor == common.CARD_PRODUCT_VENDOR_PAYCRYPTO {
			amount = card.Amount.Copy()
		}

		result = &entities.CardVO{
			ID:            card.ID,
			UserID:        card.UserID,
			MerchantID:    card.MerchantID,
			CategoryID:    card.CategoryID,
			PreferredName: card.PreferredName,
			Type:          common.ASSET_TYPE_CARD_PRODUCT.String(),
			Amount:        amount,
			PANNumber:     panNumber,
			Issuer:        card.Issuer,
			Currency:      card.Currency.String(),
			Format:        card.Format.String(),
			Alias:         card.Alias,
			CardSwitch:    cardSwitch,
			Status:        card.Status.String(),
			FreezeStatus:  card.FreezeStatus.String(),
			Design:        category.CustomDesign,
			DisplayStatus: displayStatus,
			CreatedAt:     card.CreatedAt.UnixMilli(),
			UpdatedAt:     card.UpdatedAt.UnixMilli(),
		}

		if result.Design == "" {
			result.Design = "default"
		}

	case common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_POINT:
		mps := make([]*entities.SupportedVO, 0, 10)
		mpArr := strings.Split(category.Supported, ",")
		for _, mpStr := range mpArr {
			if mpStr == "" {
				break
			}
			strs := strings.Split(mpStr, "_")
			protocol := 0
			if len(strs) > 1 {
				protocol, _ = strconv.Atoi(strs[1])
			}
			mainnet, _ := strconv.Atoi(strs[0])
			mp := &entities.SupportedVO{
				Mainnet:     common.Mainnet(mainnet).String(),
				MainnetName: mainnetNames[common.Mainnet(mainnet)],
				Protocol:    common.Protocol(protocol).String(),
			}
			if common.Mainnet(mainnet) == category.RecommendedMainnet &&
				common.Protocol(protocol) == category.RecommendedProtocol {
				mp.Recommended = true
			}
			if category.Currency == common.CURRENCY_ETH {
				mp.Protocol = "erc-20"
			}
			mps = append(mps, mp)
		}

		var pointLevel string
		if card.CategoryID == uint64(common.CURRENCY_EPOINT) {
			pointLevel = user.EPointLevel.String()
		}

		result = &entities.CardVO{
			ID:            card.ID,
			UserID:        card.UserID,
			MerchantID:    card.MerchantID,
			CategoryID:    card.CategoryID,
			PreferredName: category.PreferredName,
			Amount:        asset.Amount.Copy(),
			Issuer:        category.Issuer,
			Type:          card.Currency.Type().String(),
			Currency:      card.Currency.String(),
			PointLevel:    pointLevel,
			Alias:         card.Alias,
			Supported:     mps,
			CardSwitch:    cardSwitch,
			Status:        card.Status.String(),
			FreezeStatus:  card.FreezeStatus.String(),
			Design:        category.CustomDesign,
			DisplayStatus: displayStatus,
			CreatedAt:     card.CreatedAt.UnixMilli(),
			UpdatedAt:     card.UpdatedAt.UnixMilli(),
		}
	}

	utils.ReData(
		c,
		result,
	)
}

// @Param			request			body		entities.ListCardForm	true	"body"
// @Param			X-Token			header		string					true	"User token"
// @Param			Accept-Language	header		string					false	"accept language"
// @Param			X-Extend		header		string					false	"Extend"
// @Param			X-Convert		header		string					false	"Convert"
// @Param			X-App-Version		header		string					false	"app version"
// @Success		0				{object}	entities.ListCardVO		"data"
// @Router			/web/card/list [post]
// @Description	List user cards.
// @Tags			web/card
func (ch *CardHandler) ListCard(c *gin.Context) {

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

	user, err := ch.userService.GetUserRole(c, &entities.GetUserForm{
		ID: userID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	currency := common.Currency(0).FromString(form.Currency)

	// 舊版只能拿數幣和e卡
	// 新版可以拿積分卡
	form.AssetTypeIn = []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_CARD_PRODUCT, common.ASSET_TYPE_POINT}

	// 取得 merchant id
	merchants, err := ch.userService.ListByEmailRole(c, user.Email, common.ROLE_MERCHANT_USER)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	userIDs := []uint64{userID}
	for _, m := range merchants {
		userIDs = append(userIDs, m.ID)
	}

	cards, err := ch.cardService.ListCardByUserIDIn(c, form, currency, userIDs)
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

	assets, err := ch.accountService.ListAssetsByUserIDIn(c, &entities.ListAssetsForm{
		IDIn:   cardIDs,
		TypeIn: form.AssetTypeIn,
	}, userIDs)
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

	mainCard, err := ch.cardService.GetMainCard(c, &entities.GetMainCardForm{}, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainCardID := uint64(0)
	if mainCard != nil {
		mainCardID = mainCard.CardID
	}

	categories, err := ch.accountService.ListCategories(c, &entities.ListCategoriesForm{
		IDIn: categoryIDs,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}
	categoryMap := map[uint64]*accountDao.Category{}
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	mainnetNames, err := ch.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainnetNames[0] = ""

	product, err := ch.financialService.GetProduct(c, &entities.FinancialGetProductForm{
		Code: common.FINANCIAL_CODE_AUTO_YIELD,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}
	autoYieldSupSet := make(map[common.Currency]bool)
	allSupported := false
	if product != nil && product.Status == common.FINANCIAL_PRODUCT_STATUS_ACTIVE {

		switch product.SupportType {
		case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_ALL:
			allSupported = true
		case common.FINANCIAL_PRODUCT_SUPPORT_TYPE_CURRENCY:
			sups := strings.Split(product.SupportedCurrencies, ",")
			for _, s := range sups {
				autoYieldSupSet[common.Currency(0).FromString(s)] = true
			}
		}
	}

	param, err := ch.systemService.GetSystemParameterByKey(c, common.PARAMETER_KEY_AUTO_YIELD_CRYPTO_BUTTON_TOGGLE)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	ayButton := true
	if param != nil {
		ayButton, err = strconv.ParseBool(param.Value)
		if err != nil {
			logger.Warnf("parse auto yield crypto button toggle failed, key: %s, value: %s, err: %v", common.PARAMETER_KEY_AUTO_YIELD_CRYPTO_BUTTON_TOGGLE, param.Value, err)
			ayButton = true
		}
	}

	result := &entities.ListCardVO{
		Records:    make([]*entities.CardVO, len(cards)),
		MainCardID: mainCardID,
	}

	for i, card := range cards {
		category, ok := categoryMap[card.CategoryID]
		if !ok {
			logger.Warnf("category not found, id: %d", card)
			continue
		}

		cardSwitch := make(map[string]interface{}, 32)
		for i := common.CategoryFrontendUsage(1); i < common.CATEGORY_FRONT_USAGE_NONE; i *= 2 {
			if i.String() == "" {
				break
			}
			if category.FrontendUsage&i != 0 {
				cardSwitch[i.String()] = true
				continue
			}
			cardSwitch[i.String()] = false
		}

		displayStatus := ch.getDisplayStatus(card)

		switch card.Type {
		case common.ASSET_TYPE_CARD_PRODUCT:
			panNumber := ""
			if len(card.PANNumber) > 4 {
				panNumber = "************" + card.PANNumber[len(card.PANNumber)-4:]
			}

			amount := assetMap[card.ID].Copy()
			if card.Vendor == common.CARD_PRODUCT_VENDOR_WHALE || card.Vendor == common.CARD_PRODUCT_VENDOR_PAYCRYPTO {
				amount = card.Amount.Copy()
			}
			result.Records[i] = &entities.CardVO{
				ID:             card.ID,
				UserID:         card.UserID,
				MerchantID:     card.MerchantID,
				CategoryID:     card.CategoryID,
				PreferredName:  card.PreferredName,
				Type:           common.ASSET_TYPE_CARD_PRODUCT.String(),
				Organization:   card.Organization.String(),
				Vendor:         card.Vendor.String(),
				Amount:         amount,
				PANNumber:      panNumber,
				Issuer:         card.Issuer,
				Currency:       card.Currency.String(),
				Format:         card.Format.String(),
				CardSwitch:     cardSwitch,
				Alias:          card.Alias,
				Status:         card.Status.String(),
				DeliveryStatus: card.DeliveryStatus.String(),
				FreezeStatus:   card.FreezeStatus.String(),
				Design:         category.CustomDesign,
				DisplayStatus:  displayStatus,
				CreatedAt:      card.CreatedAt.UnixMilli(),
				UpdatedAt:      card.UpdatedAt.UnixMilli(),
			}

			if result.Records[i].Design == "" {
				result.Records[i].Design = "default"
			}

			if card.CustomDesign != "" {
				result.Records[i].Design = card.CustomDesign
			}

			if ayButton {
				if allSupported || autoYieldSupSet[card.Currency] {
					result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_ACTIVE.String()
				} else {
					result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_INACTIVE.String()
				}
			} else {
				result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_INACTIVE.String()
			}

			if card.EarningStatus == common.CARD_EARNING_STATUS_ENABLED {
				result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_ACTIVE.String()
			}

			if card.Vendor != common.CARD_PRODUCT_VENDOR_REAP {
				result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_INACTIVE.String()
			}

		case common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_POINT:
			mps := make([]*entities.SupportedVO, 0, 10)
			mpArr := strings.Split(category.Supported, ",")
			for _, mpStr := range mpArr {
				if mpStr == "" {
					break
				}
				strs := strings.Split(mpStr, "_")
				protocol := 0
				if len(strs) > 1 {
					protocol, _ = strconv.Atoi(strs[1])
				}
				mainnet, _ := strconv.Atoi(strs[0])
				mp := &entities.SupportedVO{
					Mainnet:     common.Mainnet(mainnet).String(),
					MainnetName: mainnetNames[common.Mainnet(mainnet)],
					Protocol:    common.Protocol(protocol).String(),
				}
				if common.Mainnet(mainnet) == category.RecommendedMainnet &&
					common.Protocol(protocol) == category.RecommendedProtocol {
					mp.Recommended = true
				}
				if category.Currency == common.CURRENCY_ETH {
					mp.Protocol = "erc-20"
				}
				mps = append(mps, mp)
			}

			var pointLevel, nextPointLevel string
			var pointToNextLevel decimal.Decimal
			if card.CategoryID == uint64(common.CURRENCY_EPOINT) {
				pointLevel = user.EPointLevel.String()
				if (user.EPointLevel + 1).String() != "" {
					param, err := ch.systemService.GetSystemParameterByKey(c, common.PARAMETER_KEY_EPOINT_LEVEL_REQUIREMENT)
					if err != nil {
						logger.Warn("get epoint level requirement failed,", err)
						param = &systemDao.Parameter{
							Value: "0,100000,500000",
						}
					}
					if param == nil {
						logger.Error("get epoint level requirement failed, default value used")
						param = &systemDao.Parameter{
							Value: "0,100000,500000",
						}
					}
					levelRequirements := strings.Split(param.Value, ",")
					nextPointLevel = (user.EPointLevel + 1).String()
					pointToNextLevel, err = decimal.NewFromString(levelRequirements[user.EPointLevel])
					if err != nil {
						logger.Error("parse epoint level requirement failed,", err)
						switch user.EPointLevel + 1 {
						case common.EPOINT_LEVEL_SILVER:
							pointToNextLevel = decimal.NewFromFloat(100000)
						case common.EPOINT_LEVEL_GOLD:
							pointToNextLevel = decimal.NewFromFloat(500000)
						}
					}
				}
			}

			result.Records[i] = &entities.CardVO{
				ID:               card.ID,
				UserID:           card.UserID,
				MerchantID:       card.MerchantID,
				CategoryID:       card.CategoryID,
				PreferredName:    category.PreferredName,
				Amount:           assetMap[card.ID].Copy(),
				Issuer:           category.Issuer,
				Type:             card.Currency.Type().String(),
				Currency:         card.Currency.String(),
				PointLevel:       pointLevel,
				NextPointLevel:   nextPointLevel,
				PointToNextLevel: &pointToNextLevel,
				Alias:            card.Alias,
				Supported:        mps,
				CardSwitch:       cardSwitch,
				Status:           card.Status.String(),
				FreezeStatus:     card.FreezeStatus.String(),
				Design:           category.CustomDesign,
				DisplayStatus:    displayStatus,
				CreatedAt:        card.CreatedAt.UnixMilli(),
				UpdatedAt:        card.UpdatedAt.UnixMilli(),
			}

			if ayButton {
				if allSupported {
					result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_ACTIVE.String()
				} else if autoYieldSupSet[card.Currency] && (card.Type == common.ASSET_TYPE_CRYPTO || card.Type == common.ASSET_TYPE_FIAT) {
					result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_ACTIVE.String()
				} else if card.Type == common.ASSET_TYPE_CRYPTO || card.Type == common.ASSET_TYPE_FIAT {
					result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_INACTIVE.String()
				}
			} else {
				result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_INACTIVE.String()
			}

			if card.EarningStatus == common.CARD_EARNING_STATUS_ENABLED {
				result.Records[i].AutoYieldStatus = common.FINANCIAL_PRODUCT_STATUS_ACTIVE.String()
			}
		}
	}

	utils.ReData(
		c,
		result,
	)
}

// 舊版的app不能拿新卡片
func (ch *CardHandler) ListCryptpAndProductCard(c *gin.Context) {

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
	}

	userID, err := strconv.ParseUint(userIDString, 10, 64)
	if err != nil {
		logger.Error("X-Uid parse failed,", err)
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_SYSTEM_ERROR))
	}

	currency := common.Currency(0).FromString(form.Currency)

	// 舊版只能拿數幣和e卡
	// 新版可以拿積分卡
	form.AssetTypeIn = []common.AssetType{common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_CARD_PRODUCT}

	cards, err := ch.cardService.ListCard(c, form, currency, userID)
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

	assets, err := ch.accountService.ListAssets(c, &entities.ListAssetsForm{
		UserID: userID,
		IDIn:   cardIDs,
		TypeIn: form.AssetTypeIn,
	}, userID)
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

	mainCard, err := ch.cardService.GetMainCard(c, &entities.GetMainCardForm{}, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainCardID := uint64(0)
	if mainCard != nil {
		mainCardID = mainCard.CardID
	}

	categories, err := ch.accountService.ListCategories(c, &entities.ListCategoriesForm{
		IDIn: categoryIDs,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}
	categoryMap := map[uint64]*accountDao.Category{}
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	mainnetNames, err := ch.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainnetNames[0] = ""

	result := &entities.ListCardVO{
		Records:    make([]*entities.CardVO, len(cards)),
		MainCardID: mainCardID,
	}

	for i, card := range cards {
		category, ok := categoryMap[card.CategoryID]
		if !ok {
			logger.Warnf("category not found, id: %d", card)
			continue
		}

		cardSwitch := make(map[string]interface{}, 32)
		for i := common.CategoryFrontendUsage(1); i < common.CATEGORY_FRONT_USAGE_NONE; i *= 2 {
			if i.String() == "" {
				break
			}
			if category.FrontendUsage&i != 0 {
				cardSwitch[i.String()] = true
				continue
			}
			cardSwitch[i.String()] = false
		}

		displayStatus := ch.getDisplayStatus(card)

		switch card.Type {
		case common.ASSET_TYPE_CARD_PRODUCT:
			panNumber := ""
			if len(card.PANNumber) > 4 {
				panNumber = "************" + card.PANNumber[len(card.PANNumber)-4:]
			}

			amount := assetMap[card.ID].Copy()
			if card.Vendor == common.CARD_PRODUCT_VENDOR_WHALE || card.Vendor == common.CARD_PRODUCT_VENDOR_PAYCRYPTO {
				amount = card.Amount.Copy()
			}

			result.Records[i] = &entities.CardVO{
				ID:            card.ID,
				CategoryID:    card.CategoryID,
				Vendor:        card.Vendor.String(),
				PreferredName: card.PreferredName,
				Type:          common.ASSET_TYPE_CARD_PRODUCT.String(),
				Amount:        amount,
				PANNumber:     panNumber,
				Issuer:        card.Issuer,
				Currency:      card.Currency.String(),
				Format:        card.Format.String(),
				CardSwitch:    cardSwitch,
				Alias:         card.Alias,
				Status:        card.Status.String(),
				FreezeStatus:  card.FreezeStatus.String(),
				Design:        category.CustomDesign,
				DisplayStatus: displayStatus,
				CreatedAt:     card.CreatedAt.UnixMilli(),
				UpdatedAt:     card.UpdatedAt.UnixMilli(),
			}

			if result.Records[i].Design == "" {
				result.Records[i].Design = "default"
			}

		case common.ASSET_TYPE_CRYPTO, common.ASSET_TYPE_FIAT, common.ASSET_TYPE_POINT:
			mps := make([]*entities.SupportedVO, 0, 10)
			mpArr := strings.Split(category.Supported, ",")
			for _, mpStr := range mpArr {
				if mpStr == "" {
					break
				}
				strs := strings.Split(mpStr, "_")
				protocol := 0
				if len(strs) > 1 {
					protocol, _ = strconv.Atoi(strs[1])
				}
				mainnet, _ := strconv.Atoi(strs[0])
				mp := &entities.SupportedVO{
					Mainnet:     common.Mainnet(mainnet).String(),
					MainnetName: mainnetNames[common.Mainnet(mainnet)],
					Protocol:    common.Protocol(protocol).String(),
				}
				if common.Mainnet(mainnet) == category.RecommendedMainnet &&
					common.Protocol(protocol) == category.RecommendedProtocol {
					mp.Recommended = true
				}
				if category.Currency == common.CURRENCY_ETH {
					mp.Protocol = "erc-20"
				}
				mps = append(mps, mp)
			}

			result.Records[i] = &entities.CardVO{
				ID:            card.ID,
				CategoryID:    card.CategoryID,
				PreferredName: category.PreferredName,
				Amount:        assetMap[card.ID].Copy(),
				Issuer:        category.Issuer,
				Type:          card.Currency.Type().String(),
				Currency:      card.Currency.String(),
				Alias:         card.Alias,
				Supported:     mps,
				CardSwitch:    cardSwitch,
				Status:        card.Status.String(),
				FreezeStatus:  card.FreezeStatus.String(),
				Design:        category.CustomDesign,
				DisplayStatus: displayStatus,
				CreatedAt:     card.CreatedAt.UnixMilli(),
				UpdatedAt:     card.UpdatedAt.UnixMilli(),
			}
		}

	}

	utils.ReData(
		c,
		result,
	)
}

// @Param			request			body	entities.SetMainCardForm	true	"body"
// @Param			X-Token			header	string						true	"User token"
// @Param			Accept-Language	header	string						false	"accept language"
// @Param			X-Extend		header	string						false	"Extend"
// @Param			X-Convert		header	string						false	"Convert"
// @Router			/web/card/setMainCard [post]
// @Description	設定主卡
// @Tags			web/card
func (ch *CardHandler) SetMainCard(c *gin.Context) {
	form := &entities.SetMainCardForm{}

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

	err = ch.cardService.SetMainCard(c, form, userID)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(
		c,
		nil,
	)
}

// @Param			request			body		entities.ListCardCategoryForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Param			X-App-Version		header		string						false	"App-Version"
// @Success		0				{object}	entities.ListCardCategoryVO		"data"
// @Router			/web/card/listCategory [post]
// @Description	List card categories. <br> type: 1:虛擬貨幣, 2:法幣, 3:卡片
// @Tags			web/card
func (ch *CardHandler) ListCardCategory(c *gin.Context) {
	form := &entities.ListCardCategoryForm{}

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

	user, err := ch.userService.GetUserRole(c, &entities.GetUserForm{
		ID: userID,
	})
	if err != nil {
		utils.ReError(c, err)
		return
	}

	categories, err := ch.cardService.ListCardCategory(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	if categories == nil {
		logger.Warn("categories is nil")
		return
	}

	keys := []common.ParameterKey{
		common.PARAMETER_KEY_MONETA_PROMOTION_CODE,
		common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE,
		common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_EXCHANGE_FEE,
		common.PARAMETER_KEY_REAP_DECLINE_FINE,
		common.PARAMETER_KEY_REAP_ATM_RESTRICTED_COUNTRIES,
	}
	params, err := ch.systemService.GetParams(c, keys)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	var monetaLists []string
	if params[common.PARAMETER_KEY_MONETA_PROMOTION_CODE] != "" {
		// 全部轉小寫
		rawList := strings.Split(params[common.PARAMETER_KEY_MONETA_PROMOTION_CODE], ",")
		for _, v := range rawList {
			monetaLists = append(monetaLists, strings.ToLower(strings.TrimSpace(v)))
		}
	}

	// user.PromotionCode 轉小寫
	promotionCodeLower := strings.ToLower(strings.TrimSpace(user.PromotionCode))

	if slices.Contains(monetaLists, promotionCodeLower) {
		categories = slices.DeleteFunc(categories, func(c *accountDao.Category) bool {
			return c.Type == common.ASSET_TYPE_CARD_PRODUCT &&
				c.CardKind != common.CARD_KIND_REAP_MONETA_VIRTUAL &&
				c.CardKind != common.CARD_KIND_REAP_MONETA_PHYSICAL
		})
	}

	decimals, err := ch.coinsdoService.ListDisplayDecimals(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	decimals[0] = 10

	mainnetNames, err := ch.coinsdoService.ListMainnetNames(c)
	if err != nil {
		utils.ReError(c, err)
		return
	}
	mainnetNames[0] = ""

	sceneFormatMap := make(map[common.ContentScene][]string)

	for _, category := range categories {
		scene, format := ch.getSceneFormatByCategory(category)
		if format != "" {
			sceneFormatMap[scene] = append(sceneFormatMap[scene], format)
		}
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
	lang := languages[0]

	sceneContentMap := make(map[common.ContentScene]map[string]map[string]interface{})
	for scene, customIDs := range sceneFormatMap {
		contentList, err := ch.systemService.ListStructuredContentBySceneCustomIDsLanguage(c, scene, customIDs, lang)
		if err != nil {
			utils.ReError(c, err)
			return
		}
		if len(contentList) == 0 {
			contentList, err = ch.systemService.ListStructuredContentBySceneCustomIDsLanguage(c, scene, customIDs, "en")
			if err != nil {
				utils.ReError(c, err)
				return
			}
		}

		parsedMap := make(map[string]map[string]interface{})
		for _, content := range contentList {
			m := make(map[string]interface{})
			err = json.Unmarshal([]byte(content.Content), &m)
			if err != nil {
				logger.Warnf("fail to unmarshal: [%#v]", content)
				logger.Warnf("content: [%s]", content.Content)
				utils.ReError(c, err)
				return
			}
			parsedMap[content.CustomID] = m
		}
		sceneContentMap[scene] = parsedMap
	}

	result := &entities.ListCardCategoryVO{
		Records: make([]*entities.CardCategoryVO, 0, len(categories)),
	}

	limitCountryUSParam, err := ch.systemService.GetSystemParameterByKey(c, common.PARAMETER_KEY_WHALE_CARD_LIMIT_COUNTRY_US)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	var carBinList []string
	if limitCountryUSParam != nil {
		carBinList = strings.Split(limitCountryUSParam.Value, ",")
	}

	for _, category := range categories {

		categoryCopy := &entities.CardCategoryVO{}
		if err := copier.Copy(&categoryCopy, &category); err != nil {
			utils.ReError(c, err)
			return
		}
		contentMap := make(map[string]map[string]interface{})
		scene, customID := ch.getSceneFormatByCategory(category)
		contentMap = sceneContentMap[scene]

		if content, ok := contentMap[customID]; ok {
			if t, ok := content["title"]; ok {
				categoryCopy.Title = t.(string)
			}
			if fArr, ok := content["feature"]; ok {
				arr := fArr.([]interface{})
				features := make([]entities.FeatureVO, 0, len(arr))
				for _, item := range arr {
					if fItem, ok := item.(map[string]interface{}); ok {
						var feature entities.FeatureVO
						if icon, ok := fItem["icon"].(string); ok {
							feature.Icon = icon
						}
						if text, ok := fItem["text"].(string); ok {
							feature.Text = text
						}
						features = append(features, feature)
					}
				}
				categoryCopy.Feature = features
			}
			if fArr, ok := content["fees"]; ok {
				arr := fArr.([]interface{})
				fees := make([]entities.FeeVO, 0, len(arr))

				topUpExchangeFeeStr := params[common.PARAMETER_KEY_TOP_UP_TOP_UP_EXCHANGE_FEE]

				// 轉成 float
				topUpExchangeFee, err := strconv.ParseFloat(topUpExchangeFeeStr, 64)
				if err != nil {
					// 錯誤處理，例如回傳錯誤
					utils.ReError(c, fmt.Errorf("invalid top_up_exchange_fee: %v", err))
					return
				}

				// 計算
				topUpExchangeFee = topUpExchangeFee * 100

				topDownExchangeFeeStr := params[common.PARAMETER_KEY_TOP_DOWN_TOP_DOWN_EXCHANGE_FEE]

				// 轉成 float
				topDownExchangeFee, err := strconv.ParseFloat(topDownExchangeFeeStr, 64)
				if err != nil {
					// 錯誤處理，例如回傳錯誤
					utils.ReError(c, fmt.Errorf("invalid top_up_exchange_fee: %v", err))
					return
				}

				// 計算
				topDownExchangeFee = topDownExchangeFee * 100

				var countryNames []string
				countryCodes := strings.Split(params[common.PARAMETER_KEY_REAP_ATM_RESTRICTED_COUNTRIES], ",")
				for _, code := range countryCodes {
					if !common.NationCode(code).IsValid() {
						continue
					}
					countryNames = append(countryNames, utils.Translate(c, common.TranslateMsg("text_system_country_"+strings.ToLower(string(common.NationCode(code))))))
				}

				textParams := map[common.TemplateVariable]string{
					common.TEMPLATE_VARIABLE_RENEWAL_FEE:                   category.RenewalFee.String(),
					common.TEMPLATE_VARIABLE_REAP_DECLINE_FINE:             params[common.PARAMETER_KEY_REAP_DECLINE_FINE],
					common.TEMPLATE_VARIABLE_TOP_UP_EXCHANGE_FEE:           fmt.Sprintf("%.0f", topUpExchangeFee),
					common.TEMPLATE_VARIABLE_TOP_DOWN_EXCHANGE_FEE:         fmt.Sprintf("%.0f", topDownExchangeFee),
					common.TEMPLATE_VARIABLE_REAP_ATM_RESTRICTED_COUNTRIES: strings.Join(countryNames, ", "),
				}

				for _, item := range arr {
					if fItem, ok := item.(map[string]interface{}); ok {
						var fee entities.FeeVO
						if icon, ok := fItem["key"].(string); ok {
							fee.Key = icon
						}
						if text, ok := fItem["value"].(string); ok {
							for k, v := range textParams {
								text = strings.ReplaceAll(text, "{$"+string(k)+"}", v)
							}
							fee.Value = text
						}
						fees = append(fees, fee)
					}
				}
				categoryCopy.Fees = fees
			}
			if p, ok := content["purchasableItems"]; ok {
				is := p.([]interface{})
				p2 := make([]string, 0, len(is))
				for _, itf := range is {
					p2 = append(p2, itf.(string))
				}
				categoryCopy.PurchasableItems = p2
			}
			if d, ok := content["description"]; ok {
				categoryCopy.Description = d.(string)
			}
			if tagline, ok := content["tagline"]; ok {
				categoryCopy.Tagline = tagline.(string)
			}
		}

		if category.Supported != "" {
			mps := make([]*entities.SupportedVO, 0, 10)
			mpArr := strings.Split(category.Supported, ",")
			for _, mpStr := range mpArr {
				strs := strings.Split(mpStr, "_")
				protocol := 0
				if len(strs) > 1 {
					protocol, _ = strconv.Atoi(strs[1])
				}
				mainnet, _ := strconv.Atoi(strs[0])
				mp := &entities.SupportedVO{
					Mainnet:     common.Mainnet(mainnet).String(),
					MainnetName: mainnetNames[common.Mainnet(mainnet)],
					Protocol:    common.Protocol(protocol).String(),
				}
				if common.Mainnet(mainnet) == category.RecommendedMainnet &&
					common.Protocol(protocol) == category.RecommendedProtocol {
					mp.Recommended = true
				}
				mps = append(mps, mp)
			}
			categoryCopy.Supported = mps
		}

		if category.ActivationDeposit != nil {
			categoryCopy.ActivationDeposit = category.ActivationDeposit.StringFixed(int32(decimals[category.Currency]))
		}

		if category.AnnualFee != nil {
			categoryCopy.AnnualFee = category.AnnualFee.StringFixed(int32(decimals[category.Currency]))
		}

		if slices.Contains(carBinList, category.WhaleCardBin) {
			categoryCopy.LimitCountryUS = true
		} else {
			categoryCopy.LimitCountryUS = false
		}

		categorySwitch := make(map[string]interface{}, 32)
		var wallets []string
		for i := common.CategoryFrontendUsage(1); i < common.CATEGORY_FRONT_USAGE_NONE; i *= 2 {
			if i.String() == "" {
				break
			}
			if category.FrontendUsage&i != 0 {
				categorySwitch[i.String()] = true
				continue
			}
			categorySwitch[i.String()] = false
		}

		categoryCopy.Wallets = wallets
		categoryCopy.CategorySwitch = categorySwitch
		categoryCopy.Type = category.Type.String()
		categoryCopy.Currency = category.Currency.String()
		categoryCopy.CurrencyType = category.CurrencyType.String()
		categoryCopy.FeeCurrency = category.FeeCurrency.String()
		categoryCopy.Vendor = category.Vendor.String()
		categoryCopy.Organization = category.Organization.String()
		categoryCopy.CreatedAt = category.CreatedAt.UnixMilli()
		categoryCopy.UpdatedAt = category.UpdatedAt.UnixMilli()

		result.Records = append(result.Records, categoryCopy)
	}

	utils.ReData(
		c,
		result,
	)
}

func (ch *CardHandler) getSceneFormatByCategory(category *accountDao.Category) (common.ContentScene, string) {
	var scene common.ContentScene
	var format string

	switch category.Vendor {
	case common.CARD_PRODUCT_VENDOR_REAP:
		scene = common.CONTENT_SCENE_CARD_FORMAT
	case common.CARD_PRODUCT_VENDOR_PAYCRYPTO:
		scene = common.CONTENT_SCENE_PAYCRYPTO_CARD
	case common.CARD_PRODUCT_VENDOR_WHALE:
		scene = common.CONTENT_SCENE_WHALE_CARD_TYPE_BIN
	case common.CARD_PRODUCT_VENDOR_ETHERFI:
		scene = common.CONTENT_SCENE_ETHERFI_CARD
	default:
		return scene, format
	}

	format = fmt.Sprintf("%d", category.CardKind)

	return scene, format
}

func (ch *CardHandler) getDisplayStatus(card *cardDao.Card) common.DisplayStatus {

	var displayStatus common.DisplayStatus

	switch card.Status {
	case common.CARD_STATUS_NOT_CREATED:
		displayStatus = common.DISPLAY_STATUS_NOT_CREATED
	case common.CARD_STATUS_ACTIVATED:
		displayStatus = common.DISPLAY_STATUS_ACTIVATED
	case common.CARD_STATUS_BLOCKED:
		displayStatus = common.DISPLAY_STATUS_BLOCKED
		if card.BlockReason != nil && *card.BlockReason == common.CARD_STATUS_FUND_EXCEED_MAX_CONSECUTIVE_FAILURES {
			displayStatus = common.DISPLAY_STATUS_BLOCKED_DUE_TO_DECLINES
		}
	case common.CARD_STATUS_CREATED:
		displayStatus = common.DISPLAY_STATUS_CREATED
	case common.CARD_STATUS_NOT_ACTIVATED:
		switch card.Vendor {
		case common.CARD_PRODUCT_VENDOR_PAYCRYPTO:
			displayStatus = common.DISPLAY_STATUS_ACTIVATED
			if card.Format == common.CARD_FORMAT_PHYSICAL {
				displayStatus = common.DISPLAY_STATUS_NOT_ACTIVATED
			}
		default:
			displayStatus = common.DISPLAY_STATUS_NOT_ACTIVATED
		}
	}

	if card.Status != common.CARD_STATUS_BLOCKED &&
		card.FreezeStatus == common.CARD_FREEZE_STATUS_FROZEN &&
		card.FreezeReason != nil {
		displayStatus = common.DISPLAY_STATUS_FROZEN
	}

	return displayStatus
}
