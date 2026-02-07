package services

import (
	accountDao "api-server/dao/account"
	cardDao "api-server/dao/card"
	orderDao "api-server/dao/order"
	userDao "api-server/dao/user"
	"api-server/lib"
	"context"
	"shared-modules/common"
)

type OrderService struct {
	transactionRecordDao *orderDao.TransactionRecordDao
	cardDao              *cardDao.CardDao
	categoryDao          *accountDao.CategoryDao
	userDao              *userDao.UserDao
	logger               lib.Logger
	beBuilder            *lib.BEBuilder
}

func NewOrderService(
	transactionRecordDao *orderDao.TransactionRecordDao,
	cardDao *cardDao.CardDao,
	categoryDao *accountDao.CategoryDao,
	userDao *userDao.UserDao,
	logger lib.Logger,
	beBuilder *lib.BEBuilder,
) *OrderService {

	return &OrderService{
		transactionRecordDao: transactionRecordDao,
		cardDao:              cardDao,
		categoryDao:          categoryDao,
		userDao:              userDao,
		logger:               logger,
		beBuilder:            beBuilder,
	}
}

func (os *OrderService) PageTransactionRecords(ctx context.Context, walletID uint64, userID uint64, current int, size int) (records []*orderDao.TransactionRecord, pageCurrent int, pageSize int, total int, err error) {

	wallet, err := os.cardDao.GetByIDUserID(ctx, walletID, userID)
	if err != nil {
		os.logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}
	if wallet == nil {
		return nil, 0, 0, 0, os.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_WALLET)
	}
	records, pageCurrent, pageSize, total, err = os.transactionRecordDao.PageByUserIDWalletID(ctx, userID, walletID, current, size)
	if err != nil {
		os.logger.Warn("get failed,", err)
		return nil, 0, 0, 0, err
	}

	return records, pageCurrent, pageSize, total, nil
}
