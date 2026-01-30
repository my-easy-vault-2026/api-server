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

type ManualOrder struct {
	ID              uint64
	BusinessNO      string
	OrderNO         string
	UserID          uint64
	AssetID         uint64
	Currency        common.Currency
	Amount          decimal.Decimal
	AgainstUserID   uint64 `gorm:"default:null"`
	AgainstAssetID  uint64 `gorm:"default:null"`
	TransactionType common.TransactionType
	CreatedBy       string `gorm:"default:null"`
	CreatedByID     string `gorm:"default:null"`
	ReviewedBy      string `gorm:"default:null"`
	ReviewedByID    string `gorm:"default:null"`
	Memo            string `gorm:"default:null"`
	ReviewMemo      string `gorm:"default:null"`
	Status          common.ManualStatus
	CreatedAt       time.Time `gorm:"default:null"`
	UpdatedAt       time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type ManualOrderQuery struct {
	ManualOrder
	Attrs     ManualOrder
	ForUpdate bool
	ForShare  bool
	StatusIn  []common.ExchangeStatus
	utils.Page
}
type ManualOrderDao struct {
	// Add any necessary fields or methods here
}

func NewManualOrderDao() *ManualOrderDao {
	return &ManualOrderDao{}
}

func (ManualOrder) TableName() string {
	return "manual_order"
}

func (ed *ManualOrderDao) Save(ctx context.Context, model *ManualOrder) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(ManualOrder{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ed *ManualOrderDao) Update(ctx context.Context, query *ManualOrderQuery) (int64, error) {
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
		Model(&ManualOrder{}).
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

func (ed *ManualOrderDao) queryChain(query *ManualOrderQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.ManualOrder)
		structValue := reflect.ValueOf(query.ManualOrder)
		structPtrValue := reflect.ValueOf(&query.ManualOrder)
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

func (ed *ManualOrderDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ed *ManualOrderDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
