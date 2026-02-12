package order

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/my-easy-vault-2026/shared-modules/common"

	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type TransactionRecordDao struct {
	db  infra.Database
	env *lib.Env
}

func NewTransactionRecordDao(db infra.Database, env *lib.Env) *TransactionRecordDao {
	return &TransactionRecordDao{db: db, env: env}
}

func (trd *TransactionRecordDao) WithTx(tx *gorm.DB) *TransactionRecordDao {
	newDao := *trd
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

type TransactionRecord struct {
	ID                      uint64
	Type                    common.TransactionRecordType
	TransactionNO           string
	UserID                  uint64
	WalletID                uint64
	Income                  decimal.NullDecimal
	IncomeCategoryID        uint64
	IncomeCurrency          common.Currency
	AgainstIncome           decimal.NullDecimal    `gorm:"default:null"`
	AgainstIncomeCategoryID uint64                 `gorm:"default:null"`
	AgainstIncomeCurrency   common.Currency        `gorm:"default:null"`
	Side                    common.TransactionSide `gorm:"default:null"`
	FromType                common.AssetType       `gorm:"default:null"`
	FromWalletID            uint64                 `gorm:"default:null"`
	FromCategoryID          uint64                 `gorm:"default:null"`
	FromCurrency            common.Currency        `gorm:"default:null"`
	FromAmount              decimal.NullDecimal    `gorm:"default:null"`
	FromDiscount            decimal.NullDecimal    `gorm:"default:null"`
	FromUserID              uint64                 `gorm:"default:null"`
	ToType                  common.AssetType       `gorm:"default:null"`
	ToWalletID              uint64                 `gorm:"default:null"`
	ToCategoryID            uint64                 `gorm:"default:null"`
	ToCurrency              common.Currency        `gorm:"default:null"`
	ToAmount                decimal.NullDecimal    `gorm:"default:null"`
	ToUserID                uint64                 `gorm:"default:null"`
	ExchangeRate            *decimal.Decimal       `gorm:"default:null"`
	ExchangeFee             *decimal.Decimal       `gorm:"default:null"`
	ExchangeFeeCurrency     common.Currency        `gorm:"default:null"`
	TransferFee             *decimal.Decimal       `gorm:"default:null"`
	TransferFeeCurrency     common.Currency        `gorm:"default:null"`
	Status                  common.TransactionStatus
	CreatedAt               time.Time `gorm:"default:null"`
	UpdatedAt               time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type TransactionRecordQuery struct {
	TransactionRecord
	Attrs                   TransactionRecord
	StatusIn                []common.TransactionStatus
	ForUpdate               bool
	ForShare                bool
	OrderBy                 string
	OrderDirection          common.OrderDirection
	UserIDIn                []uint64
	TransactionNOIn         []string
	TransactionRecordTypeIn []common.TransactionRecordType
	CreatedAtFrom           time.Time
	CreatedAtTo             time.Time
	IsNull                  []string
	common.Page
}

func (TransactionRecord) TableName() string {
	return "transaction_record"
}

func (trd *TransactionRecordDao) PageByUserIDWalletID(ctx context.Context, userID uint64, walletID uint64, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := trd.db.WithContext(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				WalletID: walletID,
				UserID:   userID,
			},
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			Page: common.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (trd *TransactionRecordDao) Get(ctx context.Context, query *TransactionRecordQuery) (*TransactionRecord, error) {
	result := &TransactionRecord{}
	db := trd.db.WithContext(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) Gets(ctx context.Context, query *TransactionRecordQuery) ([]*TransactionRecord, error) {
	result := make([]*TransactionRecord, 0)
	db := trd.db.WithContext(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*TransactionRecord, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (trd *TransactionRecordDao) Save(ctx context.Context, model *TransactionRecord) (uint64, error) {

	db := trd.db.WithContext(ctx)

	ret := db.
		Model(TransactionRecord{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *TransactionRecordDao) Saves(ctx context.Context, models []*TransactionRecord) (int64, error) {

	db := trd.db.WithContext(ctx)

	ret := db.
		Model(TransactionRecord{}).
		CreateInBatches(models, len(models))

	if ret.Error != nil {
		return ret.RowsAffected, ret.Error
	}
	return ret.RowsAffected, nil
}

func (trd *TransactionRecordDao) Update(ctx context.Context, query *TransactionRecordQuery) (int64, error) {

	db := trd.db.WithContext(ctx)

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
		kind := structValue.Field(i).Kind()
		switch kind {
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
			default:
				continue
			}
		default:
			continue
		}
	}

	ret := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (trd *TransactionRecordDao) queryChain(query *TransactionRecordQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.TransactionRecord)
		structValue := reflect.ValueOf(query.TransactionRecord)
		structPtrValue := reflect.ValueOf(&query.TransactionRecord)
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
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(trd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.StatusIn != nil {
			db.Scopes(trd.inScope("status", query.StatusIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(trd.inScope("user_id", query.UserIDIn))
		}

		if query.TransactionNOIn != nil {
			db.Scopes(trd.inScope("transaction_no", query.TransactionNOIn))
		}

		if query.TransactionRecordTypeIn != nil {
			db.Scopes(trd.inScope("type", query.TransactionRecordTypeIn))
		}

		if !query.CreatedAtFrom.IsZero() && query.CreatedAtFrom.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtFrom, true, true))
		}

		if !query.CreatedAtTo.IsZero() && query.CreatedAtTo.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtTo, false, true))
		}

		if query.IsNull != nil {
			db.Scopes(trd.nullScope(query.IsNull, true))
		}

		return db.
			Scopes(trd.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(trd.pageScope(query.Current, query.PageSize))
	}
}

func (trd *TransactionRecordDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (trd *TransactionRecordDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (trd *TransactionRecordDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *TransactionRecordDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (trd *TransactionRecordDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (trd *TransactionRecordDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && size != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
