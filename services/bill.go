package services

import (
	accountDao "api-server/dao/account"
	"context"
	"shared-modules/common"
	"shared-modules/logger"
	"shared-modules/utils"
	"time"
)

type BillService struct {
	billDao *accountDao.BillDao
}

func NewBillService() *BillService {
	return &BillService{
		billDao: accountDao.NewBillDao(),
	}
}

func (bs *BillService) GetByDateRange(ctx context.Context, currency common.Currency, startDate, endDate time.Time) ([]*accountDao.Bill, error) {
	bills, err := bs.billDao.GetByDateRange(ctx, currency, startDate, endDate)
	if err != nil {
		logger.Warn("get failed,", err)
		return nil, utils.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR)
	}
	return bills, nil
}
