package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	coinsdoDao "api-server/dao/coinsdo"
	notifyDao "api-server/dao/notify"
	orderDao "api-server/dao/order"
	systemDao "api-server/dao/system"
	userDao "api-server/dao/user"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
)

type CardService struct {
	userDao              *userDao.UserDao
	cardDao              *cardDao.CardDao
	mainCardDao          *cardDao.MainCardDao
	categoryDao          *accountDao.CategoryDao
	assetDao             *accountDao.AssetDao
	assetTransactionDao  *accountDao.AssetTransactionDao
	parameterDao         *systemDao.ParameterDao
	previewDao           *cardDao.PreviewDao
	transactionRecordDao *orderDao.TransactionRecordDao
	cryptoCurrencyDao    *coinsdoDao.CryptoCurrencyDao
	currencyConfigDao    *orderDao.CurrencyConfigDao
	callbackTaskDao      *notifyDao.CallbackTaskDao
	callbackLogDao       *notifyDao.CallbackLogDao
}

func NewCardService() *CardService {

	return &CardService{
		userDao:              userDao.NewUserDao(),
		cardDao:              cardDao.NewCardDao(),
		mainCardDao:          cardDao.NewMainCardDao(),
		previewDao:           cardDao.NewPreviewDao(),
		categoryDao:          accountDao.NewCategoryDao(),
		assetDao:             accountDao.NewAssetDao(),
		assetTransactionDao:  accountDao.NewAssetTransactionDao(),
		parameterDao:         systemDao.NewParameterDao(),
		transactionRecordDao: orderDao.NewTransactionRecordDao(),
		cryptoCurrencyDao:    coinsdoDao.NewCryptoCurrencyDao(),
		currencyConfigDao:    orderDao.NewCurrencyConfigDao(),
		callbackTaskDao:      notifyDao.NewCallbackTaskDao(),
		callbackLogDao:       notifyDao.NewCallbackLogDao(),
	}
}

func (cs *CardService) ListCard(ctx context.Context, form *entities.ListCardForm, currency common.Currency, userID uint64) ([]*cardDao.Card, error) {

	cards, err := cs.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID:   userID,
			IssueID:  form.IssueID,
			ID:       form.ID,
			Currency: currency,
		},
		IDIn:        form.IDIn,
		AssetTypeIn: form.AssetTypeIn,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
}

func (cs *CardService) ListCardByUserIDIn(ctx context.Context, form *entities.ListCardForm, currency common.Currency, userIDs []uint64) ([]*cardDao.Card, error) {

	cards, err := cs.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			IssueID:  form.IssueID,
			ID:       form.ID,
			Currency: currency,
		},
		IDIn:        form.IDIn,
		AssetTypeIn: form.AssetTypeIn,
		UserIDIn:    userIDs,
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
}

func (cs *CardService) PageCardProducts(ctx context.Context, c int, s int, userIDs []uint64) (cards []*cardDao.Card, current int, pageSize int, total int, err error) {

	cards, _, _, total, err = cs.cardDao.PageByUserIDInType(ctx, userIDs, common.ASSET_TYPE_CARD_PRODUCT, current, pageSize)
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, c, s, total, nil
}

func (cs *CardService) SetMainCard(ctx context.Context, form *entities.SetMainCardForm, userID uint64) error {

	card, err := cs.GetCardByID(ctx, form.CardID, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	if card == nil {
		logger.Warnf("card not found [%d]", form.CardID)
		return utils.NewBusinessError(ctx, common.CODE_CARD_NO_SUCH_CARD)
	}

	mainCards, err := cs.mainCardDao.ListByUserID(ctx, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	var mcID, currencyMCID, categoryMCID uint64
	for _, c := range mainCards {
		if c.Currency == 0 && c.CategoryID == 0 {
			mcID = c.CardID
		}
		if c.Currency == card.Currency && c.CategoryID == 0 {
			currencyMCID = c.CardID
		}
		if c.CategoryID == card.CategoryID {
			categoryMCID = c.CardID
		}
	}

	if mcID != 0 {
		rowsAffected, err := cs.mainCardDao.Update(ctx, &cardDao.MainCardQuery{
			MainCard: cardDao.MainCard{
				ID: mcID,
			},
			Attrs: cardDao.MainCard{
				CardID: form.CardID,
			},
		})
		if err != nil {
			logger.Warn("update failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if rowsAffected == 0 {
			logger.Warn("update failed [%d],", mcID)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	} else {
		_, err = cs.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID: userID,
			CardID: form.CardID,
		})
		if err != nil {
			logger.Warn("save failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}

	if currencyMCID != 0 {
		rowsAffected, err := cs.mainCardDao.Update(ctx, &cardDao.MainCardQuery{
			MainCard: cardDao.MainCard{
				ID: currencyMCID,
			},
			Attrs: cardDao.MainCard{
				CardID: form.CardID,
			},
		})
		if err != nil {
			logger.Warn("update failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if rowsAffected == 0 {
			logger.Warn("update failed [%d],", mcID)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	} else {
		_, err = cs.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:   userID,
			CardID:   form.CardID,
			Currency: card.Currency,
		})
		if err != nil {
			logger.Warn("save failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}

	if categoryMCID != 0 {
		rowsAffected, err := cs.mainCardDao.Update(ctx, &cardDao.MainCardQuery{
			MainCard: cardDao.MainCard{
				ID: categoryMCID,
			},
			Attrs: cardDao.MainCard{
				CardID: form.CardID,
			},
		})
		if err != nil {
			logger.Warn("update failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
		if rowsAffected == 0 {
			logger.Warn("update failed [%d],", mcID)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	} else {
		_, err = cs.mainCardDao.Save(ctx, &cardDao.MainCard{
			UserID:     userID,
			CardID:     form.CardID,
			Currency:   card.Currency,
			CategoryID: card.CategoryID,
		})
		if err != nil {
			logger.Warn("save failed,", err)
			return utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
		}
	}

	return nil
}

func (cs *CardService) GetMainCard(ctx context.Context, form *entities.GetMainCardForm, userID uint64) (*cardDao.MainCard, error) {

	currency := common.Currency(0).FromString(form.Currency)

	var mainCard *cardDao.MainCard
	var err error
	switch true {
	case form.CategoryID == 0 && currency == 0:
		mainCard, err = cs.mainCardDao.GetByUserID(ctx, userID)
	case form.CategoryID != 0:
		mainCard, err = cs.mainCardDao.GetByUserIDCategoryID(ctx, userID, form.CategoryID)
	case currency != 0:
		mainCard, err = cs.mainCardDao.GetByUserIDCurrency(ctx, userID, currency)
	}

	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return mainCard, nil
}

func (cs *CardService) GetCardByID(ctx context.Context, id uint64, userID uint64) (*cardDao.Card, error) {
	card, err := cs.cardDao.GetByIDUserID(ctx, id, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) GetCardByIDUserIDIn(ctx context.Context, id uint64, userIDs []uint64) (*cardDao.Card, error) {
	card, err := cs.cardDao.GetByIDUserIDIn(ctx, id, userIDs)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) GetCardByCategory(ctx context.Context, category string, categoryID uint64, userID uint64) (*cardDao.Card, error) {
	if category != "" {
		categoryID = uint64(common.Currency(0).FromString(category))
	}
	card, err := cs.cardDao.GetByUserIDCategoryID(ctx, userID, categoryID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) GetCardByUserIDInCategory(ctx context.Context, category string, categoryID uint64, userIDs []uint64) (*cardDao.Card, error) {
	if category != "" {
		categoryID = uint64(common.Currency(0).FromString(category))
	}
	card, err := cs.cardDao.GetByUserIDInCategoryID(ctx, userIDs, categoryID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) GetCardByIssueID(ctx context.Context, issueID string) (*cardDao.Card, error) {
	card, err := cs.cardDao.GetByIssueID(ctx, issueID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) GetFinancialCardByUserIDCurrency(ctx context.Context, userID uint64, currency string) (*cardDao.Card, error) {
	currencyID := common.Currency(0).FromString(currency)
	card, err := cs.cardDao.GetByUserIDCurrencyType(ctx, userID, currencyID, common.ASSET_TYPE_AUTO_YIELD)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return card, nil
}

func (cs *CardService) ListCardCategory(ctx context.Context, form *entities.ListCardCategoryForm) ([]*accountDao.Category, error) {

	categories, err := cs.categoryDao.ListByTypeUsage(ctx, form.Type, []common.CategoryUsage{common.CATEGORY_USAGE_USER_DISPLAY})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*accountDao.Category{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return categories, nil
}
