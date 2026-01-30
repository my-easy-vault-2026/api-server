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

	"github.com/gobeam/stringy"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DeleteUser struct {
	ID          uint64
	Email       string          `gorm:"column:email;uniqueIndex"`
	PinCode     string          `gorm:"default:null"`
	CountryCode int             `gorm:"default:null"`
	PhoneNumber string          `gorm:"default:null"`
	FirstName   string          `gorm:"default:null"`
	LastName    string          `gorm:"default:null"`
	NationCode  string          `gorm:"default:null"`
	Gender      common.Gender   `gorm:"default:null"`
	KycLevel    common.KYCLevel `gorm:"default:null"`
	CreatedAt   time.Time       `gorm:"default:null"`
	UpdatedAt   time.Time       `gorm:"default:null;autoUpdateTime:false"`
}

type DeleteUserQuery struct {
	DeleteUser
	Attrs     DeleteUser
	ForUpdate bool
	ForShare  bool
	utils.Page
}

type DeleteUserDao struct {
}

func NewDeleteUserDao() *DeleteUserDao {
	return &DeleteUserDao{}
}

func (DeleteUser) TableName() string {
	return "delete_user"
}

func (ud *DeleteUserDao) GetByUserID(ctx context.Context, userID uint64) (*DeleteUser, error) {
	result := &DeleteUser{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&DeleteUser{}).
		Scopes(ud.queryChain(&DeleteUserQuery{
			DeleteUser: DeleteUser{
				ID: userID,
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

func (ud *DeleteUserDao) GetByEmail(ctx context.Context, email string) (*DeleteUser, error) {
	result := &DeleteUser{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&DeleteUser{}).
		Scopes(ud.queryChain(&DeleteUserQuery{
			DeleteUser: DeleteUser{
				Email: email,
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

func (ud *DeleteUserDao) Save(ctx context.Context, user *DeleteUser) (uint64, error) {

	db := utils.GetDB(ctx)
	ret := db.Create(user)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return user.ID, nil
}

func (ud *DeleteUserDao) Update(ctx context.Context, query *DeleteUserQuery) (int64, error) {

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
		Model(&DeleteUser{}).
		Where("id = ?", query.ID).
		Scopes(ud.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ud *DeleteUserDao) queryChain(query *DeleteUserQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.DeleteUser)
		structValue := reflect.ValueOf(query.DeleteUser)
		structPtrValue := reflect.ValueOf(&query.DeleteUser)
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
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(ud.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(ud.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		return db
	}
}

func (ud *DeleteUserDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}
