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
	"gorm.io/gorm/schema"
)

type Bill struct {
	ID                  uint64
	UserID              uint64
	AssetID             uint64
	OrderNo             string
	Amount              decimal.Decimal
	CurrentAmount       decimal.Decimal
	CurrentFrozenAmount decimal.Decimal
	Currency            common.Currency
	TransactionType     common.TransactionType
	BillType            common.BillType
	OrderType           common.TransactionRecordType
	CreatedAt           time.Time `gorm:"default:null"`
	UpdatedAt           time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type BillQuery struct {
	Bill
	Attrs          Bill
	IDIn           []uint64
	UserIDIn       []uint64
	AssetIDIn      []uint64
	NotEqual       map[string]interface{}
	ForUpdate      bool
	ForShare       bool
	OrderBy        []string
	OrderDirection common.OrderDirection
	CreatedAtFrom  time.Time
	CreatedAtTo    time.Time
	common.Page
}

type BillDao struct {
	db        infra.Database
	env       *lib.Env
	beBuilder *lib.BEBuilder
}

func NewBillDao(db infra.Database, env *lib.Env, beBuilder *lib.BEBuilder) *BillDao {
	return &BillDao{db: db, env: env, beBuilder: beBuilder}
}

func (bd *BillDao) WithTx(tx *gorm.DB) *BillDao {
	if bd == nil {
		return bd
	}
	newDao := *bd
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Bill) TableName() string {
	return "bill"
}

func (bd *BillDao) Get(ctx context.Context, query *BillQuery) (*Bill, error) {
	result := &Bill{}
	db := bd.db.WithContext(ctx)

	err := db.
		Model(Bill{}).
		Scopes(bd.queryChain(query)).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (bd *BillDao) Gets(ctx context.Context, query *BillQuery) ([]*Bill, error) {
	result := make([]*Bill, 0)
	db := bd.db.WithContext(ctx)

	err := db.
		Model(Bill{}).
		Scopes(bd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Bill{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BillDao) Save(ctx context.Context, model *Bill, clauses ...utils.Clause) (uint64, error) {

	db := bd.db.WithContext(ctx)

	for _, c := range clauses {
		db = db.Scopes(c.Scope())
	}

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (bd *BillDao) queryChain(query *BillQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Bill)
		structValue := reflect.ValueOf(query.Bill)
		structPtrValue := reflect.ValueOf(&query.Bill)
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
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(bd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(bd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if !query.CreatedAtFrom.IsZero() && query.CreatedAtFrom.Unix() != 0 {
			db.Scopes(bd.compareScope("created_at", query.CreatedAtFrom, true, true))
		}

		if !query.CreatedAtTo.IsZero() && query.CreatedAtTo.Unix() != 0 {
			db.Scopes(bd.compareScope("created_at", query.CreatedAtTo, false, true))
		}

		if query.IDIn != nil {
			db.Scopes(bd.inScope("id", query.IDIn))
		}

		if query.AssetIDIn != nil {
			db.Scopes(bd.inScope("asset_id", query.AssetIDIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(bd.inScope("user_id", query.UserIDIn))
		}

		if len(query.NotEqual) > 0 {
			db.Scopes(bd.notEqualScope(query.NotEqual))
		}

		return db.
			Scopes(bd.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(bd.pageScope(query.Current, query.PageSize))
	}
}

func (bd *BillDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (bd *BillDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (bd *BillDao) orderByScope(fieldName []string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var orderQuery string
		for _, f := range fieldName {
			if f != "" && direction != 0 {
				if orderQuery != "" {
					orderQuery += ","
				}
				orderQuery += fmt.Sprintf("`%s` %s", f, direction.String())
			}
		}
		if orderQuery != "" {
			return db.Order(orderQuery)
		}
		return db
	}
}

func (bd *BillDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && size != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}

func (bd *BillDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (bd *BillDao) notEqualScope(notEqual map[string]interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for fieldName, field := range notEqual {
			if fieldName == "" {
				continue
			}
			db = db.Where(fmt.Sprintf("%s != ?", fieldName), field)
		}
		return db
	}
}
