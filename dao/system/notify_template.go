package system

import (
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/utils"

	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type NotifyTemplate struct {
	ID              uint64            `json:"id"`
	Code            string            `json:"code"`
	SubCode         string            `json:"subCode"`
	Name            string            `json:"name"`
	Subject         string            `json:"subject"`
	Template        string            `json:"template"`
	NotifyType      common.NotifyType `json:"notifyType"`
	Language        common.Language   `json:"language"`
	Extra           string            `json:"extra"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `gorm:"default:null;autoUpdateTime:false"`
	CreateUser      uint64            `json:"createUser"`
	UpdateUser      uint64            `json:"updateUser"`
	ForceUpdateHash int64             `json:"forceUpdateHash,omitempty"` // 用於強制更新模板的hash值
}

type NotifyTemplateQuery struct {
	NotifyTemplate
	Attrs          NotifyTemplate
	ForUpdate      bool
	ForShare       bool
	OrderBy        string
	OrderDirection common.OrderDirection
	utils.Page
}

type NotifyTemplateDao struct {
}

func NewNotifyTemplateDao() *NotifyTemplateDao {
	return &NotifyTemplateDao{}
}

func (NotifyTemplate) TableName() string {
	return "notify_template"
}

func (nd *NotifyTemplateDao) GetNotifyTemplate(ctx context.Context, code common.MsgOPCode, language common.Language, notifyType common.NotifyType) (*NotifyTemplate, error) {
	result := &NotifyTemplate{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(NotifyTemplate{}).
			Where("code=? AND language=? AND notify_type=?", code, language, notifyType).
			First(result)
	}, func(tx *gorm.DB) (*NotifyTemplate, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&utils.L2CacheConfig{
			Level:         []common.L2CacheLevel{common.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: utils.Config.System.L2CacheExpireSeconds,
		})
}

func (nd *NotifyTemplateDao) GetNotifyTemplateByCodeSubCode(ctx context.Context, code common.MsgOPCode, subCode string, language common.Language, notifyType common.NotifyType) (*NotifyTemplate, error) {
	result := &NotifyTemplate{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(NotifyTemplate{}).
			Where("code=? AND sub_code=? AND language=? AND notify_type=?", code, subCode, language, notifyType).
			First(result)
	}, func(tx *gorm.DB) (*NotifyTemplate, error) {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		if tx.Error != nil {
			return nil, tx.Error
		}

		return result, nil
	},
		&utils.L2CacheConfig{
			Level:         []common.L2CacheLevel{common.L2_CACHE_LEVEL_REDIS},
			ExpireSeconds: utils.Config.System.L2CacheExpireSeconds,
		})
}

func (nd *NotifyTemplateDao) GetNotifyTemplateList(ctx context.Context, form *entities.GetTemplateListForm) ([]*NotifyTemplate, error) {
	result := make([]*NotifyTemplate, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&NotifyTemplate{}).
		Scopes(nd.queryChain(&NotifyTemplateQuery{
			NotifyTemplate: NotifyTemplate{
				Code:       form.Code,
				NotifyType: form.NotifyType,
				Language:   form.Language,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*NotifyTemplate, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (nd *NotifyTemplateDao) GetTemplatePageByLanguageAndType(ctx context.Context, form *entities.GetTemplatePageForm, language common.Language, notifyType common.NotifyType) (records []*NotifyTemplate, current int, size int, total int, err error) {
	s := int64(0)
	result := make([]*NotifyTemplate, 0)
	db := utils.GetDB(ctx)

	err = db.
		Model(NotifyTemplate{}).
		Scopes(nd.queryChain(&NotifyTemplateQuery{
			NotifyTemplate: NotifyTemplate{
				Language:   language,
				NotifyType: notifyType,
				Code:       form.Code,
			},
		})).
		Count(&s).
		Scopes(nd.queryChain(&NotifyTemplateQuery{
			Page: utils.Page{
				Current:  form.Current,
				PageSize: form.PageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, form.Current, form.PageSize, int(s), nil
}

func (nd *NotifyTemplateDao) DeleteByID(ctx context.Context, id string) error {
	db := utils.GetDB(ctx)

	return db.
		Where("id=?", id).
		Delete(&NotifyTemplate{}).Error
}

func (nd *NotifyTemplateDao) Save(ctx context.Context, notifyTemplate *NotifyTemplate) (uint64, error) {

	db := utils.GetDB(ctx)
	ret := db.Create(notifyTemplate)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return notifyTemplate.ID, nil
}

func (nd *NotifyTemplateDao) Update(ctx context.Context, query *NotifyTemplateQuery) (int64, error) {

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
			}
		default:
			continue
		}
	}

	ret := db.
		Model(&NotifyTemplate{}).
		Where("id = ?", query.ID).
		Scopes(nd.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (nd *NotifyTemplateDao) queryChain(query *NotifyTemplateQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.NotifyTemplate)
		structValue := reflect.ValueOf(query.NotifyTemplate)
		structPtrValue := reflect.ValueOf(&query.NotifyTemplate)
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
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(nd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(nd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		return db.
			Scopes(nd.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(nd.pageScope(query.Current, query.PageSize))
	}
}

func (nd *NotifyTemplateDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (nd *NotifyTemplateDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (nd *NotifyTemplateDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && size != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
