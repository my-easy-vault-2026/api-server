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

type NotifyRecord struct {
	ID              uint64 `gorm:"default:null"`
	UserID          uint64
	TemplateID      uint64
	TemplateCode    string
	Language        common.Language
	MessageID       uint64 `gorm:"default:null"`
	Sender          string `gorm:"default:null"`
	Recipient       string `gorm:"default:null"`
	Token           string `gorm:"default:null"`
	Subject         string `gorm:"default:null"`
	Content         string
	Extra           string `gorm:"default:null"`
	NotifyType      common.NotifyType
	Passageway      common.Passageway `gorm:"default:null"`
	Status          common.NotifyStatus
	ResponseCode    string    `gorm:"default:null"`
	ResponseContent string    `gorm:"default:null"`
	CreatedAt       time.Time `gorm:"default:null"`
	UpdatedAt       time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type NotifyRecordQuery struct {
	NotifyRecord
	Attrs          NotifyRecord
	ForUpdate      bool
	ForShare       bool
	OrderBy        string
	OrderDirection common.OrderDirection
	CreatedAtFrom  time.Time
	CreatedAtTo    time.Time
	Limit          int
	utils.Page
}
type NotifyRecordDao struct {
}

func NewNotifyRecordDao() *NotifyRecordDao {
	return &NotifyRecordDao{}
}

func (NotifyRecord) TableName() string {
	return "notify_record"
}

func (trd *NotifyRecordDao) Gets(ctx context.Context, query *NotifyRecordQuery) ([]*NotifyRecord, error) {
	result := make([]*NotifyRecord, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(NotifyRecord{}).
		Scopes(trd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*NotifyRecord{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (trd *NotifyRecordDao) Save(ctx context.Context, model *NotifyRecord) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(NotifyRecord{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *NotifyRecordDao) ListByRecipientTemplateCodeNotifyTypeCreatedAt(ctx context.Context, recipient string, templateCode string, notifyType common.NotifyType, createdAtFrom time.Time, createdAtTo time.Time) ([]*NotifyRecord, error) {

	if recipient == "" || templateCode == "" || notifyType == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*NotifyRecord, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(NotifyRecord{}).
		Scopes(trd.queryChain(&NotifyRecordQuery{
			NotifyRecord: NotifyRecord{
				Recipient:    recipient,
				TemplateCode: templateCode,
				NotifyType:   notifyType,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
			CreatedAtFrom:  createdAtFrom,
			CreatedAtTo:    createdAtTo,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*NotifyRecord, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *NotifyRecordDao) ListByUserIDTemplateCodeNotifyTypeCreatedAt(ctx context.Context, userID uint64, templateCode string, notifyType common.NotifyType, createdAtFrom time.Time, createdAtTo time.Time) ([]*NotifyRecord, error) {

	if userID == 0 || templateCode == "" || notifyType == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*NotifyRecord, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(NotifyRecord{}).
		Scopes(trd.queryChain(&NotifyRecordQuery{
			NotifyRecord: NotifyRecord{
				UserID:       userID,
				TemplateCode: templateCode,
				NotifyType:   notifyType,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
			CreatedAtFrom:  createdAtFrom,
			CreatedAtTo:    createdAtTo,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*NotifyRecord, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *NotifyRecordDao) ListAnnouncements(ctx context.Context, templateCode common.MsgOPCode, notifyType common.NotifyType, notifyStatus common.NotifyStatus, limit int) ([]*NotifyRecord, error) {
	if notifyType == 0 || notifyStatus == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*NotifyRecord, 0)
	db := utils.GetDB(ctx)
	err := db.
		Model(NotifyRecord{}).
		Scopes(trd.queryChain(&NotifyRecordQuery{
			NotifyRecord: NotifyRecord{
				NotifyType: notifyType,
				Status:     notifyStatus,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_ASC,
			Limit:          limit,
		})).
		Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*NotifyRecord, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *NotifyRecordDao) Update(ctx context.Context, query *NotifyRecordQuery) (int64, error) {
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
		Model(&NotifyRecord{}).
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

func (trd *NotifyRecordDao) queryChain(query *NotifyRecordQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		structType := reflect.TypeOf(query.NotifyRecord)
		structValue := reflect.ValueOf(query.NotifyRecord)
		structPtrValue := reflect.ValueOf(&query.NotifyRecord)
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

		if !query.CreatedAtFrom.IsZero() && query.CreatedAtFrom.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtFrom, true, true))
		}

		if !query.CreatedAtTo.IsZero() && query.CreatedAtTo.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtTo, false, true))
		}

		if query.Limit != 0 {
			db.Limit(query.Limit)
		}

		return db.
			Scopes(trd.orderByScope(query.OrderBy, query.OrderDirection))
	}
}

func (trd *NotifyRecordDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (trd *NotifyRecordDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *NotifyRecordDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (trd *NotifyRecordDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}
