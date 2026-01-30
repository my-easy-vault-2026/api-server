package notify

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

type CallbackTask struct {
	ID                  uint64
	OrderNO             string              `gorm:"default:null"`
	ThreedsID           uint64              `gorm:"default:null"`
	UserID              uint64              `gorm:"default:null"`
	MerchantID          uint64              `gorm:"default:null"`
	URL                 string              `gorm:"column:url"`
	Type                common.CallbackType `gorm:"column:type"`
	Scene               common.CallbackScene
	MaxAttempt          int                     `gorm:"default:null"`
	IntervalSeconds     int                     `gorm:"default:null"`
	Attempt             int                     `gorm:"default:null"`
	Criteria            common.CallbackCriteria `gorm:"default:null"`
	TimeoutSeconds      int                     `gorm:"default:null"`
	OriginalRequest     string                  `gorm:"default:null"`
	EncryptedRequest    string                  `gorm:"default:null"`
	OriginalResponse    string                  `gorm:"default:null"`
	EncryptedResponse   string                  `gorm:"default:null"`
	Key                 string                  `gorm:"default:null"`
	RequestEncryptType  common.EncryptType      `gorm:"default:null"`
	ResponseEncryptType common.EncryptType      `gorm:"default:null"`
	HTTPCode            int                     `gorm:"default:null;column:http_code"`
	ErrorCode           int                     `gorm:"default:null"`
	ErrorMessage        string                  `gorm:"default:null"`
	Status              common.CallbackStatus
	NotifiedAt          time.Time `gorm:"default:null"`
	CreatedAt           time.Time `gorm:"default:null"`
	UpdatedAt           time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type CallbackTaskQuery struct {
	CallbackTask
	Attrs          CallbackTask
	ForUpdate      bool
	ForShare       bool
	StatusIn       []common.TransferStatus
	OrderBy        string
	OrderDirection common.OrderDirection
	utils.Page
}
type CallbackTaskDao struct {
}

func NewCallbackTaskDao() *CallbackTaskDao {
	return &CallbackTaskDao{}
}

func (CallbackTask) TableName() string {
	return "callback_task"
}

func (cd *CallbackTaskDao) Gets(ctx context.Context, query *CallbackTaskQuery) ([]*CallbackTask, error) {
	result := make([]*CallbackTask, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(CallbackTask{}).
		Scopes(cd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*CallbackTask{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (cd *CallbackTaskDao) ListByTypeStatusOrderByNotifiedAt(ctx context.Context, t common.CallbackType, s common.CallbackStatus) ([]*CallbackTask, error) {
	result := make([]*CallbackTask, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(CallbackTask{}).
		Scopes(cd.queryChain(&CallbackTaskQuery{
			CallbackTask: CallbackTask{
				Type:   t,
				Status: s,
			},
			OrderBy:        "notified_at",
			OrderDirection: common.ORDER_DIRECTION_ASC,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*CallbackTask{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (trd *CallbackTaskDao) Save(ctx context.Context, model *CallbackTask) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(CallbackTask{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *CallbackTaskDao) Update(ctx context.Context, query *CallbackTaskQuery) (int64, error) {
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
		Model(&CallbackTask{}).
		Scopes(trd.queryChain(query)).
		Updates(attrs)
	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if ret.Error != nil {
		return 0, ret.Error
	}
	return ret.RowsAffected, nil
}

func (trd *CallbackTaskDao) queryChain(query *CallbackTaskQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.CallbackTask)
		structValue := reflect.ValueOf(query.CallbackTask)
		structPtrValue := reflect.ValueOf(&query.CallbackTask)
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
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Interface:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Interface()))
			case reflect.Pointer:
				db.Scopes(trd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(trd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}
		if query.StatusIn != nil {
			db.Scopes(trd.inScope("status", query.StatusIn))
		}
		return db.
			Scopes(trd.orderByScope(query.OrderBy, query.OrderDirection))
	}
}

func (trd *CallbackTaskDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (trd *CallbackTaskDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *CallbackTaskDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}
