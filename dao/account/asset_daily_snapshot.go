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
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type AssetDailySnapshot struct {
	SnapshotDate    string
	ID              uint64
	CategoryID      uint64
	Type            common.AssetType
	UserID          uint64
	Currency        common.Currency
	CurrencyType    common.CurrencyType
	Amount          decimal.Decimal
	FreezedAmount   decimal.Decimal
	Signature       string
	DeletedAt       time.Time `gorm:"default:null"`
	SourceCreatedAt time.Time `gorm:"default:null"`
	SourceUpdatedAt time.Time `gorm:"default:null"`
	CreatedAt       time.Time `gorm:"default:null"`
	UpdatedAt       time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type AssetDailySnapshotQuery struct {
	AssetDailySnapshot
	Attrs          AssetDailySnapshot
	CategoryIDIn   []uint64
	IDIn           []uint64
	TypeIn         []common.AssetType
	CurrencyIn     []common.Currency
	UserIDIn       []uint64
	ForUpdate      bool
	ForShare       bool
	Deleted        bool
	OrderBy        string
	OrderDirection common.OrderDirection
	utils.Page
}

type AssetDailySnapshotDao struct {
}

func NewAssetDailySnapshotDao() *AssetDailySnapshotDao {
	return &AssetDailySnapshotDao{}
}

func (AssetDailySnapshot) TableName() string {
	return "asset_daily_snapshot"
}

func (ad *AssetDailySnapshotDao) Get(ctx context.Context, query *AssetDailySnapshotQuery) (*AssetDailySnapshot, error) {
	result := &AssetDailySnapshot{}
	db := utils.GetDB(ctx)

	err := db.
		Model(AssetDailySnapshot{}).
		Scopes(ad.queryChain(query)).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *AssetDailySnapshotDao) Gets(ctx context.Context, query *AssetDailySnapshotQuery) ([]*AssetDailySnapshot, error) {
	result := make([]*AssetDailySnapshot, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(AssetDailySnapshot{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*AssetDailySnapshot{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *AssetDailySnapshotDao) Save(ctx context.Context, model *AssetDailySnapshot, clauses ...utils.Clause) (uint64, error) {

	db := utils.GetDB(ctx)

	for _, c := range clauses {
		db = db.Scopes(c.Scope())
	}

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *AssetDailySnapshotDao) Saves(ctx context.Context, models []*AssetDailySnapshot) (int64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(AssetDailySnapshot{}).
		CreateInBatches(models, len(models))

	if ret.Error != nil {
		return ret.RowsAffected, ret.Error
	}
	return ret.RowsAffected, nil
}

func (ad *AssetDailySnapshotDao) SoftDeleteByID(ctx context.Context, id uint64) (int64, error) {

	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	return ad.Update(ctx, &AssetDailySnapshotQuery{
		AssetDailySnapshot: AssetDailySnapshot{
			ID: id,
		},
		Attrs: AssetDailySnapshot{
			DeletedAt: utils.DBQueryTime(time.Now()),
		},
	})
}

func (ad *AssetDailySnapshotDao) DeleteByID(ctx context.Context, id uint64) (int64, error) {
	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	db := utils.GetDB(ctx)

	ret := db.
		Delete(&AssetDailySnapshot{
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

func (ad *AssetDailySnapshotDao) Update(ctx context.Context, query *AssetDailySnapshotQuery) (int64, error) {

	if query.ID == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
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
		Model(AssetDailySnapshot{}).
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

func (ad *AssetDailySnapshotDao) queryChain(query *AssetDailySnapshotQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.AssetDailySnapshot)
		structValue := reflect.ValueOf(query.AssetDailySnapshot)
		structPtrValue := reflect.ValueOf(&query.AssetDailySnapshot)
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

		if !query.Deleted {
			db.Scopes(ad.nullScope([]string{"deleted_at"}, true))
		}

		if query.TypeIn != nil {
			db.Scopes(ad.inScope("type", query.TypeIn))
		}

		if query.CurrencyIn != nil {
			db.Scopes(ad.inScope("currency", query.CurrencyIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(ad.inScope("user_id", query.UserIDIn))
		}

		return db.
			Scopes(ad.categoryIDInScope(query.CategoryIDIn)).
			Scopes(ad.IDInScope(query.IDIn)).
			Scopes(ad.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(ad.pageScope(query.Current, query.PageSize))
	}
}

func (ad *AssetDailySnapshotDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *AssetDailySnapshotDao) categoryIDInScope(categories []uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(categories) == 0 {
			return db
		}
		return db.Where("category_id IN ? ", categories)
	}
}

func (ad *AssetDailySnapshotDao) IDInScope(ids []uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(ids) == 0 {
			return db
		}
		return db.Where("id IN ? ", ids)
	}
}

func (ad *AssetDailySnapshotDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (ad *AssetDailySnapshotDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (ad *AssetDailySnapshotDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *AssetDailySnapshotDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (cd *AssetDailySnapshotDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (cd *AssetDailySnapshotDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
