package card

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"

	"shared-modules/common"
	"shared-modules/utils"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MainCard struct {
	ID                       uint64
	UserID                   uint64
	Currency                 common.Currency `gorm:"default:null"`
	CategoryID               uint64          `gorm:"default:null"`
	CardID                   uint64
	UserIDCurrencyCategoryID string    `gorm:"default:null"`
	CreatedAt                time.Time `gorm:"default:null"`
	UpdatedAt                time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type MainCardQuery struct {
	MainCard
	Attrs     MainCard
	IsNull    []string
	ForUpdate bool
	ForShare  bool
	utils.Page
}

type MainCardDao struct {
}

func NewMainCardDao() *MainCardDao {
	return &MainCardDao{}
}

func (MainCard) TableName() string {
	return "main_card"
}

func (md *MainCardDao) GetByUserIDCategoryID(ctx context.Context, userID uint64, categoryID uint64) (*MainCard, error) {
	if userID == 0 || categoryID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &MainCard{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(&MainCardQuery{
			MainCard: MainCard{
				UserID:     userID,
				CategoryID: categoryID,
			},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (md *MainCardDao) GetByUserIDCurrency(ctx context.Context, userID uint64, currency common.Currency) (*MainCard, error) {
	if userID == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &MainCard{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(&MainCardQuery{
			MainCard: MainCard{
				UserID:   userID,
				Currency: currency,
			},
			IsNull: []string{"category_id"},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (md *MainCardDao) GetByUserID(ctx context.Context, userID uint64) (*MainCard, error) {
	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &MainCard{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(&MainCardQuery{
			MainCard: MainCard{
				UserID: userID,
			},
			IsNull: []string{"category_id", "currency"},
		})).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (md *MainCardDao) Get(ctx context.Context, query *MainCardQuery) (*MainCard, error) {
	result := &MainCard{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (md *MainCardDao) ListByUserID(ctx context.Context, userID uint64) ([]*MainCard, error) {
	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*MainCard, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(&MainCardQuery{
			MainCard: MainCard{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*MainCard, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (md *MainCardDao) ListByCardID(ctx context.Context, cardID uint64) ([]*MainCard, error) {
	if cardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*MainCard, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(&MainCardQuery{
			MainCard: MainCard{
				CardID: cardID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*MainCard, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (md *MainCardDao) Gets(ctx context.Context, query *MainCardQuery) ([]MainCard, error) {
	result := make([]MainCard, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&MainCard{}).
		Scopes(md.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]MainCard, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (md *MainCardDao) Save(ctx context.Context, mainCard *MainCard) (uint64, error) {

	db := utils.GetDB(ctx)
	mainCard.UserIDCurrencyCategoryID = fmt.Sprintf("%d_%d_%d", mainCard.UserID, mainCard.Currency, mainCard.CategoryID)
	ret := db.Create(mainCard)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return mainCard.ID, nil
}

func (md *MainCardDao) DeleteByID(ctx context.Context, id uint64) (int64, error) {
	if id == 0 {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	db := utils.GetDB(ctx)

	ret := db.
		Delete(&MainCard{
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

func (md *MainCardDao) Update(ctx context.Context, query *MainCardQuery) (int64, error) {

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
		Model(MainCard{}).
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

func (md *MainCardDao) queryChain(query *MainCardQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.MainCard)
		structValue := reflect.ValueOf(query.MainCard)
		structPtrValue := reflect.ValueOf(&query.MainCard)
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
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(md.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(md.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.IsNull != nil {
			db.Scopes(md.nullScope(query.IsNull, true))
		}

		return db
	}
}

func (md *MainCardDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (md *MainCardDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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
