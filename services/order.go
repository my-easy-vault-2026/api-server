package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	orderDao "api-server/dao/order"
	userDao "api-server/dao/user"
	walletDao "api-server/dao/wallet"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"golang.org/x/exp/slices"
)

type OrderService struct {
	transactionRecordDao *orderDao.TransactionRecordDao
	cardDao              *cardDao.CardDao
	categoryDao          *accountDao.CategoryDao
	walletDao            *walletDao.WalletDao
	userDao              *userDao.UserDao
	logger               lib.Logger
}

func NewOrderService(
	transactionRecordDao *orderDao.TransactionRecordDao,
	cardDao *cardDao.CardDao,
	categoryDao *accountDao.CategoryDao,
	walletDao *walletDao.WalletDao,
	userDao *userDao.UserDao,
	logger lib.Logger,
) *OrderService {

	return &OrderService{
		transactionRecordDao: transactionRecordDao,
		cardDao:              cardDao,
		categoryDao:          categoryDao,
		walletDao:            walletDao,
		userDao:              userDao,
		logger:               logger,
	}
}

func (os *OrderService) PageTransactionRecords(ctx context.Context, form *entities.PageTransactionRecordsForm, userIDs []uint64) (records []*orderDao.TransactionRecord, pageCurrent int, pageSize int, total int, err error) {

	var typeIn []common.TransactionRecordType
	for _, t := range form.Types {
		typeArray := common.TransactionRecordType(0).FromString(t)
		typeIn = append(typeIn, typeArray...)
	}

	if form.CardID != 0 {
		card, err := os.cardDao.GetByIDUserIDIn(ctx, form.CardID, userIDs)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, 0, 0, 0, err
		}
		if card == nil {
			return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_ORDER_USER_HAS_NO_SUCH_CARD)
		}
		records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDInCardIDType(ctx, userIDs, form.CardID, typeIn, form.Current, form.PageSize)
		if err != nil {
			logger.Warn("get failed,", err)
			return nil, 0, 0, 0, err
		}

		return records, pageCurrent, pageSize, total, nil

	} else if form.Category != "" {
		currency := common.Currency(0).FromString(form.Category)
		if currency == 0 {
			return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_ORDER_NO_SUCH_CATEGORY)
		}
		form.CategoryID = uint64(currency)
	} else if form.CategoryID != 0 {
		// no-op
	} else {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_MISSING_PARAMETER)
	}

	records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDInCategoryIDType(ctx, userIDs, form.CategoryID, typeIn, form.Current, form.PageSize)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}

	return records, pageCurrent, pageSize, total, nil
}

func (os *OrderService) GetTransactionRecord(ctx context.Context, form *entities.GetTransactionRecordForm, userIDs []uint64) (record *orderDao.TransactionRecord, err error) {

	side := common.TransactionSide(0).FromString(form.Side)

	record, err = os.transactionRecordDao.GetByTransactionNOSideCardID(ctx, form.OrderNO, side, form.CardID, userIDs)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, err
	}
	if !slices.Contains(userIDs, record.UserID) {
		return nil, nil
	}
	return record, nil
}
