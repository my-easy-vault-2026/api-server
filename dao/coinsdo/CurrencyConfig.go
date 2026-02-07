package coinsdo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"shared-modules/common"
	"time"

	"api-server/infra"
	"api-server/lib"

	"github.com/gobeam/stringy"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type CurrencyConfig struct {
	ID               uint64
	Type             common.AssetType
	CurrencyType     common.Currency
	CurrencyName     string
	CurrencyFullName string
	CurrencyStatus   common.CurrencyConfigStatus
	Decimals         int       `gorm:"default:null"`
	CreatedAt        time.Time `gorm:"default:null"`
	UpdatedAt        time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type CurrencyConfigQuery struct {
	CurrencyConfig
	Attrs      CurrencyConfig
	ForUpdate  bool
	ForShare   bool
	CurrencyIn []common.Currency
	Select     []string
	Distinct   string
	GroupBy    string
	IsNull     []string
	IsNotNull  []string
	common.Page
}

type CurrencyConfigDao struct {
	db        infra.Database
	env       *lib.Env
	redis     infra.Redis
	beBuilder *lib.BEBuilder
}

func NewCurrencyConfigDao(db infra.Database, env *lib.Env, redis infra.Redis, beBuilder *lib.BEBuilder) *CurrencyConfigDao {
	return &CurrencyConfigDao{db: db, env: env, redis: redis, beBuilder: beBuilder}
}

func (cc *CurrencyConfigDao) WithTx(tx *gorm.DB) *CurrencyConfigDao {
	if cc == nil {
		return cc
	}
	newDao := *cc
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (CurrencyConfig) TableName() string {
	return "currency_config"
}

func (cc *CurrencyConfigDao) GetCryptoCurrencies(ctx context.Context) ([]*CurrencyConfig, error) {
	result := []*CurrencyConfig{}
	db := cc.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, cc.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CurrencyConfig{}).
			Scopes(cc.queryChain(&CurrencyConfigQuery{
				CurrencyConfig: CurrencyConfig{
					CurrencyStatus: common.CURRENCY_CONFIG_STATUS_ON,
				},
			})).Scan(&result)
	},
		func(tx *gorm.DB) ([]*CurrencyConfig, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: cc.env.L2CacheExpire,
		})
}

func (cc *CurrencyConfigDao) ListByType(ctx context.Context, t common.AssetType) ([]*CurrencyConfig, error) {
	if t == 0 {
		return nil, cc.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := []*CurrencyConfig{}
	db := cc.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, cc.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CurrencyConfig{}).
			Scopes(cc.queryChain(&CurrencyConfigQuery{
				CurrencyConfig: CurrencyConfig{
					Type:           t,
					CurrencyStatus: common.CURRENCY_CONFIG_STATUS_ON,
				},
			})).Scan(&result)
	},
		func(tx *gorm.DB) ([]*CurrencyConfig, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: cc.env.L2CacheExpire,
		})
}

func (cc *CurrencyConfigDao) ListByCurrencies(ctx context.Context, currencies []common.Currency) ([]*CurrencyConfig, error) {
	if len(currencies) == 0 {
		return nil, cc.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := []*CurrencyConfig{}
	db := cc.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, cc.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CurrencyConfig{}).
			Scopes(cc.queryChain(&CurrencyConfigQuery{
				CurrencyConfig: CurrencyConfig{},
				CurrencyIn:     currencies,
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*CurrencyConfig, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: cc.env.L2CacheExpire,
		})
}

func (cc *CurrencyConfigDao) ListDisplayDecimalsByDistCurrencies(ctx context.Context) ([]*CurrencyConfig, error) {

	result := []*CurrencyConfig{}
	db := cc.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, cc.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CurrencyConfig{}).
			Scopes(cc.queryChain(&CurrencyConfigQuery{
				Select:  []string{"currency_type", "MAX(`display_decimals`) AS `display_decimals`"},
				GroupBy: "currency_type",
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*CurrencyConfig, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &infra.L2CacheConfig{
			Level:         []infra.L2CacheLevel{infra.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: cc.env.L2CacheExpire,
		})
}

func (cc *CurrencyConfigDao) queryChain(query *CurrencyConfigQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.CurrencyConfig)
		structValue := reflect.ValueOf(query.CurrencyConfig)
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
				db.Scopes(cc.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(cc.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(cc.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(cc.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(cc.equalScope(fieldName, structValue.Field(i).Bool()))
			default:
				continue
			}
		}

		if query.CurrencyIn != nil {
			db.Scopes(cc.inScope("currency_type", query.CurrencyIn))
		}

		if query.Select != nil {
			db.Scopes(cc.selectScope(query.Select))
		}

		return db.Scopes(cc.distinctScope(query.Distinct)).
			Scopes(cc.groupByScope(query.GroupBy)).
			Scopes(cc.nullScope(query.IsNull, true)).
			Scopes(cc.nullScope(query.IsNotNull, false))
	}
}

func (cc *CurrencyConfigDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (cc *CurrencyConfigDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (cc *CurrencyConfigDao) selectScope(field []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(field) > 0 {
			return db.Select(field)
		}
		return db
	}
}

func (cc *CurrencyConfigDao) distinctScope(field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !reflect.ValueOf(field).IsZero() {
			return db.Distinct(field)
		}
		return db
	}
}

func (cc *CurrencyConfigDao) groupByScope(field string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if field != "" {
			return db.Group(field)
		}
		return db
	}
}

func (cc *CurrencyConfigDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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
