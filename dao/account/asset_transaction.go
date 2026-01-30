package account

import (
	"context"
	"shared-modules/common"
	"shared-modules/utils"

	"github.com/shopspring/decimal"
)

type AssetTransaction struct {
	UserID          uint64
	OrderNO         string
	CardID          uint64
	CategoryID      uint64
	Currency        common.Currency
	Amount          decimal.Decimal
	TransactionType common.TransactionType
	BillType        common.BillType
	Remark          string
}

type AssetTransactionDao struct {
	assetDao *AssetDao
	billDao  *BillDao
}

func NewAssetTransactionDao() *AssetTransactionDao {
	return &AssetTransactionDao{
		assetDao: NewAssetDao(),
		billDao:  NewBillDao(),
	}
}

func (td *AssetTransactionDao) BatchTransaction(ctx context.Context, transactions []*AssetTransaction, orderType common.TransactionRecordType, allowNeg bool) error {

	var err error

	for i := 0; i < len(transactions); i++ {

		switch transactions[i].TransactionType {
		case common.TRANSACTION_TYPE_ASSET_ADD:
			err = td.assetDao.AddAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount)
		case common.TRANSACTION_TYPE_ASSET_DEDUCT:
			err = td.assetDao.DeductAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_ASSET_FREEZE:
			err = td.assetDao.FreezeAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_ASSET_UNFREEZE:
			err = td.assetDao.UnfreezeAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_ADD:
			err = td.assetDao.AddFreezedAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT:
			err = td.assetDao.DeductFreezeAssets(ctx, transactions[i].UserID, transactions[i].CardID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		default:
			return utils.NewBusinessError(ctx, common.CODE_NO_SUCH_TRANSACTION_TYPE)
		}
		if err != nil {
			return err
		}

		err = td.saveBill(ctx, transactions[i], orderType)

		if err != nil {
			return err
		}
	}

	return nil
}

func (td *AssetTransactionDao) saveBill(ctx context.Context, transaction *AssetTransaction, orderType common.TransactionRecordType) error {

	var asset *Asset
	var err error
	if common.IsSystemAccount(transaction.UserID) {
		asset, err = td.assetDao.GetByUserIDCurrency(ctx, transaction.UserID, transaction.Currency)
		if err != nil {
			return err
		}
	} else {
		asset, err = td.assetDao.GetByUserIDCategoryIDCardID(ctx, transaction.UserID, transaction.CategoryID, transaction.CardID)
		if err != nil {
			return err
		}
	}

	bill := &Bill{
		UserID:               transaction.UserID,
		OrderNo:              transaction.OrderNO,
		AssetID:              asset.ID,
		Amount:               transaction.Amount,
		CurrentAmount:        asset.Amount,
		CurrentFreezedAmount: asset.FreezedAmount,
		Currency:             transaction.Currency,
		TransactionType:      transaction.TransactionType,
		BillType:             transaction.BillType,
		OrderType:            orderType,
	}

	_, err = td.billDao.Save(ctx, bill)
	if err != nil {
		return err
	}

	return nil
}
