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

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type InterestOrder struct {
	ID              uint64                     `gorm:"default:null"`
	Code            common.FinancialCode       `gorm:"default:null"`
	ParentOrderNO   string                     `gorm:"default:null"`
	OrderNO         string                     `gorm:"default:null"`
	UserID          uint64                     `gorm:"default:null"`
	FromCardType    common.AssetType           `gorm:"default:null"`
	FromCardID      uint64                     `gorm:"default:null"`
	FromCategoryID  uint64                     `gorm:"default:null"`
	FromCurrency    common.Currency            `gorm:"default:null"`
	ToCardType      common.AssetType           `gorm:"default:null"`
	ToCardID        uint64                     `gorm:"default:null"`
	ToCategoryID    uint64                     `gorm:"default:null"`
	ToCurrency      common.Currency            `gorm:"default:null"`
	InterestRate    decimal.Decimal            `gorm:"default:null"`
	PrincipalAmount *decimal.Decimal           `gorm:"default:null"`
	InterestAmount  decimal.Decimal            `gorm:"default:null"`
	TruncateAmount  *decimal.Decimal           `gorm:"default:null"`
	CalculateCount  int                        `gorm:"default:null"`
	Status          common.InterestOrderStatus `gorm:"default:null"`
	CalculatedAt    time.Time                  `gorm:"default:null"`
	CreatedAt       time.Time                  `gorm:"default:null"`
	UpdatedAt       time.Time                  `gorm:"default:null;autoUpdateTime:false"`
}

type InterestOrderQuery struct {
	InterestOrder
	Attrs              InterestOrder
	ForUpdate          bool
	ForShare           bool
	StatusIn           []common.InterestOrderStatus
	IsNull             []string
	CalculatedAtFrom   time.Time
	CalculatedAtTo     time.Time
	CalculatedAtString string
	OrderBy            string
	OrderDirection     common.OrderDirection
	utils.Page
}

type InterestOrderDao struct{}

func NewInterestOrderDao() *InterestOrderDao {
	return &InterestOrderDao{}
}

func (InterestOrder) TableName() string {
	return "interest_order"
}

func (ad *InterestOrderDao) Save(ctx context.Context, model *InterestOrder) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(InterestOrder{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *InterestOrderDao) Saves(ctx context.Context, models []*InterestOrder) (int64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(InterestOrder{}).
		CreateInBatches(models, len(models))

	if ret.Error != nil {
		return ret.RowsAffected, ret.Error
	}
	return ret.RowsAffected, nil
}

func (ad *InterestOrderDao) GetByToCardIDCodeCalculatedAt(ctx context.Context, toCardID uint64, code common.FinancialCode, calculatedAt time.Time) (*InterestOrder, error) {

	if toCardID == 0 || code == "" || calculatedAt.IsZero() {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &InterestOrder{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&InterestOrder{}).
		Scopes(ad.queryChain(&InterestOrderQuery{
			InterestOrder: InterestOrder{
				ToCardID:     toCardID,
				Code:         code,
				CalculatedAt: calculatedAt,
			},
			IsNull: []string{"parent_order_no"},
			// CalculatedAtString: calculatedAt.Format("2006-01-02 15:04:05.000000"),
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

func (ad *InterestOrderDao) GetByFromCardIDCodeCalculatedAt(ctx context.Context, fromCardID uint64, code common.FinancialCode, calculatedAt time.Time) (*InterestOrder, error) {

	if fromCardID == 0 || code == "" || calculatedAt.IsZero() {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &InterestOrder{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&InterestOrder{}).
		Scopes(ad.queryChain(&InterestOrderQuery{
			InterestOrder: InterestOrder{
				FromCardID:   fromCardID,
				Code:         code,
				CalculatedAt: calculatedAt,
			},
			IsNull: []string{"parent_order_no"},
			// CalculatedAtString: calculatedAt.Format("2006-01-02 15:04:05.000000"),
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

func (ad *InterestOrderDao) ListByCodeToCardIDCalculatedAt(ctx context.Context, code common.FinancialCode, cardID uint64, calculatedAtFrom time.Time) ([]*InterestOrder, error) {

	if cardID == 0 || calculatedAtFrom.IsZero() {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*InterestOrder, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(InterestOrder{}).
		Scopes(ad.queryChain(&InterestOrderQuery{
			InterestOrder: InterestOrder{
				Code:     code,
				ToCardID: cardID,
			},
			CalculatedAtFrom: calculatedAtFrom,
			IsNull:           []string{"parent_order_no"},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*InterestOrder{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *InterestOrderDao) ListByCodeToCardIDCalculatedAtOrderByCalculatedAt(ctx context.Context, code common.FinancialCode, cardID uint64, calculatedAtFrom time.Time) ([]*InterestOrder, error) {

	if cardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*InterestOrder, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(InterestOrder{}).
		Scopes(ad.queryChain(&InterestOrderQuery{
			InterestOrder: InterestOrder{
				Code:     code,
				ToCardID: cardID,
			},
			IsNull:           []string{"parent_order_no"},
			CalculatedAtFrom: calculatedAtFrom,
			OrderBy:          "calculated_at",
			OrderDirection:   common.ORDER_DIRECTION_ASC,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*InterestOrder{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (td *InterestOrderDao) Update(ctx context.Context, query *InterestOrderQuery) (int64, error) {
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
		kind := structValue.Field(i).Kind()
		switch kind {
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
			default:
				continue
			}
		default:
			continue
		}
	}
	ret := db.
		Model(&InterestOrder{}).
		Scopes(td.queryChain(query)).
		Updates(attrs)
	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if ret.Error != nil {
		return 0, ret.Error
	}
	return ret.RowsAffected, nil
}

func (td *InterestOrderDao) queryChain(query *InterestOrderQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.InterestOrder)
		structValue := reflect.ValueOf(query.InterestOrder)
		structPtrValue := reflect.ValueOf(&query.InterestOrder)
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
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Interface:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Interface()))
			case reflect.Pointer:
				db.Scopes(td.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(td.equalScope(fieldName, value))
				case reflect.TypeOf((*time.Time)(nil)):
					t := ptr.(*time.Time)
					if t.IsZero() {
						continue
					}
					db.Scopes(td.equalScope(fieldName, t))
				}
			default:
				continue
			}
		}
		if query.StatusIn != nil {
			db.Scopes(td.inScope("status", query.StatusIn))
		}

		if !query.CalculatedAtFrom.IsZero() && query.CalculatedAtFrom.Unix() != 0 {
			db.Scopes(td.compareScope("calculated_at", query.CalculatedAtFrom, true, true))
		}

		if !query.CalculatedAtTo.IsZero() && query.CalculatedAtTo.Unix() != 0 {
			db.Scopes(td.compareScope("calculated_at", query.CalculatedAtTo, false, true))
		}

		if query.CalculatedAtString != "" {
			db.Scopes(td.equalScope("calculated_at", query.CalculatedAtString))
		}

		if query.IsNull != nil {
			db.Scopes(td.nullScope(query.IsNull, true))
		}

		return db.
			Scopes(td.orderByScope(query.OrderBy, query.OrderDirection))
	}
}

func (td *InterestOrderDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (td *InterestOrderDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *InterestOrderDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (bd *InterestOrderDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (cd *InterestOrderDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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
