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
	"shared-modules/logger"
	"shared-modules/utils"
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

func (os *OrderService) PageTransactionRecords(ctx context.Context, walletID uint64, userID uint64, current int, size int) (records []*orderDao.TransactionRecord, pageCurrent int, pageSize int, total int, err error) {

	card, err := os.cardDao.GetByIDUserID(ctx, walletID, userID)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}
	if card == nil {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_ORDER_USER_HAS_NO_SUCH_CARD)
	}
	records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDCardID(ctx, userID, walletID, current, size)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}

	return records, pageCurrent, pageSize, total, nil
}
