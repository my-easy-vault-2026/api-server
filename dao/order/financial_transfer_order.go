package order

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
	"gorm.io/gorm/schema"
)

type FinancialTransferOrder struct {
	ID             uint64                                `gorm:"default:null"`
	OrderNO        string                                `gorm:"default:null"`
	FromType       common.AssetType                      `gorm:"default:null"`
	FromAmount     decimal.Decimal                       `gorm:"default:null"`
	FromUserID     uint64                                `gorm:"default:null"`
	FromCardID     uint64                                `gorm:"default:null"`
	FromCategoryID uint64                                `gorm:"default:null"`
	FromCurrency   common.Currency                       `gorm:"default:null"`
	ToType         common.AssetType                      `gorm:"default:null"`
	ToAmount       decimal.Decimal                       `gorm:"default:null"`
	ToUserID       uint64                                `gorm:"default:null"`
	ToCardID       uint64                                `gorm:"default:null"`
	ToCategoryID   uint64                                `gorm:"default:null"`
	ToCurrency     common.Currency                       `gorm:"default:null"`
	Direction      common.FinancialTransferDirection     `gorm:"default:null"`
	TriggerMethod  common.FinancialTransferTriggerMethod `gorm:"default:null"`
	Status         common.FinancialTransferStatus        `gorm:"default:null"`
	CreatedAt      time.Time                             `gorm:"default:null"`
	UpdatedAt      time.Time                             `gorm:"default:null;autoUpdateTime:false"`
}

type FinancialTransferOrderQuery struct {
	FinancialTransferOrder
	Attrs          FinancialTransferOrder
	ForUpdate      bool
	ForShare       bool
	StatusIn       []common.FinancialTransferStatus
	OrderBy        string
	OrderDirection common.OrderDirection
	utils.Page
}
type FinancialTransferOrderDao struct {
	db  infra.Database
	env *lib.Env
}

func NewFinancialTransferOrderDao(db infra.Database, env *lib.Env) *FinancialTransferOrderDao {
	return &FinancialTransferOrderDao{db: db, env: env}
}

func (trd *FinancialTransferOrderDao) WithTx(tx *gorm.DB) *FinancialTransferOrderDao {
	newDao := *trd
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (FinancialTransferOrder) TableName() string {
	return "financial_transfer_order"
}

func (trd *FinancialTransferOrderDao) Save(ctx context.Context, model *FinancialTransferOrder) (uint64, error) {

	db := trd.db.WithContext(ctx)

	ret := db.
		Model(FinancialTransferOrder{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *FinancialTransferOrderDao) GetByFromCardIDOrderByCreatedAtDESC(ctx context.Context, fromCardID uint64) (*FinancialTransferOrder, error) {
	result := &FinancialTransferOrder{}
	db := trd.db.WithContext(ctx)

	err := db.
		Model(FinancialTransferOrder{}).
		Scopes(trd.queryChain(&FinancialTransferOrderQuery{
			FinancialTransferOrder: FinancialTransferOrder{
				FromCardID: fromCardID,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Limit(1).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *FinancialTransferOrderDao) Update(ctx context.Context, query *FinancialTransferOrderQuery) (int64, error) {
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
		Model(&FinancialTransferOrder{}).
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

func (trd *FinancialTransferOrderDao) queryChain(query *FinancialTransferOrderQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.FinancialTransferOrder)
		structValue := reflect.ValueOf(query.FinancialTransferOrder)
		structPtrValue := reflect.ValueOf(&query.FinancialTransferOrder)
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
			case reflect.Interface:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Interface()))
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
		return db.
			Scopes(trd.orderByScope(query.OrderBy, query.OrderDirection))
	}
}

func (trd *FinancialTransferOrderDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (trd *FinancialTransferOrderDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *FinancialTransferOrderDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}
