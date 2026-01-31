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

	"api-server/infra"
	"api-server/lib"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type AssetSnapshot struct {
	ID             uint64                       `gorm:"default:null"`
	UserID         uint64                       `gorm:"default:null"`
	UserRole       common.Role                  `gorm:"default:null"`
	FinancialCode  common.FinancialCode         `gorm:"default:null"`
	ParentOrderNO  string                       `gorm:"default:null"`
	OrderNO        string                       `gorm:"default:null"`
	CardType       common.AssetType             `gorm:"default:null"`
	CardID         uint64                       `gorm:"default:null"`
	CardCategoryID uint64                       `gorm:"default:null"`
	CardCurrency   common.Currency              `gorm:"default:null"`
	Balance        *decimal.Decimal             `gorm:"default:null"`
	Interest       *decimal.Decimal             `gorm:"default:null"`
	Missing        common.SnapshotMissingStatus `gorm:"default:null"`
	EarningStatus  common.CardEarningStatus     `gorm:"default:null"`
	TakenAt        time.Time                    `gorm:"default:null"`
	CreatedAt      time.Time                    `gorm:"default:null"`
	UpdatedAt      time.Time                    `gorm:"default:null;autoUpdateTime:false"`
}

type AssetSnapshotQuery struct {
	AssetSnapshot
	Attrs     AssetSnapshot
	ForUpdate bool
	ForShare  bool
	utils.Page
}

type AssetSnapshotDao struct {
	db    infra.Database
	env   *lib.Env
	redis infra.Redis
}

func NewAssetSnapshotDao(db infra.Database, env *lib.Env, redis infra.Redis) *AssetSnapshotDao {
	return &AssetSnapshotDao{db: db, env: env, redis: redis}
}

func (ad *AssetSnapshotDao) WithTx(tx *gorm.DB) *AssetSnapshotDao {
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (AssetSnapshot) TableName() string {
	return "asset_snapshot"
}

func (ad *AssetSnapshotDao) GetByCode(ctx context.Context, code common.FinancialCode) (*AssetSnapshot, error) {
	result := &AssetSnapshot{}
	db := ad.db.WithContext(ctx)

	return infra.L2CQuery(ctx, db, ad.redis, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(&AssetSnapshot{}).
			Scopes(ad.queryChain(&AssetSnapshotQuery{
				AssetSnapshot: AssetSnapshot{
					FinancialCode: code,
				},
			})).First(result)
	},
		func(tx *gorm.DB) (*AssetSnapshot, error) {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			if tx.Error != nil {
				return nil, tx.Error
			}
			return result, nil
		}, &infra.L2CacheConfig{
			Level:         []common.L2CacheLevel{common.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: ad.env.L2CacheExpire,
		})
}

func (ad *AssetSnapshotDao) List(ctx context.Context) ([]*AssetSnapshot, error) {

	result := make([]*AssetSnapshot, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(AssetSnapshot{}).
		Scopes(ad.queryChain(&AssetSnapshotQuery{})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*AssetSnapshot{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetSnapshotDao) Get(ctx context.Context, query *AssetSnapshotQuery) (*AssetSnapshot, error) {
	result := &AssetSnapshot{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(AssetSnapshot{}).
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

func (ad *AssetSnapshotDao) Gets(ctx context.Context, query *AssetSnapshotQuery) ([]*AssetSnapshot, error) {
	result := make([]*AssetSnapshot, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(AssetSnapshot{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*AssetSnapshot{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetSnapshotDao) Save(ctx context.Context, model *AssetSnapshot) (uint64, error) {

	db := ad.db.WithContext(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *AssetSnapshotDao) Saves(ctx context.Context, models []*AssetSnapshot) (int64, error) {

	db := ad.db.WithContext(ctx)

	ret := db.
		Model(AssetSnapshot{}).
		CreateInBatches(models, len(models))

	if ret.Error != nil {
		return ret.RowsAffected, ret.Error
	}
	return ret.RowsAffected, nil
}

func (ad *AssetSnapshotDao) Update(ctx context.Context, query *AssetSnapshotQuery) (int64, error) {

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
		Model(AssetSnapshot{}).
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

func (ad *AssetSnapshotDao) DeleteByID(ctx context.Context, id uint64) (int64, error) {
	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	db := ad.db.WithContext(ctx)

	ret := db.
		Delete(&AssetSnapshot{
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

func (ad *AssetSnapshotDao) queryChain(query *AssetSnapshotQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.AssetSnapshot)
		structValue := reflect.ValueOf(query.AssetSnapshot)
		structPtrValue := reflect.ValueOf(&query.AssetSnapshot)
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

		if query.ForUpdate {
			db.Scopes(ad.forScope("UPDATE"))
		}

		if query.ForShare {
			db.Scopes(ad.forScope("SHARE"))
		}

		return db
	}
}

func (ad *AssetSnapshotDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *AssetSnapshotDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
	}
}

func (ad *AssetSnapshotDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (ad *AssetSnapshotDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *AssetSnapshotDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *AssetSnapshotDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (ad *AssetSnapshotDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
