package card

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

type Card struct {
	ID           uint64              `gorm:"column:id"`
	Name         string              `gorm:"default:null"`
	UserID       uint64              `gorm:"default:null"`
	Type         common.AssetType    `gorm:"default:null"`
	CategoryID   uint64              `gorm:"default:null"`
	Currency     common.Currency     `gorm:"default:null"`
	CurrencyType common.CurrencyType `gorm:"default:null"`
	NationCode   common.NationCode   `gorm:"default:null"`
	Nation       string              `gorm:"default:null"`
	Status       common.CardStatus
	FrozenAt     time.Time `gorm:"default:null"`
	BlockedAt    time.Time `gorm:"default:null"`
	DeletedAt    time.Time `gorm:"default:null"`
	CreatedAt    time.Time `gorm:"default:null"`
	UpdatedAt    time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type CardQuery struct {
	Card
	Attrs             Card
	ForUpdate         bool
	ForShare          bool
	TypeIn            []common.AssetType
	IDIn              []uint64
	UserIDIn          []uint64
	IssueIdIn         []string
	AssetTypeIn       []common.AssetType
	OrderBy           string
	OrderDirection    common.OrderDirection
	BlockedAtLessThan time.Time
	Deleted           bool
	NotEqual          Card
	common.Page
}

type CardDao struct {
	db        infra.Database
	env       *lib.Env
	beBuilder *lib.BEBuilder
	logger    lib.Logger
}

func NewCardDao(db infra.Database,
	env *lib.Env,
	beBuilder *lib.BEBuilder,
	logger lib.Logger,
) *CardDao {
	return &CardDao{
		db:        db,
		env:       env,
		beBuilder: beBuilder,
		logger:    logger,
	}
}

func (ad *CardDao) WithTx(tx *gorm.DB) *CardDao {
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Card) TableName() string {
	return "card"
}

func (ad *CardDao) GetByUserIDCategoryIDForUpdate(ctx context.Context, userID uint64, categoryID uint64) (*Card, error) {

	if userID == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				CategoryID: categoryID,
			},
			ForUpdate: true,
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

func (ad *CardDao) GetByUserIDCurrencyType(ctx context.Context, userID uint64, currency common.Currency, t common.AssetType) (*Card, error) {

	if userID == 0 || currency == 0 || t == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:   userID,
				Currency: currency,
				Type:     t,
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

func (ad *CardDao) GetByUserIDCategoryID(ctx context.Context, userID uint64, categoryID uint64) (*Card, error) {

	if userID == 0 || categoryID == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
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

func (ad *CardDao) GetByUserIDInCategoryID(ctx context.Context, userIDs []uint64, categoryID uint64) (*Card, error) {

	if len(userIDs) == 0 || categoryID == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				CategoryID: categoryID,
			},
			UserIDIn: userIDs,
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

func (ad *CardDao) ListByUserID(ctx context.Context, userID uint64) ([]*Card, error) {

	if userID == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) GetByIDUserID(ctx context.Context, id uint64, userID uint64) (*Card, error) {
	if id == 0 || userID == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:     id,
				UserID: userID,
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

func (ad *CardDao) GetByIDUserIDIn(ctx context.Context, id uint64, userIDs []uint64) (*Card, error) {
	if id == 0 || len(userIDs) == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: id,
			},
			UserIDIn: userIDs,
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

func (ad *CardDao) GetByID(ctx context.Context, id uint64) (*Card, error) {
	if id == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: id,
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

func (ad *CardDao) GetByIDForUpdate(ctx context.Context, id uint64) (*Card, error) {
	if id == 0 {
		return nil, ad.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: id,
			},
			ForUpdate: true,
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

func (ad *CardDao) Get(ctx context.Context, query *CardQuery) (*Card, error) {
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
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

func (ad *CardDao) Gets(ctx context.Context, query *CardQuery) ([]*Card, error) {
	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) Save(ctx context.Context, model *Card) (uint64, error) {

	db := ad.db.WithContext(ctx)
	if model.ID == 0 {
		cardID, err := utils.RandomIDWithPrefix(model.CategoryID)
		if err != nil {
			return 0, err
		}
		model.ID = cardID
	}

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *CardDao) Update(ctx context.Context, query *CardQuery) (int64, error) {

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
		Model(Card{}).
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

func (ad *CardDao) queryChain(query *CardQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Card)
		structValue := reflect.ValueOf(query.Card)
		structPtrValue := reflect.ValueOf(&query.Card)
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

		notEqualType := reflect.TypeOf(query.NotEqual)
		notEqualValue := reflect.ValueOf(query.NotEqual)
		for i := 0; i < notEqualType.NumField(); i++ {
			if notEqualValue.Field(i).IsZero() {
				continue
			}
			fieldName := stringy.New(notEqualType.Field(i).Name).SnakeCase().ToLower()
			settings := schema.ParseTagSetting(notEqualType.Field(i).Tag.Get("gorm"), ";")
			if f, ok := settings["COLUMN"]; ok {
				fieldName = f
			}
			db.Scopes(ad.notEqualScope(fieldName, notEqualValue.Field(i).Interface()))
		}

		if query.TypeIn != nil {
			db.Scopes(ad.inScope("type", query.TypeIn))
		}

		if query.IDIn != nil {
			db.Scopes(ad.inScope("id", query.IDIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(ad.inScope("user_id", query.UserIDIn))
		}

		if query.AssetTypeIn != nil {
			db.Scopes(ad.inScope("type", query.AssetTypeIn))
		}

		if query.IssueIdIn != nil {
			db.Scopes(ad.inScope("issue_id", query.IssueIdIn))
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

		if !query.BlockedAtLessThan.IsZero() && query.BlockedAtLessThan.UnixMilli() != 0 {
			db.Scopes(ad.compareScope("blocked_at", query.BlockedAtLessThan, false, true))
		}

		return db.
			Scopes(ad.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(ad.pageScope(query.Current, query.PageSize))
	}
}

func (ad *CardDao) GetCountByUserIdAndCurrency(ctx context.Context, userID uint64, currency common.Currency) (int64, error) {
	db := ad.db.WithContext(ctx)
	var count int64
	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Currency: currency,
				UserID:   userID,
			},
		})).
		Count(&count).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ad *CardDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *CardDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
	}
}

func (ad *CardDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (ad *CardDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *CardDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *CardDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (ad *CardDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}

func (ad *CardDao) notEqualScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` != ?", fieldName), field)
	}
}
