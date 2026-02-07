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
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type User struct {
	ID        uint64
	Email     string `gorm:"column:email;uniqueIndex"`
	PinCode   string `gorm:"default:null"`
	Salt      string `gorm:"default:null"`
	Role      common.Role
	CreatedAt time.Time `gorm:"default:null"`
	UpdatedAt time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type UserQuery struct {
	User
	Attrs          User
	ForUpdate      bool
	ForShare       bool
	IDIn           []uint64
	RoleIn         []common.Role
	EmailIn        []string
	OrderBy        string
	OrderDirection common.OrderDirection
	common.Page
}

type UserDao struct {
	db        infra.Database
	env       *lib.Env
	beBuilder *lib.BEBuilder
}

func NewUserDao(db infra.Database, env *lib.Env, beBuilder *lib.BEBuilder) *UserDao {
	return &UserDao{db: db, env: env, beBuilder: beBuilder}
}

func (ud *UserDao) WithTx(tx *gorm.DB) *UserDao {
	newDao := *ud
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (User) TableName() string {
	return "user"
}

func (ud *UserDao) GetByUserID(ctx context.Context, userID uint64) (*User, error) {
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
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

func (ud *UserDao) GetByUserIDForUpdate(ctx context.Context, userID uint64) (*User, error) {
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID: userID,
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

func (ud *UserDao) GetByEmailRole(ctx context.Context, email string, role common.Role) (*User, error) {
	if role == 0 || email == "" {
		return nil, ud.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Email: email,
				Role:  role,
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

func (ud *UserDao) ListByUserIDIn(ctx context.Context, userIDs []uint64) ([]*User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{},
			IDIn: userIDs,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*User, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ud *UserDao) ListByEmailRole(ctx context.Context, email string, role common.Role) ([]*User, error) {
	if email == "" || role == 0 {
		return nil, ud.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Email: email,
				Role:  role,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*User, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ud *UserDao) DeleteByID(ctx context.Context, userID uint64) error {
	db := ud.db.WithContext(ctx)

	return db.
		Where("id=?", userID).
		Delete(&User{}).Error
}

func (ud *UserDao) Get(ctx context.Context, query *UserQuery) (*User, error) {
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ud *UserDao) Gets(ctx context.Context, query *UserQuery) ([]User, error) {
	result := make([]User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]User, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ud *UserDao) Save(ctx context.Context, user *User) (uint64, error) {

	db := ud.db.WithContext(ctx)
	userID := utils.RandomID()
	user.ID = userID
	ret := db.Create(user)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return userID, nil
}

func (ud *UserDao) Update(ctx context.Context, query *UserQuery) (int64, error) {

	if query.ID == 0 {
		return 0, ud.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	db := ud.db.WithContext(ctx)

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
			case reflect.TypeOf((*decimal.Decimal)(nil)):
				value := ptr.(*decimal.Decimal)
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
		Model(&User{}).
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

func (ud *UserDao) UpdateIncr(ctx context.Context, query *UserQuery, field string, amount decimal.Decimal) (int64, error) {

	if query.ID == 0 {
		return 0, ud.beBuilder.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	db := ud.db.WithContext(ctx)

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
		Model(&User{}).
		Where("id = ?", query.ID).
		Scopes(ud.queryChain(query)).
		Update(field, gorm.Expr(fmt.Sprintf("%s + ?", field), amount))

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ud *UserDao) queryChain(query *UserQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.User)
		structValue := reflect.ValueOf(query.User)
		structPtrValue := reflect.ValueOf(&query.User)
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

		if query.IDIn != nil {
			db.Scopes(ud.inScope("id", query.IDIn))
		}

		if query.RoleIn != nil {
			db.Scopes(ud.inScope("role", query.RoleIn))
		}

		if query.EmailIn != nil {
			db.Scopes(ud.inScope("email", query.EmailIn))
		}

		if query.ForUpdate {
			db.Scopes(ud.forScope("UPDATE"))
		}

		if query.ForShare {
			db.Scopes(ud.forScope("SHARE"))
		}

		return db.
			Scopes(ud.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(ud.pageScope(query.Current, query.PageSize))
	}
}

func (ud *UserDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ud *UserDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (ud *UserDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (ud *UserDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (ud *UserDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && size != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}

func (ud *UserDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}
