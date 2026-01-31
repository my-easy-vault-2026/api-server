package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
)

type CardService struct {
	userDao     *userDao.UserDao
	cardDao     *cardDao.CardDao
	mainCardDao *cardDao.MainCardDao
	categoryDao *accountDao.CategoryDao
	logger      lib.Logger
}

func NewCardService(
	userDao *userDao.UserDao,
	cardDao *cardDao.CardDao,
	mainCardDao *cardDao.MainCardDao,
	categoryDao *accountDao.CategoryDao,
	logger lib.Logger,
) *CardService {
	return &CardService{
		userDao:     userDao,
		cardDao:     cardDao,
		mainCardDao: mainCardDao,
		categoryDao: categoryDao,
		logger:      logger,
	}
}
func (cs *CardService) ListWalletsByUserID(ctx context.Context, userID uint64) ([]*cardDao.Card, error) {

	cards, err := cs.cardDao.Gets(ctx, &cardDao.CardQuery{
		Card: cardDao.Card{
			UserID: userID,
		},
	})
	if err != nil {
		logger.Warn("get failed,", err)
		return []*cardDao.Card{}, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}

	return cards, nil
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

func (cs *CardService) GetCardByIDUserIDIn(ctx context.Context, id uint64, userIDs []uint64) (*cardDao.Card, error) {
	card, err := cs.cardDao.GetByIDUserIDIn(ctx, id, userIDs)
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
