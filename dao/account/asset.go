package account

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"api-server/infra"
	"api-server/lib"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Asset struct {
	ID            uint64
	CategoryID    uint64
	Type          common.AssetType
	UserID        uint64
	Currency      common.Currency
	CurrencyType  common.CurrencyType
	Amount        decimal.Decimal
	FreezedAmount decimal.Decimal
	Signature     string
	DeletedAt     time.Time `gorm:"default:null"`
	CreatedAt     time.Time `gorm:"default:null"`
	UpdatedAt     time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type AssetQuery struct {
	Asset
	Attrs          Asset
	CategoryIDIn   []uint64
	IDIn           []uint64
	TypeIn         []common.AssetType
	CurrencyIn     []common.Currency
	UserIDIn       []uint64
	ForUpdate      bool
	ForShare       bool
	Deleted        bool
	OrderBy        string
	OrderDirection common.OrderDirection
	common.Page
}

type AssetDao struct {
	db        infra.Database
	env       *lib.Env
	beBuilder *lib.BEBuilder
}

func NewAssetDao(db infra.Database, env *lib.Env, beBuilder *lib.BEBuilder) *AssetDao {
	return &AssetDao{db: db, env: env, beBuilder: beBuilder}
}

func (ad *AssetDao) WithTx(tx *gorm.DB) *AssetDao {
	if ad == nil {
		return ad
	}
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Asset) TableName() string {
	return "asset"
}

func (ad *AssetDao) AddAssets(ctx context.Context, userId uint64, cardID uint64, categoryID uint64, currency common.Currency, amount decimal.Decimal) error {

	db := ad.db.WithContext(ctx)
	var ret *gorm.DB

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
		ret = db.Update("amount", gorm.Expr("amount + ?", amount))
	} else {
		result := &Asset{}
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)

		// for update 鎖定資料
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewBusinessError(ctx, common.CODE_DATA_NOT_EXIST)
		}
		//取得新的 hash值
		addAmount := result.Amount.Add(amount)

		ret = db.Updates(map[string]interface{}{
			"amount": addAmount,
		})

	}

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_ADD_ASSETS_FAILED)
	}

	return nil
}

func (ad *AssetDao) DeductAssets(
	ctx context.Context,
	userId uint64,
	cardID uint64,
	categoryID uint64,
	currency common.Currency,
	amount decimal.Decimal,
	allowNeg bool) error {

	db := ad.db.WithContext(ctx)

	var ret *gorm.DB

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
		ret = db.Update("amount", gorm.Expr("amount - ?", amount))
	} else {
		result := &Asset{}
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)
		if !allowNeg {
			db = db.Where("amount >= ?", amount)
		}

		// for update 鎖定資料
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewBusinessError(ctx, common.CODE_DATA_NOT_EXIST)
		}
		//取得新的 hash值
		subAmount := result.Amount.Sub(amount)
		ret = db.Updates(map[string]interface{}{
			"amount": subAmount,
		})

	}

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_DEDUCT_ASSETS_FAILED)
	}

	return nil
}
func (ad *AssetDao) AddFreezedAssets(ctx context.Context, userId uint64, cardID uint64, categoryID uint64, currency common.Currency, amount decimal.Decimal) error {

	db := ad.db.WithContext(ctx)

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
	} else {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)
	}

	ret := db.Update("freezed_amount", gorm.Expr("freezed_amount + ?", amount))

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_ADD_FREEZED_ASSETS_FAILED)
	}

	return nil
}

func (ad *AssetDao) DeductFreezeAssets(
	ctx context.Context,
	userId uint64,
	cardID uint64,
	categoryID uint64,
	currency common.Currency,
	amount decimal.Decimal,
	allowNeg bool) error {

	db := ad.db.WithContext(ctx)

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
	} else {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)
		if !allowNeg {
			db = db.Where("freezed_amount >= ?", amount)
		}
	}
	ret := db.Update("freezed_amount", gorm.Expr("freezed_amount - ?", amount))

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_DEDUCT_FREEZED_ASSETS_FAILED)
	}

	return nil
}

func (ad *AssetDao) FreezeAssets(
	ctx context.Context,
	userId uint64,
	cardID uint64,
	categoryID uint64,
	currency common.Currency,
	amount decimal.Decimal,
	allowNeg bool) error {

	db := ad.db.WithContext(ctx)

	updates := map[string]interface{}{
		"amount":         gorm.Expr("amount - ?", amount),
		"freezed_amount": gorm.Expr("freezed_amount + ?", amount),
	}

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
	} else {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)
		if !allowNeg {
			db = db.Where("amount >= ?", amount)
		}
		result := &Asset{}

		// for update 鎖定資料
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewBusinessError(ctx, common.CODE_DATA_NOT_EXIST)
		}
	}

	ret := db.Updates(updates)
	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_FREEZE_ASSETS_FAILED)
	}

	return nil
}

func (ad *AssetDao) UnfreezeAssets(
	ctx context.Context,
	userId uint64,
	cardID uint64,
	categoryID uint64,
	currency common.Currency,
	amount decimal.Decimal,
	allowNeg bool) error {

	db := ad.db.WithContext(ctx)

	updates := map[string]interface{}{
		"amount":         gorm.Expr("amount + ?", amount),
		"freezed_amount": gorm.Expr("freezed_amount - ?", amount),
	}

	if common.IsSystemAccount(userId) {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND currency=?", userId, currency)
	} else {
		db = db.
			Model(Asset{}).
			Where("user_id=? AND id=? AND category_id=? AND currency=?", userId, cardID, categoryID, currency)

		if !allowNeg {
			db = db.Where("freezed_amount >= ?", amount)
		}
		result := &Asset{}

		// for update 鎖定資料
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewBusinessError(ctx, common.CODE_DATA_NOT_EXIST)
		}
	}

	ret := db.Updates(updates)
	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected != 1 && !amount.IsZero() {
		return utils.NewBusinessError(ctx, common.CODE_UNFREEZE_ASSETS_FAILED)
	}

	return nil
}

func (ad *AssetDao) GetByUserIDCardID(ctx context.Context, userID uint64, cardID uint64) (*Asset, error) {
	if userID == 0 || cardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Where("user_id=? AND id=?", userID, cardID).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDCategoryID(ctx context.Context, userID uint64, categoryID uint64) (*Asset, error) {
	if userID == 0 || categoryID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID:     userID,
				CategoryID: categoryID,
			},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDTypeCurrency(ctx context.Context, userID uint64, t common.AssetType, currency common.Currency) (*Asset, error) {
	if userID == 0 || currency == 0 || t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID:   userID,
				Type:     t,
				Currency: currency,
			},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDCategoryIDCardID(ctx context.Context, userID uint64, categoryID uint64, cardID uint64) (*Asset, error) {
	if userID == 0 || categoryID == 0 || cardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				ID:         cardID,
				UserID:     userID,
				CategoryID: categoryID,
			},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByIDForUpdate(ctx context.Context, id uint64) (*Asset, error) {
	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				ID: id,
			},
			ForUpdate: true,
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByIDUserID(ctx context.Context, id uint64, userID uint64) (*Asset, error) {
	if id == 0 || userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				ID:     id,
				UserID: userID,
			},
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByID(ctx context.Context, id uint64) (*Asset, error) {
	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				ID: id,
			},
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDTypeInCurrencyOrderByAmount(ctx context.Context, userID uint64, typeIn []common.AssetType, currency common.Currency) (*Asset, error) {
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID:   userID,
				Currency: currency,
			},
			TypeIn:         typeIn,
			OrderBy:        "amount",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDTypeInCurrencyCategoryIDInOrderByAmount(ctx context.Context, userID uint64, typeIn []common.AssetType, currency common.Currency, categoryIDIn []uint64) (*Asset, error) {
	if userID == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	if len(typeIn) == 0 || len(categoryIDIn) == 0 {
		return nil, nil
	}

	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID:   userID,
				Currency: currency,
			},
			TypeIn:         typeIn,
			CategoryIDIn:   categoryIDIn,
			OrderBy:        "amount",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDCurrency(ctx context.Context, userID uint64, currency common.Currency) (*Asset, error) {
	if userID == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID:   userID,
				Currency: currency,
			},
		})).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserID(ctx context.Context, userID uint64) ([]*Asset, error) {
	if userID == 0 {
		return make([]*Asset, 0), nil
	}
	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*Asset, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByUserIDTypesIn(ctx context.Context, userID uint64, types []common.AssetType) ([]*Asset, error) {
	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	if len(types) == 0 {
		return make([]*Asset, 0), nil
	}
	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*Asset, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetByCurrency(ctx context.Context, currency common.Currency) ([]*Asset, error) {

	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				Currency: currency,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) GetTotalAmountByCurrency(ctx context.Context, currency common.Currency) (decimal.Decimal, error) {

	var totalAmount decimal.Decimal
	db := ad.db.WithContext(ctx)

	err := db.
		Model(&Asset{}).
		Select("COALESCE(SUM(amount), 0) + COALESCE(SUM(freezed_amount), 0) AS total").
		Where("currency = ?", currency).
		Scan(&totalAmount).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}
	return totalAmount, nil
}

func (ad *AssetDao) ListByUserID(ctx context.Context, userID uint64) ([]*Asset, error) {
	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) ListByTypeCurrencyInUserIDIn(ctx context.Context, t common.AssetType, currencyIn []common.Currency, userIDIn []uint64) ([]*Asset, error) {

	if len(currencyIn) == 0 || len(userIDIn) == 0 {
		return []*Asset{}, nil
	}

	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				Type: t,
			},
			CurrencyIn: currencyIn,
			UserIDIn:   userIDIn,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) ListByIDIn(ctx context.Context, ids []uint64) ([]*Asset, error) {
	if len(ids) == 0 {
		return []*Asset{}, nil
	}
	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			IDIn: ids,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) Page(ctx context.Context, pageCurrent int, pageSize int) (records []*Asset, current int, size int, total int, err error) {
	result := make([]*Asset, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Deleted: true,
		})).
		Count(&s).
		Scopes(ad.queryChain(&AssetQuery{
			Deleted: true,
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ad *AssetDao) PageOrderByID(ctx context.Context, pageCurrent int, pageSize int) (records []*Asset, current int, size int, total int, err error) {
	result := make([]*Asset, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{})).
		Count(&s).
		Scopes(ad.queryChain(&AssetQuery{
			OrderBy:        "id",
			OrderDirection: common.ORDER_DIRECTION_ASC,
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ad *AssetDao) PageByTypeCurrencyInOrderByID(ctx context.Context, t common.AssetType, currencyIn []common.Currency, pageCurrent int, pageSize int) (records []*Asset, current int, size int, total int, err error) {

	if len(currencyIn) == 0 {
		return make([]*Asset, 0), pageCurrent, pageSize, 0, nil
	}

	result := make([]*Asset, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{})).
		Count(&s).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				Type: t,
			},
			CurrencyIn:     currencyIn,
			OrderBy:        "id",
			OrderDirection: common.ORDER_DIRECTION_ASC,
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ad *AssetDao) ListByUserIDType(ctx context.Context, userID uint64, t common.AssetType) ([]*Asset, error) {

	if userID == 0 || t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID: userID,
				Type:   t,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) ListByUserIDTypeCurrencyIn(ctx context.Context, userID uint64, t common.AssetType, currencyIn []common.Currency) ([]*Asset, error) {

	if userID == 0 || t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	if len(currencyIn) == 0 {
		return []*Asset{}, nil
	}

	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(&AssetQuery{
			Asset: Asset{
				UserID: userID,
				Type:   t,
			},
			CurrencyIn: currencyIn,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) Get(ctx context.Context, query *AssetQuery) (*Asset, error) {
	result := &Asset{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(query)).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDao) Gets(ctx context.Context, query *AssetQuery) ([]*Asset, error) {
	result := make([]*Asset, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Asset{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Asset{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDao) Save(ctx context.Context, model *Asset, clauses ...utils.Clause) (uint64, error) {

	db := ad.db.WithContext(ctx)

	for _, c := range clauses {
		db = db.Scopes(c.Scope())
	}

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *AssetDao) SoftDeleteByID(ctx context.Context, id uint64) (int64, error) {

	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	return ad.Update(ctx, &AssetQuery{
		Asset: Asset{
			ID: id,
		},
		Attrs: Asset{
			DeletedAt: utils.DBQueryTime(time.Now()),
		},
	})
}

func (ad *AssetDao) DeleteByID(ctx context.Context, id uint64) (int64, error) {
	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	db := ad.db.WithContext(ctx)

	ret := db.
		Delete(&Asset{
			ID: id,
		})

	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ad *AssetDao) Update(ctx context.Context, query *AssetQuery) (int64, error) {

	if query.ID == 0 {
		// TODO: define dao error
		return 0, errors.ErrUnsupported
	}

	db := ad.db.WithContext(ctx)

	attrs := map[string]interface{}{}
	structType := reflect.TypeOf(query.Attrs)
	structValue := reflect.ValueOf(query.Attrs)
	structPtrValue := reflect.ValueOf(&query.Attrs)
	for i := 0; i < structType.NumField(); i++ {

		if structValue.Field(i).IsZero() {
			continue
		}
		fieldName := "`" + stringy.New(structType.Field(i).Name).SnakeCase().ToLower() + "`"
		settings := schema.ParseTagSetting(structType.Field(i).Tag.Get("gorm"), ";")
		if f, ok := settings["COLUMN"]; ok {
			fieldName = "`" + f + "`"
		}
		switch structValue.Field(i).Kind() {
		case reflect.String:
			attrs[fieldName] = structValue.Field(i).String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			attrs[fieldName] = int(structValue.Field(i).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			attrs[fieldName] = int(structValue.Field(i).Uint())
		case reflect.Float32, reflect.Float64:
			attrs[fieldName] = structValue.Field(i).Float()
		case reflect.Bool:
			attrs[fieldName] = structValue.Field(i).Bool()
		case reflect.Interface:
			attrs[fieldName] = structValue.Field(i).Interface()
		case reflect.Pointer:
			attrs[fieldName] = structValue.Field(i).Interface()

		case reflect.Struct:
			ptr := structPtrValue.Elem().Field(i).Addr().Interface()
			switch reflect.TypeOf(ptr) {
			case reflect.TypeOf((*sql.NullBool)(nil)),
				reflect.TypeOf((*sql.NullInt16)(nil)),
				reflect.TypeOf((*sql.NullInt32)(nil)),
				reflect.TypeOf((*sql.NullInt64)(nil)),
				reflect.TypeOf((*sql.NullString)(nil)),
				reflect.TypeOf((*sql.NullFloat64)(nil)),
				reflect.TypeOf((*sql.NullByte)(nil)),
				reflect.TypeOf((*sql.NullString)(nil)),
				reflect.TypeOf((*sql.NullTime)(nil)),
				reflect.TypeOf((*decimal.NullDecimal)(nil)):
				valuer := ptr.(driver.Valuer)
				value, _ := valuer.Value()
				if value == nil {
					continue
				}
				attrs[fieldName] = value
			case reflect.TypeOf((*time.Time)(nil)):
				t := ptr.(*time.Time)
				if t.IsZero() {
					continue
				}
				attrs[fieldName] = t
			}
		default:
			continue
		}
	}

	ret := db.
		Model(Asset{}).
		Where("id = ?", query.ID).
		Scopes(ad.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ad *AssetDao) queryChain(query *AssetQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Asset)
		structValue := reflect.ValueOf(query.Asset)
		structPtrValue := reflect.ValueOf(&query.Asset)
		for i := 0; i < structType.NumField(); i++ {
			if structValue.Field(i).IsZero() {
				continue
			}
			fieldName := stringy.New(structType.Field(i).Name).SnakeCase().ToLower()
			settings := schema.ParseTagSetting(structType.Field(i).Tag.Get("gorm"), ";")
			if f, ok := settings["COLUMN"]; ok {
				fieldName = f
			}
			switch structValue.Field(i).Kind() {
			case reflect.String:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Interface()))
			case reflect.Struct:
				ptr := structPtrValue.Elem().Field(i).Addr().Interface()
				switch reflect.TypeOf(ptr) {
				case reflect.TypeOf((*sql.NullBool)(nil)),
					reflect.TypeOf((*sql.NullInt16)(nil)),
					reflect.TypeOf((*sql.NullInt32)(nil)),
					reflect.TypeOf((*sql.NullInt64)(nil)),
					reflect.TypeOf((*sql.NullString)(nil)),
					reflect.TypeOf((*sql.NullFloat64)(nil)),
					reflect.TypeOf((*sql.NullByte)(nil)),
					reflect.TypeOf((*sql.NullString)(nil)),
					reflect.TypeOf((*sql.NullTime)(nil)),
					reflect.TypeOf((*decimal.NullDecimal)(nil)):
					valuer := ptr.(driver.Valuer)
					value, _ := valuer.Value()
					if value == nil {
						continue
					}
					db.Scopes(ad.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.ForUpdate {
			db.Scopes(ad.forScope("UPDATE"))
		}

		if query.ForShare {
			db.Scopes(ad.forScope("SHARE"))
		}

		if !query.Deleted {
			db.Scopes(ad.nullScope([]string{"deleted_at"}, true))
		}

		if query.TypeIn != nil {
			db.Scopes(ad.inScope("type", query.TypeIn))
		}

		if query.CurrencyIn != nil {
			db.Scopes(ad.inScope("currency", query.CurrencyIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(ad.inScope("user_id", query.UserIDIn))
		}

		return db.
			Scopes(ad.categoryIDInScope(query.CategoryIDIn)).
			Scopes(ad.IDInScope(query.IDIn)).
			Scopes(ad.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(ad.pageScope(query.Current, query.PageSize))
	}
}

func (ad *AssetDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *AssetDao) categoryIDInScope(categories []uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(categories) == 0 {
			return db
		}
		return db.Where("category_id IN ? ", categories)
	}
}

func (ad *AssetDao) IDInScope(ids []uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(ids) == 0 {
			return db
		}
		return db.Where("id IN ? ", ids)
	}
}

func (ad *AssetDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (ad *AssetDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (ad *AssetDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(fieldNames) != 0 {
			for _, fieldName := range fieldNames {
				if fieldName == "" {
					continue
				}
				determinator := "IS"
				if !isNull {
					determinator = "IS NOT"
				}
				db = db.Where(gorm.Expr(fmt.Sprintf("`%s` %s NULL", fieldName, determinator)))
			}
		}
		return db
	}
}

func (ad *AssetDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if greater {
			if equal {
				return db.Where(fmt.Sprintf("`%s` >= ? ", fieldName), field)
			}
			return db.Where(fmt.Sprintf("`%s` > ? ", fieldName), field)
		}
		if equal {
			return db.Where(fmt.Sprintf("`%s` <= ? ", fieldName), field)
		}
		return db.Where(fmt.Sprintf("`%s` < ? ", fieldName), field)
	}
}

func (ad *AssetDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (ad *AssetDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
