package coinsdo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"api-server/infra"
	"api-server/lib"

	"github.com/gobeam/stringy"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type CryptoCurrency struct {
	ID               uint64
	Type             common.AssetType
	NationCode       string          `gorm:"default:null"`
	Mainnet          common.Mainnet  `gorm:"default:null"`
	MainnetName      string          `gorm:"default:null"`
	MainnetFullName  string          `gorm:"default:null"`
	Protocol         common.Protocol `gorm:"default:null"`
	ProtocolName     string          `gorm:"default:null"`
	CurrencyType     common.Currency
	CurrencyName     string
	CurrencyFullName string
	Flag             string `gorm:"default:null"`
	CurrencyStatus   common.CryptoCurrencyStatus
	Decimals         int                  `gorm:"default:null"`
	DisplayDecimals  int                  `gorm:"default:null"`
	InterestDecimals int                  `gorm:"default:null"`
	CoinsdoID        uint64               `gorm:"default:null"`
	CaseSensitive    common.CaseSensitive `gorm:"default:null"`
	CoinType         common.CoinTokenType `gorm:"default:null"`
	ExplorerURL      string               `gorm:"default:null"`
	CreatedAt        time.Time            `gorm:"default:null"`
	UpdatedAt        time.Time            `gorm:"default:null;autoUpdateTime:false"`
}

type CryptoCurrencyQuery struct {
	CryptoCurrency
	Attrs      CryptoCurrency
	ForUpdate  bool
	ForShare   bool
	CurrencyIn []common.Currency
	Select     []string
	Distinct   string
	GroupBy    string
	IsNull     []string
	IsNotNull  []string
	utils.Page
}

type CryptoCurrencyDao struct {
	db  infra.Database
	env *lib.Env
}

func NewCryptoCurrencyDao(db infra.Database, env *lib.Env) *CryptoCurrencyDao {
	return &CryptoCurrencyDao{db: db, env: env}
}

func (cc *CryptoCurrencyDao) WithTx(tx *gorm.DB) *CryptoCurrencyDao {
	if cc == nil {
		return cc
	}
	newDao := *cc
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (CryptoCurrency) TableName() string {
	return "crypto_currency"
}

func (cc *CryptoCurrencyDao) GetCryptoCurrency(ctx context.Context, mainnet common.Mainnet, currency string) (*CryptoCurrency, error) {
	result := &CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Mainnet:        mainnet,
					CurrencyName:   currency,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetCryptoTypeCurrencies(ctx context.Context) ([]*CryptoCurrency, error) {
	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Type:           common.ASSET_TYPE_CRYPTO,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetCryptoCurrencies(ctx context.Context) ([]*CryptoCurrency, error) {
	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) ListByType(ctx context.Context, t common.AssetType) ([]*CryptoCurrency, error) {
	if t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Type:           t,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetCryptoCurrencyByCurrencyType(ctx context.Context, currency common.Currency) (*CryptoCurrency, error) {
	result := &CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					CurrencyType: currency,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetByMainnetProtocolCurrencyType(ctx context.Context, mainnet common.Mainnet, protocol common.Protocol, currency common.Currency) (*CryptoCurrency, error) {

	if mainnet == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Mainnet:        mainnet,
					Protocol:       protocol,
					CurrencyType:   currency,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetByMainnetCurrency(ctx context.Context, mainnet common.Mainnet, currency common.Currency) (*CryptoCurrency, error) {

	if mainnet == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Mainnet:        mainnet,
					CurrencyType:   currency,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) GetByMainnet(ctx context.Context, mainnet common.Mainnet) (*CryptoCurrency, error) {

	if mainnet == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{
					Mainnet:        mainnet,
					CurrencyStatus: common.CRYPTO_CURRENCY_STATUS_ON,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) ListByCurrencies(ctx context.Context, currencies []common.Currency) ([]*CryptoCurrency, error) {
	if len(currencies) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				CryptoCurrency: CryptoCurrency{},
				CurrencyIn:     currencies,
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) ListDisplayDecimalsByDistCurrencies(ctx context.Context) ([]*CryptoCurrency, error) {

	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				Select:  []string{"currency_type", "MAX(`display_decimals`) AS `display_decimals`"},
				GroupBy: "currency_type",
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cc *CryptoCurrencyDao) ListMainnetNames(ctx context.Context) ([]*CryptoCurrency, error) {

	result := []*CryptoCurrency{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&CryptoCurrency{}).
			Scopes(cc.queryChain(&CryptoCurrencyQuery{
				Select:  []string{"mainnet", "MAX(`mainnet_full_name`) AS `mainnet_full_name`"},
				GroupBy: "mainnet",
			})).
			Scan(&result)
	},
		func(tx *gorm.DB) ([]*CryptoCurrency, error) {
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

func (cd *CryptoCurrencyDao) queryChain(query *CryptoCurrencyQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.CryptoCurrency)
		structValue := reflect.ValueOf(query.CryptoCurrency)
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
			default:
				continue
			}
		}

		if query.CurrencyIn != nil {
			db.Scopes(cd.inScope("currency_type", query.CurrencyIn))
		}

		if query.Select != nil {
			db.Scopes(cd.selectScope(query.Select))
		}

		return db.Scopes(cd.distinctScope(query.Distinct)).
			Scopes(cd.groupByScope(query.GroupBy)).
			Scopes(cd.nullScope(query.IsNull, true)).
			Scopes(cd.nullScope(query.IsNotNull, false))
	}
}

func (cd *CryptoCurrencyDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (cd *CryptoCurrencyDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (cd *CryptoCurrencyDao) selectScope(field []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(field) > 0 {
			return db.Select(field)
		}
		return db
	}
}

func (cd *CryptoCurrencyDao) distinctScope(field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !reflect.ValueOf(field).IsZero() {
			return db.Distinct(field)
		}
		return db
	}
}

func (cd *CryptoCurrencyDao) groupByScope(field string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if field != "" {
			return db.Group(field)
		}
		return db
	}
}

func (cd *CryptoCurrencyDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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
