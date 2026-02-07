package account

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"shared-modules/common"
	"time"

	"api-server/infra"
	"api-server/lib"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Category struct {
	ID           uint64 `gorm:"primarykey"`
	Name         string
	Type         common.AssetType
	Currency     common.Currency
	CurrencyType common.CurrencyType
	Nation       string
	NationCode   common.NationCode
	CreatedAt    time.Time `gorm:"default:null"`
	UpdatedAt    time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type CategoryQuery struct {
	Category
	Attrs     Category
	IDIn      []uint64
	ForUpdate bool
	ForShare  bool
	Select    []string
	Distinct  string
	GroupBy   string
	common.Page
}

type CategoryDao struct {
	db    infra.Database
	env   *lib.Env
	redis infra.Redis
}

func (Category) TableName() string {
	return "asset_category"
}

func NewCategoryDao(db infra.Database, env *lib.Env, redis infra.Redis) *CategoryDao {
	return &CategoryDao{db: db, env: env, redis: redis}
}

func (md *CategoryDao) WithTx(tx *gorm.DB) *CategoryDao {
	if md == nil {
		return md
	}
	newDao := *md
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (md *CategoryDao) GetByID(ctx context.Context, id uint64) (*Category, error) {
	result := &Category{}
	db := md.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, md.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(Category{}).
			Scopes(md.queryChain(&CategoryQuery{
				Category: Category{
					ID: id,
				},
			})).
			First(result)
	}, func(tx *gorm.DB) (*Category, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: md.env.L2CacheExpire,
		})
}

func (md *CategoryDao) List(ctx context.Context) ([]*Category, error) {
	result := make([]*Category, 0)
	db := md.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, md.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(Category{}).
			Scopes(md.queryChain(&CategoryQuery{})).
			Scan(&result)
	}, func(tx *gorm.DB) ([]*Category, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return []*Category{}, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: md.env.L2CacheExpire,
		})
}

func (md *CategoryDao) Gets(ctx context.Context, query *CategoryQuery) ([]*Category, error) {
	result := make([]*Category, 0)
	db := md.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, md.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(Category{}).
			Scopes(md.queryChain(query)).
			Scan(&result)
	}, func(tx *gorm.DB) ([]*Category, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return []*Category{}, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: md.env.L2CacheExpire,
		})
}

func (md *CategoryDao) Save(ctx context.Context, model *Category) (uint64, error) {

	db := md.db.WithContext(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (md *CategoryDao) Update(ctx context.Context, query *CategoryQuery) (int64, error) {

	if query.ID == 0 {
		// TODO: define dao error
		return 0, errors.ErrUnsupported
	}

	db := md.db.WithContext(ctx)

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
		Model(Category{}).
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

func (ad *CategoryDao) queryChain(query *CategoryQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Category)
		structValue := reflect.ValueOf(query.Category)
		structPtrValue := reflect.ValueOf(&query.Category)
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

func (md *CategoryDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (md *CategoryDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (md *CategoryDao) hasScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if field == nil {
			return db
		}
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(field).Len(); i++ {
				db = db.Where(fmt.Sprintf("%d & `%s` > 0", reflect.ValueOf(field).Index(i).Int(), fieldName))
			}
		}
		return db
	}
}
