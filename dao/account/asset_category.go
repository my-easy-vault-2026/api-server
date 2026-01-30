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

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type AssetCategory struct {
	ID                uint64 `gorm:"primarykey"`
	Name              string
	PreferredName     string `gorm:"default:null"`
	SecondaryName     string `gorm:"default:null"`
	Type              common.AssetType
	CardType          common.CardType
	Currency          common.Currency
	CurrencyType      common.CurrencyType
	ActivationDeposit decimal.Decimal   `gorm:"default:null"`
	Format            common.CardFormat `gorm:"default:null"`
	SpendLimit        decimal.Decimal   `gorm:"default:null"`
	ValidMonths       int               `gorm:"default:null"`
	Design            string            `gorm:"default:null"`
	Fee               decimal.Decimal   `gorm:"default:null"`
	FeeCurrency       common.Currency   `gorm:"default:null"`
	UserID            uint64
	Amount            decimal.Decimal
	CreatedAt         time.Time `gorm:"default:null"`
	UpdatedAt         time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type AssetCategoryQuery struct {
	AssetCategory
	Attrs     AssetCategory
	IDIn      []uint64
	ForUpdate bool
	ForShare  bool
	utils.Page
}

type AssetCategoryDao struct {
}

func (AssetCategory) TableName() string {
	return "asset_category"
}

func NewAssetCategoryDao() *AssetCategoryDao {
	return &AssetCategoryDao{}
}

func (md *AssetCategoryDao) ListByIDs(ctx context.Context, ids []uint64) ([]*AssetCategory, error) {
	result := make([]*AssetCategory, 0)
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(AssetCategory{}).
			Scopes(md.queryChain(&AssetCategoryQuery{
				IDIn: ids,
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*AssetCategory, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return []*AssetCategory{}, tx.Error
			}
			return result, nil
		}, &utils.L2CacheConfig{
			Level:         []common.L2CacheLevel{common.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: utils.Config.System.L2CacheExpireSeconds,
		})
}

func (md *AssetCategoryDao) Get(ctx context.Context, query *AssetCategoryQuery) (*AssetCategory, error) {
	result := &AssetCategory{}
	db := utils.GetDB(ctx)

	err := db.
		Model(AssetCategory{}).
		Scopes(md.queryChain(query)).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (md *AssetCategoryDao) Gets(ctx context.Context, query *AssetCategoryQuery) ([]*AssetCategory, error) {
	result := make([]*AssetCategory, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(AssetCategory{}).
		Scopes(md.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*AssetCategory{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (md *AssetCategoryDao) Save(ctx context.Context, model *AssetCategory) error {

	db := utils.GetDB(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return ret.Error
	}
	return nil
}

func (md *AssetCategoryDao) Update(ctx context.Context, query *AssetCategoryQuery) (int64, error) {

	if query.ID == 0 {
		// TODO: define dao error
		return 0, errors.ErrUnsupported
	}

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
			}
		default:
			continue
		}
	}

	ret := db.
		Model(AssetCategory{}).
		Where("id = ?", query.ID).
		Scopes(md.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ad *AssetCategoryDao) queryChain(query *AssetCategoryQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.AssetCategory)
		structValue := reflect.ValueOf(query.AssetCategory)
		structPtrValue := reflect.ValueOf(&query.AssetCategory)
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

		if query.IDIn != nil {
			db.Scopes(ad.inScope("id", query.IDIn))
		}

		return db
	}
}

func (ad *AssetCategoryDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *AssetCategoryDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
