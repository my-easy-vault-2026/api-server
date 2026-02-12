package account

import (
	"context"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AssetTransaction struct {
	UserID          uint64
	OrderNO         string
	WalletID        uint64
	CategoryID      uint64
	Currency        common.Currency
	Amount          decimal.Decimal
	TransactionType common.TransactionType
	BillType        common.BillType
	Remark          string
}

type AssetTransactionDao struct {
	db        infra.Database
	env       *lib.Env
	assetDao  *AssetDao
	billDao   *BillDao
	beBuilder *lib.BEBuilder
}

func NewAssetTransactionDao(db infra.Database,
	env *lib.Env,
	assetDao *AssetDao,
	billDao *BillDao,
	beBuilder *lib.BEBuilder) *AssetTransactionDao {
	return &AssetTransactionDao{
		db:        db,
		env:       env,
		assetDao:  assetDao,
		billDao:   billDao,
		beBuilder: beBuilder,
	}
}

func (td *AssetTransactionDao) WithTx(tx *gorm.DB) *AssetTransactionDao {
	if td == nil {
		return td
	}
	newDao := *td
	newDao.db = infra.Database{DB: tx}
	newDao.assetDao = td.assetDao.WithTx(tx)
	newDao.billDao = td.billDao.WithTx(tx)
	return &newDao
}

func (td *AssetTransactionDao) BatchTransaction(ctx context.Context, transactions []*AssetTransaction, orderType common.TransactionRecordType, allowNeg bool) error {

	var err error

	for i := 0; i < len(transactions); i++ {

		switch transactions[i].TransactionType {
		case common.TRANSACTION_TYPE_ASSET_ADD:
			err = td.assetDao.AddAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount)
		case common.TRANSACTION_TYPE_ASSET_DEDUCT:
			err = td.assetDao.DeductAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_ASSET_FREEZE:
			err = td.assetDao.FreezeAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_ASSET_UNFREEZE:
			err = td.assetDao.UnfreezeAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_ADD:
			err = td.assetDao.AddFreezedAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount)
		case common.TRANSACTION_TYPE_FROZEN_ASSET_DEDUCT:
			err = td.assetDao.DeductFreezeAssets(ctx, transactions[i].UserID, transactions[i].WalletID, transactions[i].CategoryID, transactions[i].Currency, transactions[i].Amount, allowNeg)
		default:
			return td.beBuilder.NewBusinessError(ctx, common.CODE_NO_SUCH_TRANSACTION_TYPE)
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
	asset, err = td.assetDao.GetByUserIDCategoryIDCardID(ctx, transaction.UserID, transaction.CategoryID, transaction.WalletID)
	if err != nil {
		return err
	}

	bill := &Bill{
		UserID:              transaction.UserID,
		OrderNo:             transaction.OrderNO,
		AssetID:             asset.ID,
		Amount:              transaction.Amount,
		CurrentAmount:       asset.Amount,
		CurrentFrozenAmount: asset.FrozenAmount,
		Currency:            transaction.Currency,
		TransactionType:     transaction.TransactionType,
		BillType:            transaction.BillType,
		OrderType:           orderType,
	}

	_, err = td.billDao.Save(ctx, bill)
	if err != nil {
		return err
	}

	return nil
}
