package user

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

type Group struct {
	ID           uint64
	Name         string
	Description  string      `gorm:"default:null"`
	Role         common.Role `gorm:"default:null"`
	ExclusiveIDs string      `gorm:"default:null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GroupQuery struct {
	Group
	Attrs     Group
	ForUpdate bool
	ForShare  bool
	common.Page
}

type GroupDao struct {
	db        infra.Database
	env       *lib.Env
	beBuilder *lib.BEBuilder
}

func NewGroupDao(db infra.Database, env *lib.Env, beBuilder *lib.BEBuilder) *GroupDao {
	return &GroupDao{db: db, env: env, beBuilder: beBuilder}
}

func (gd *GroupDao) WithTx(tx *gorm.DB) *GroupDao {
	if gd == nil {
		return gd
	}
	newDao := *gd
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Group) TableName() string {
	return "group"
}

func (gd *GroupDao) GetByID(ctx context.Context, id uint64) (*Group, error) {
	if id == 0 {
		return nil, gd.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Group{}
	db := gd.db.WithContext(ctx)

	err := db.
		Model(&Group{}).
		Scopes(gd.queryChain(&GroupQuery{
			Group: Group{
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

func (gd *GroupDao) Get(ctx context.Context, query *GroupQuery) (*Group, error) {
	result := &Group{}
	db := gd.db.WithContext(ctx)

	err := db.
		Model(&Group{}).
		Scopes(gd.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (gd *GroupDao) Gets(ctx context.Context, query *GroupQuery) ([]Group, error) {
	result := make([]Group, 0)
	db := gd.db.WithContext(ctx)

	err := db.
		Model(&Group{}).
		Scopes(gd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]Group, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (gd *GroupDao) Save(ctx context.Context, group *Group) (uint64, error) {

	db := gd.db.WithContext(ctx)
	groupID := utils.RandomID()
	group.ID = groupID
	ret := db.Create(group)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return groupID, nil
}

func (gd *GroupDao) Update(ctx context.Context, query *GroupQuery) (int64, error) {

	if query.ID == 0 {
		// TODO: define dao error
		return 0, errors.ErrUnsupported
	}

	db := gd.db.WithContext(ctx)

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
		Model(Group{}).
		Where("id = ?", query.ID).
		Scopes(gd.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (gd *GroupDao) queryChain(query *GroupQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Group)
		structValue := reflect.ValueOf(query.Group)
		structPtrValue := reflect.ValueOf(&query.Group)
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
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(gd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(gd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		return db
	}
}

func (gd *GroupDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}
