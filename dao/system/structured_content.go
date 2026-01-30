package system

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"shared-modules/common"
	"shared-modules/utils"
	"strings"
	"time"

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type StructuredContent struct {
	ID           uint64                   `json:"id"`
	CategoryID   common.ParameterCategory `json:"categoryId"`
	CategoryName string                   `json:"categoryName"`
	Name         string                   `json:"name"`
	Content      string                   `json:"content" gorm:"column:content;default:null"`
	Language     common.Language          `json:"language" gorm:"default:null"`
	Scene        common.ContentScene      `json:"scene"`
	CustomID     string                   `json:"customId" gorm:"default:null"`
	Channel      common.ContentChannel    `json:"channel" gorm:"default:null"`
	Status       common.ContentStatus     `json:"status"`
	Remark       string                   `json:"remark" gorm:"default:null"`
	CreatedAt    time.Time                `json:"createdAt" gorm:"default:null"`
	UpdatedAt    time.Time                `json:"updatedAt" gorm:"default:null;autoUpdateTime:false"`
}

type StructuredContentQuery struct {
	StructuredContent
	Attrs      StructuredContent
	ForUpdate  bool
	ForShare   bool
	CustomIDIn []string
	utils.Page
}

type StructuredContentDao struct {
}

func NewStructuredContentDao() *StructuredContentDao {
	return &StructuredContentDao{}
}

func (StructuredContent) TableName() string {
	return "structured_content"
}

func (sd *StructuredContentDao) GetBySceneLanguage(ctx context.Context, scene common.ContentScene, language string) (*StructuredContent, error) {
	if scene == 0 || language == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &StructuredContent{}
	db := utils.GetDB(ctx)

	l := common.Language(strings.ToLower(strings.ReplaceAll(string(language), "_", "-")))

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(StructuredContent{}).
			Scopes(sd.queryChain(&StructuredContentQuery{
				StructuredContent: StructuredContent{
					Scene:    scene,
					Language: l,
				},
			})).
			First(result)
	}, func(tx *gorm.DB) (*StructuredContent, error) {
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

func (sd *StructuredContentDao) GetBySceneCustomIDs(ctx context.Context, scene common.ContentScene, customIDs []string) (*StructuredContent, error) {
	if scene == 0 || len(customIDs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &StructuredContent{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(StructuredContent{}).
			Scopes(sd.queryChain(&StructuredContentQuery{
				StructuredContent: StructuredContent{
					Scene: scene,
				},
				CustomIDIn: customIDs,
			})).
			First(result)
	}, func(tx *gorm.DB) (*StructuredContent, error) {
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

func (sd *StructuredContentDao) GetBySceneCustomIDLanguage(ctx context.Context, scene common.ContentScene, customID string, language string) (*StructuredContent, error) {
	if scene == 0 || customID == "" || language == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &StructuredContent{}
	db := utils.GetDB(ctx)

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			Model(StructuredContent{}).
			Scopes(sd.queryChain(&StructuredContentQuery{
				StructuredContent: StructuredContent{
					Scene:    scene,
					Language: common.Language(language),
					CustomID: customID,
				},
			})).
			First(result)
	}, func(tx *gorm.DB) (*StructuredContent, error) {
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

func (sd *StructuredContentDao) ListBySceneCustomIDsLanguage(ctx context.Context, scene common.ContentScene, customIDs []string, language string) ([]*StructuredContent, error) {
	if scene == 0 || language == "" || len(customIDs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*StructuredContent, 0)
	db := utils.GetDB(ctx)

	l := common.Language(strings.ToLower(strings.ReplaceAll(string(language), "_", "-")))

	return utils.L2CQuery(ctx, db, func(tx *gorm.DB) *gorm.DB {
		return tx.
			// Model(&StructuredContent{}).
			Table("structured_content").
			Scopes(sd.queryChain(&StructuredContentQuery{
				StructuredContent: StructuredContent{
					Scene:    scene,
					Language: l,
				},
				CustomIDIn: customIDs,
			})).
			Scan(&result)
	}, func(tx *gorm.DB) ([]*StructuredContent, error) {
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

func (sd *StructuredContentDao) Get(ctx context.Context, query *StructuredContentQuery) (*StructuredContent, error) {
	result := &StructuredContent{}
	db := utils.GetDB(ctx)

	err := db.
		Model(StructuredContent{}).
		Scopes(sd.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (sd *StructuredContentDao) Gets(ctx context.Context, query *StructuredContentQuery) ([]*StructuredContent, error) {
	result := make([]*StructuredContent, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(StructuredContent{}).
		Scopes(sd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*StructuredContent{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (sd *StructuredContentDao) Save(ctx context.Context, model *StructuredContent) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (sd *StructuredContentDao) Update(ctx context.Context, query *StructuredContentQuery) (int64, error) {

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
		Model(StructuredContent{}).
		Where("id = ?", query.ID).
		Scopes(sd.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (pd *StructuredContentDao) queryChain(query *StructuredContentQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.StructuredContent)
		structValue := reflect.ValueOf(query.StructuredContent)
		structPtrValue := reflect.ValueOf(&query.StructuredContent)
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

		if query.CustomIDIn != nil {
			db.Scopes(pd.inScope("custom_id", query.CustomIDIn))
		}

		return db
	}
}

func (pd *StructuredContentDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (pd *StructuredContentDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
