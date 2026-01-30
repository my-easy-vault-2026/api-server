package financial

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

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type FinancialProduct struct {
	ID                  uint64                             `gorm:"default:null"`
	Code                common.FinancialCode               `gorm:"default:null"`
	Name                string                             `gorm:"default:null"`
	DefaultYieldRate    *decimal.Decimal                   `gorm:"default:null"`
	Period              time.Duration                      `gorm:"default:null"` // minute
	SnapshotExpiracy    time.Duration                      `gorm:"default:null"` // minute
	SupportType         common.FinancialProductSupportType `gorm:"default:null"`
	SupportedCurrencies string                             `gorm:"default:null"`
	Status              common.FinancialProductStatus      `gorm:"default:null"`
	CreatedAt           time.Time                          `gorm:"default:null"`
	UpdatedAt           time.Time                          `gorm:"default:null;autoUpdateTime:false"`
}

type FinancialProductQuery struct {
	FinancialProduct
	Attrs     FinancialProduct
	ForUpdate bool
	ForShare  bool
	utils.Page
}

type FinancialProductDao struct {
}

func NewFinancialProductDao() *FinancialProductDao {
	return &FinancialProductDao{}
}

func (FinancialProduct) TableName() string {
	return "financial_product"
}

func (ad *FinancialProductDao) GetByCode(ctx context.Context, code common.FinancialCode) (*FinancialProduct, error) {
	result := &FinancialProduct{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&FinancialProduct{}).
			Scopes(ad.queryChain(&FinancialProductQuery{
				FinancialProduct: FinancialProduct{
					Code: code,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*FinancialProduct, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &utils.L2CacheConfig{
			Level:         []common.L2CacheLevel{common.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: utils.Config.System.L2CacheExpireSeconds,
		})
}

func (ad *FinancialProductDao) List(ctx context.Context) ([]*FinancialProduct, error) {

	result := make([]*FinancialProduct, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(FinancialProduct{}).
		Scopes(ad.queryChain(&FinancialProductQuery{})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*FinancialProduct{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *FinancialProductDao) Get(ctx context.Context, query *FinancialProductQuery) (*FinancialProduct, error) {
	result := &FinancialProduct{}
	db := utils.GetDB(ctx)

	err := db.
		Model(FinancialProduct{}).
		Scopes(ad.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *FinancialProductDao) Gets(ctx context.Context, query *FinancialProductQuery) ([]*FinancialProduct, error) {
	result := make([]*FinancialProduct, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(FinancialProduct{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*FinancialProduct{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *FinancialProductDao) Save(ctx context.Context, model *FinancialProduct) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *FinancialProductDao) Update(ctx context.Context, query *FinancialProductQuery) (int64, error) {

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
		Model(FinancialProduct{}).
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

func (cd *FinancialProductDao) DeleteByID(ctx context.Context, id uint64) (int64, error) {
	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	db := utils.GetDB(ctx)

	ret := db.
		Delete(&FinancialProduct{
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

func (cd *FinancialProductDao) queryChain(query *FinancialProductQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.FinancialProduct)
		structValue := reflect.ValueOf(query.FinancialProduct)
		structPtrValue := reflect.ValueOf(&query.FinancialProduct)
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
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(cd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(cd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.ForUpdate {
			db.Scopes(cd.forScope("UPDATE"))
		}

		if query.ForShare {
			db.Scopes(cd.forScope("SHARE"))
		}

		return db
	}
}

func (cd *FinancialProductDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (cd *FinancialProductDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
	}
}

func (cd *FinancialProductDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (cd *FinancialProductDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (cd *FinancialProductDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (cd *FinancialProductDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (cd *FinancialProductDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
