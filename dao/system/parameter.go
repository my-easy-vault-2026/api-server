package system

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

type Parameter struct {
	ID           uint64                    `json:"id"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	Key          common.ParameterKey       `json:"key" gorm:"column:key"`
	Value        string                    `json:"value"`
	SpecialValue common.SpecialValue       `json:"specialValue" gorm:"default:null"`
	ValueType    common.ParameterValueType `json:"valueType"`
	Currency     common.Currency           `json:"currency" gorm:"default:null"`
	Unit         common.UnitType           `json:"unit" gorm:"default:null"`
	Encrypt      common.EncryptType        `json:"encrypt" gorm:"default:null"`
	CategoryID   common.ParameterCategory  `json:"categoryId"`
	CategoryName string                    `json:"categoryName"`
	Remark       string                    `json:"remark" gorm:"default:null"`
	Status       common.ParameterStatus    `json:"status"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
}

type ParameterQuery struct {
	Parameter
	Attrs     Parameter
	ForUpdate bool
	ForShare  bool
	KeyIn     []common.ParameterKey
	common.Page
}

type ParameterDao struct {
	db        infra.Database
	env       *lib.Env
	redis     infra.Redis
	beBuilder *lib.BEBuilder
}

func NewParameterDao(db infra.Database, env *lib.Env, redis infra.Redis, beBuilder *lib.BEBuilder) *ParameterDao {
	return &ParameterDao{
		db:        db,
		env:       env,
		redis:     redis,
		beBuilder: beBuilder,
	}
}

func (ad *ParameterDao) WithTx(tx *gorm.DB) *ParameterDao {
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Parameter) TableName() string {
	return "parameter"
}

func (ad *ParameterDao) ListByKeys(ctx context.Context, keys []common.ParameterKey) ([]*Parameter, error) {
	if len(keys) == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*Parameter, 0)
	db := ad.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, ad.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(Parameter{}).
			Scopes(ad.queryChain(&ParameterQuery{
				KeyIn: keys,
			})).
			Scan(&result)
	}, func(tx *gorm.DB) ([]*Parameter, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return []*Parameter{}, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&infra.L2CacheConfig{
			Level:      []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpiryTime: ad.env.L2CacheExpire,
		})
}

func (ad *ParameterDao) GetByKey(ctx context.Context, key common.ParameterKey) (*Parameter, error) {
	if key == "" {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Parameter{}
	db := ad.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, ad.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(Parameter{}).
			Scopes(ad.queryChain(&ParameterQuery{
				Parameter: Parameter{
					Key: key,
				},
			})).
			First(result)
	}, func(tx *gorm.DB) (*Parameter, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&infra.L2CacheConfig{
			Level:      []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpiryTime: ad.env.L2CacheExpire,
		})
}

func (ad *ParameterDao) Get(ctx context.Context, query *ParameterQuery) (*Parameter, error) {
	result := &Parameter{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Parameter{}).
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

func (ad *ParameterDao) Gets(ctx context.Context, query *ParameterQuery) ([]*Parameter, error) {
	result := make([]*Parameter, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Parameter{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Parameter{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *ParameterDao) Save(ctx context.Context, model *Parameter) (uint64, error) {

	db := ad.db.WithContext(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *ParameterDao) Update(ctx context.Context, query *ParameterQuery) (int64, error) {

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
			}
		default:
			continue
		}
	}

	ret := db.
		Model(Parameter{}).
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

func (pd *ParameterDao) queryChain(query *ParameterQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Parameter)
		structValue := reflect.ValueOf(query.Parameter)
		structPtrValue := reflect.ValueOf(&query.Parameter)
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
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(pd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(pd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.KeyIn != nil {
			db.Scopes(pd.inScope("key", query.KeyIn))
		}

		return db
	}
}

func (ad *ParameterDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *ParameterDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
