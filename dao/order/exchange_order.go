package order

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ExchangeOrder struct {
	ID             uint64
	OrderNO        string
	UserID         uint64
	ToAmount       decimal.Decimal
	ToCardID       uint64
	ToCategoryID   uint64
	ToCurrency     common.Currency
	FromAmount     decimal.Decimal
	FromCardID     uint64
	FromCategoryID uint64
	FromCurrency   common.Currency
	ExchangeRate   decimal.Decimal
	Fee            decimal.Decimal
	TriggerMode    common.ExchangeTriggerMode `gorm:"default:null"`
	Status         common.ExchangeStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ExchangeOrderQuery struct {
	ExchangeOrder
	Attrs     ExchangeOrder
	ForUpdate bool
	ForShare  bool
	StatusIn  []common.ExchangeStatus
	utils.Page
}
type ExchangeOrderDao struct {
	// Add any necessary fields or methods here
}

func NewExchangeOrderDao() *ExchangeOrderDao {
	return &ExchangeOrderDao{}
}

func (ExchangeOrder) TableName() string {
	return "exchange_order"
}
func (ed *ExchangeOrderDao) Save(ctx context.Context, model *ExchangeOrder) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(ExchangeOrder{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ed *ExchangeOrderDao) Update(ctx context.Context, query *ExchangeOrderQuery) (int64, error) {
	db := utils.GetDB(ctx)
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
		Model(&ExchangeOrder{}).
		Scopes(ed.queryChain(query)).
		Updates(attrs)
	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if ret.Error != nil {
		return 0, ret.Error
	}
	return ret.RowsAffected, nil
}

func (ed *ExchangeOrderDao) queryChain(query *ExchangeOrderQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.ExchangeOrder)
		structValue := reflect.ValueOf(query.ExchangeOrder)
		structPtrValue := reflect.ValueOf(&query.ExchangeOrder)
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
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Interface:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Interface()))
			case reflect.Pointer:
				db.Scopes(ed.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(ed.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}
		if query.StatusIn != nil {
			db.Scopes(ed.inScope("status", query.StatusIn))
		}
		return db
	}
}

func (ed *ExchangeOrderDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ed *ExchangeOrderDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
