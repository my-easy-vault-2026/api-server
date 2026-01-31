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

type CurrencyConfig struct {
	ID                                   uint64
	Type                                 common.CurrencyType
	NationCode                           common.NationCode `gorm:"default:null"`
	Mainnet                              common.Mainnet    `gorm:"default:null"`
	Protocol                             common.Protocol   `gorm:"default:null"`
	Currency                             common.Currency
	KycLevel                             common.KYCLevel              `gorm:"default:null"`
	MerchantID                           uint64                       `gorm:"default:null"`
	Apply                                common.CurrencyConfigStatus  `gorm:"default:null"`
	ApplyMin                             *decimal.NullDecimal         `gorm:"default:null"`
	ApplyMax                             *decimal.NullDecimal         `gorm:"default:null"`
	ApplyMaxMonthly                      *decimal.NullDecimal         `gorm:"default:null"`
	ApplyLimitCurrency                   common.Currency              `gorm:"default:null"`
	ApplyCountDaily                      *int                         `gorm:"default:null"`
	ApplyCountMonthly                    *int                         `gorm:"default:null"`
	ApplyFee                             *decimal.NullDecimal         `gorm:"default:null"`
	ApplyFeeType                         common.CurrencyConfigFeeType `gorm:"default:null"`
	ApplyFeeCurrency                     common.Currency              `gorm:"default:null"`
	Transfer                             common.CurrencyConfigStatus  `gorm:"default:null"`
	TransferMin                          *decimal.NullDecimal         `gorm:"default:null"`
	TransferMax                          *decimal.NullDecimal         `gorm:"default:null"`
	TransferMaxMonthly                   *decimal.NullDecimal         `gorm:"default:null"`
	TransferLimitCurrency                common.Currency              `gorm:"default:null"`
	TransferCountDaily                   *int                         `gorm:"default:null"`
	TransferCountMonthly                 *int                         `gorm:"default:null"`
	TransferFee                          *decimal.NullDecimal         `gorm:"default:null"`
	TransferFeeType                      common.CurrencyConfigFeeType `gorm:"default:null"`
	TransferFeeCurrency                  common.Currency              `gorm:"default:null"`
	Exchange                             common.CurrencyConfigStatus  `gorm:"default:null"`
	ExchangeMin                          *decimal.NullDecimal         `gorm:"default:null"`
	ExchangeMax                          *decimal.NullDecimal         `gorm:"default:null"`
	ExchangeMaxMonthly                   *decimal.NullDecimal         `gorm:"default:null"`
	ExchangeLimitCurrency                common.Currency              `gorm:"default:null"`
	ExchangeCountDaily                   *int                         `gorm:"default:null"`
	ExchangeCountMonthly                 *int                         `gorm:"default:null"`
	ExchangeFee                          *decimal.NullDecimal         `gorm:"default:null"`
	ExchangeFeeType                      common.CurrencyConfigFeeType `gorm:"default:null"`
	ExchangeFeeCurrency                  common.Currency              `gorm:"default:null"`
	Withdraw                             common.CurrencyConfigStatus  `gorm:"default:null"`
	WithdrawMin                          *decimal.NullDecimal         `gorm:"default:null"`
	WithdrawMax                          *decimal.NullDecimal         `gorm:"default:null"`
	WithdrawMaxMonthly                   *decimal.NullDecimal         `gorm:"default:null"`
	WithdrawLimitCurrency                common.Currency              `gorm:"default:null"`
	WithdrawCountDaily                   *int                         `gorm:"default:null"`
	WithdrawCountMonthly                 *int                         `gorm:"default:null"`
	WithdrawFee                          *decimal.NullDecimal         `gorm:"default:null"`
	WithdrawFeeType                      common.CurrencyConfigFeeType `gorm:"default:null"`
	WithdrawFeeCurrency                  common.Currency              `gorm:"default:null"`
	Pay                                  common.CurrencyConfigStatus  `gorm:"default:null"`
	PayMin                               *decimal.NullDecimal         `gorm:"default:null"`
	PayMax                               *decimal.NullDecimal         `gorm:"default:null"`
	PayMaxMonthly                        *decimal.NullDecimal         `gorm:"default:null"`
	PayLimitCurrency                     common.Currency              `gorm:"default:null"`
	PayCountDaily                        *int                         `gorm:"default:null"`
	PayCountMonthly                      *int                         `gorm:"default:null"`
	PayFee                               *decimal.NullDecimal         `gorm:"default:null"`
	PayFeeType                           common.CurrencyConfigFeeType `gorm:"default:null"`
	PayFeeCurrency                       common.Currency              `gorm:"default:null"`
	Deposit                              common.CurrencyConfigStatus  `gorm:"default:null"`
	DepositMin                           *decimal.NullDecimal         `gorm:"default:null"`
	DepositMax                           *decimal.NullDecimal         `gorm:"default:null"`
	DepositMaxMonthly                    *decimal.NullDecimal         `gorm:"default:null"`
	DepositLimitCurrency                 common.Currency              `gorm:"default:null"`
	DepositCountDaily                    *int                         `gorm:"default:null"`
	DepositCountMonthly                  *int                         `gorm:"default:null"`
	DepositFee                           *decimal.NullDecimal         `gorm:"default:null"`
	DepositFeeType                       common.CurrencyConfigFeeType `gorm:"default:null"`
	DepositFeeCurrency                   common.Currency              `gorm:"default:null"`
	TopUp                                common.CurrencyConfigStatus  `gorm:"default:null"`
	TopUpMin                             *decimal.NullDecimal         `gorm:"default:null"`
	TopUpMax                             *decimal.NullDecimal         `gorm:"default:null"`
	TopUpMaxMonthly                      *decimal.NullDecimal         `gorm:"default:null"`
	TopUpLimitCurrency                   common.Currency              `gorm:"default:null"`
	TopUpCountDaily                      *int                         `gorm:"default:null"`
	TopUpCountMonthly                    *int                         `gorm:"default:null"`
	TopUpFee                             *decimal.NullDecimal         `gorm:"default:null"`
	TopUpFeeType                         common.CurrencyConfigFeeType `gorm:"default:null"`
	TopUpFeeCurrency                     common.Currency              `gorm:"default:null"`
	TopDown                              common.CurrencyConfigStatus  `gorm:"default:null"`
	TopDownMin                           *decimal.NullDecimal         `gorm:"default:null"`
	TopDownMax                           *decimal.NullDecimal         `gorm:"default:null"`
	TopDownMaxMonthly                    *decimal.NullDecimal         `gorm:"default:null"`
	TopDownLimitCurrency                 common.Currency              `gorm:"default:null"`
	TopDownCountDaily                    *int                         `gorm:"default:null"`
	TopDownCountMonthly                  *int                         `gorm:"default:null"`
	TopDownFee                           *decimal.NullDecimal         `gorm:"default:null"`
	TopDownFeeType                       common.CurrencyConfigFeeType `gorm:"default:null"`
	TopDownFeeCurrency                   common.Currency              `gorm:"default:null"`
	CardToCard                           common.CurrencyConfigStatus  `gorm:"default:null"`
	CardToCardMin                        *decimal.NullDecimal         `gorm:"default:null"`
	CardToCardMax                        *decimal.NullDecimal         `gorm:"default:null"`
	CardToCardMaxMonthly                 *decimal.NullDecimal         `gorm:"default:null"`
	CardToCardLimitCurrency              common.Currency              `gorm:"default:null"`
	CardToCardCountDaily                 *int                         `gorm:"default:null"`
	CardToCardCountMonthly               *int                         `gorm:"default:null"`
	CardToCardFee                        *decimal.NullDecimal         `gorm:"default:null"`
	CardToCardFeeType                    common.CurrencyConfigFeeType `gorm:"default:null"`
	CardToCardFeeCurrency                common.Currency              `gorm:"default:null"`
	SelfCardToCard                       common.CurrencyConfigStatus  `gorm:"default:null"`
	SelfCardToCardMin                    *decimal.NullDecimal         `gorm:"default:null"`
	SelfCardToCardMax                    *decimal.NullDecimal         `gorm:"default:null"`
	SelfCardToCardMaxMonthly             *decimal.NullDecimal         `gorm:"default:null"`
	SelfCardToCardLimitCurrency          common.Currency              `gorm:"default:null"`
	SelfCardToCardCountDaily             *int                         `gorm:"default:null"`
	SelfCardToCardCountMonthly           *int                         `gorm:"default:null"`
	SelfCardToCardFee                    *decimal.NullDecimal         `gorm:"default:null"`
	SelfCardToCardFeeType                common.CurrencyConfigFeeType `gorm:"default:null"`
	SelfCardToCardFeeCurrency            common.Currency              `gorm:"default:null"`
	WhaleCardToCard                      common.CurrencyConfigStatus  `gorm:"default:null"`
	WhaleCardToCardMin                   *decimal.NullDecimal         `gorm:"default:null"`
	WhaleCardToCardMax                   *decimal.NullDecimal         `gorm:"default:null"`
	WhaleCardToCardMaxMonthly            *decimal.NullDecimal         `gorm:"default:null"`
	WhaleCardToCardLimitCurrency         common.Currency              `gorm:"default:null"`
	WhaleCardToCardCountDaily            *int                         `gorm:"default:null"`
	WhaleCardToCardCountMonthly          *int                         `gorm:"default:null"`
	WhaleCardToCardFee                   *decimal.NullDecimal         `gorm:"default:null"`
	WhaleCardToCardFeeType               common.CurrencyConfigFeeType `gorm:"default:null"`
	WhaleCardToCardFeeCurrency           common.Currency              `gorm:"default:null"`
	WhaleSelfCardToCard                  common.CurrencyConfigStatus  `gorm:"default:null"`
	WhaleSelfCardToCardMin               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleSelfCardToCardMax               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleSelfCardToCardMaxMonthly        *decimal.NullDecimal         `gorm:"default:null"`
	WhaleSelfCardToCardLimitCurrency     common.Currency              `gorm:"default:null"`
	WhaleSelfCardToCardCountDaily        *int                         `gorm:"default:null"`
	WhaleSelfCardToCardCountMonthly      *int                         `gorm:"default:null"`
	WhaleSelfCardToCardFee               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleSelfCardToCardFeeType           common.CurrencyConfigFeeType `gorm:"default:null"`
	WhaleSelfCardToCardFeeCurrency       common.Currency              `gorm:"default:null"`
	WhaleReapCardToCard                  common.CurrencyConfigStatus  `gorm:"default:null"`
	WhaleReapCardToCardMin               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleReapCardToCardMax               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleReapCardToCardMaxMonthly        *decimal.NullDecimal         `gorm:"default:null"`
	WhaleReapCardToCardLimitCurrency     common.Currency              `gorm:"default:null"`
	WhaleReapCardToCardCountDaily        *int                         `gorm:"default:null"`
	WhaleReapCardToCardCountMonthly      *int                         `gorm:"default:null"`
	WhaleReapCardToCardFee               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleReapCardToCardFeeType           common.CurrencyConfigFeeType `gorm:"default:null"`
	WhaleReapCardToCardFeeCurrency       common.Currency              `gorm:"default:null"`
	WhaleTopUp                           common.CurrencyConfigStatus  `gorm:"default:null"`
	WhaleTopUpMin                        *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopUpMax                        *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopUpMaxMonthly                 *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopUpLimitCurrency              common.Currency              `gorm:"default:null"`
	WhaleTopUpCountDaily                 *int                         `gorm:"default:null"`
	WhaleTopUpCountMonthly               *int                         `gorm:"default:null"`
	WhaleTopUpFee                        *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopUpFeeType                    common.CurrencyConfigFeeType `gorm:"default:null"`
	WhaleTopUpFeeCurrency                common.Currency              `gorm:"default:null"`
	WhaleTopDown                         common.CurrencyConfigStatus  `gorm:"default:null"`
	WhaleTopDownMin                      *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopDownMax                      *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopDownMaxMonthly               *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopDownLimitCurrency            common.Currency              `gorm:"default:null"`
	WhaleTopDownCountDaily               *int                         `gorm:"default:null"`
	WhaleTopDownCountMonthly             *int                         `gorm:"default:null"`
	WhaleTopDownFee                      *decimal.NullDecimal         `gorm:"default:null"`
	WhaleTopDownFeeType                  common.CurrencyConfigFeeType `gorm:"default:null"`
	WhaleTopDownFeeCurrency              common.Currency              `gorm:"default:null"`
	PaycryptoTopUp                       common.CurrencyConfigStatus  `gorm:"default:null"`
	PaycryptoTopUpMin                    *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoTopUpMax                    *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoTopUpMaxMonthly             *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoTopUpLimitCurrency          common.Currency              `gorm:"default:null"`
	PaycryptoTopUpCountDaily             *int                         `gorm:"default:null"`
	PaycryptoTopUpCountMonthly           *int                         `gorm:"default:null"`
	PaycryptoTopUpFee                    *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoTopUpFeeType                common.CurrencyConfigFeeType `gorm:"default:null"`
	PaycryptoTopUpFeeCurrency            common.Currency              `gorm:"default:null"`
	PaycryptoReapCardToCard              common.CurrencyConfigStatus  `gorm:"default:null"`
	PaycryptoReapCardToCardMin           *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoReapCardToCardMax           *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoReapCardToCardMaxMonthly    *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoReapCardToCardLimitCurrency common.Currency              `gorm:"default:null"`
	PaycryptoReapCardToCardCountDaily    *int                         `gorm:"default:null"`
	PaycryptoReapCardToCardCountMonthly  *int                         `gorm:"default:null"`
	PaycryptoReapCardToCardFee           *decimal.NullDecimal         `gorm:"default:null"`
	PaycryptoReapCardToCardFeeType       common.CurrencyConfigFeeType `gorm:"default:null"`
	PaycryptoReapCardToCardFeeCurrency   common.Currency              `gorm:"default:null"`
	CreatedAt                            time.Time                    `gorm:"default:null"`
	UpdatedAt                            time.Time                    `gorm:"default:null;autoUpdateTime:false"`
}

type CurrencyConfigQuery struct {
	CurrencyConfig
	Attrs     CurrencyConfig
	ForUpdate bool
	ForShare  bool
	IsNull    []string
	utils.Page
}

type CurrencyConfigDao struct {
	db  infra.Database
	env *lib.Env
}

func NewCurrencyConfigDao(db infra.Database, env *lib.Env) *CurrencyConfigDao {
	return &CurrencyConfigDao{db: db, env: env}
}

func (ad *CurrencyConfigDao) WithTx(tx *gorm.DB) *CurrencyConfigDao {
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (CurrencyConfig) TableName() string {
	return "currency_config"
}

func (ad *CurrencyConfigDao) ListByCurrency(ctx context.Context, currency common.Currency) ([]*CurrencyConfig, error) {
	if currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*CurrencyConfig, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(CurrencyConfig{}).
		Scopes(ad.queryChain(&CurrencyConfigQuery{
			CurrencyConfig: CurrencyConfig{
				Currency: currency,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*CurrencyConfig{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CurrencyConfigDao) List(ctx context.Context) ([]*CurrencyConfig, error) {
	result := make([]*CurrencyConfig, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(CurrencyConfig{}).
		Scopes(ad.queryChain(&CurrencyConfigQuery{})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*CurrencyConfig{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CurrencyConfigDao) Get(ctx context.Context, query *CurrencyConfigQuery) (*CurrencyConfig, error) {
	result := &CurrencyConfig{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(CurrencyConfig{}).
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

func (ad *CurrencyConfigDao) Gets(ctx context.Context, query *CurrencyConfigQuery) ([]*CurrencyConfig, error) {
	result := make([]*CurrencyConfig, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(CurrencyConfig{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*CurrencyConfig{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CurrencyConfigDao) Save(ctx context.Context, model *CurrencyConfig) (uint64, error) {

	db := ad.db.WithContext(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *CurrencyConfigDao) Update(ctx context.Context, query *CurrencyConfigQuery) (int64, error) {

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
		Model(CurrencyConfig{}).
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

func (pd *CurrencyConfigDao) queryChain(query *CurrencyConfigQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.CurrencyConfig)
		structValue := reflect.ValueOf(query.CurrencyConfig)
		structPtrValue := reflect.ValueOf(&query.CurrencyConfig)
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

		if len(query.IsNull) > 0 {
			db.Scopes(pd.nullScope(query.IsNull, true))
		}

		return db
	}
}

func (ad *CurrencyConfigDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *CurrencyConfigDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (ad *CurrencyConfigDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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
