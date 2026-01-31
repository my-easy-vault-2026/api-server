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
	ID                uint64
	Email             string `gorm:"column:email;uniqueIndex"`
	SystemEmail       string `gorm:"column:system_email;uniqueIndex"`
	PinCode           string `gorm:"default:null"`
	Salt              string `gorm:"default:null"`
	CountryCode       int    `gorm:"default:null"`
	PhoneNumber       string `gorm:"default:null"`
	MerchantID        uint64
	FirstName         string              `gorm:"default:null"`
	LastName          string              `gorm:"default:null"`
	NationCode        common.NationCode   `gorm:"default:null"`
	Channel           string              `gorm:"default:null"`
	KycLevel          common.KYCLevel     `gorm:"default:null"`
	CoinfaceMain      common.CoinfaceMain `gorm:"default:null"`
	Gender            common.Gender       `gorm:"default:null"`
	Role              common.Role
	BlockStatus       common.UserBlockStatus  `gorm:"default:null;column:block_status"`
	BlockReason       *common.UserBlockReason `gorm:"default:null;column:block_reason"`
	CulmulativeEPoint decimal.Decimal         `gorm:"default:null;column:cumulative_epoint"`
	EPointLevel       common.EPointLevel      `gorm:"default:null;column:epoint_level"`
	AutoTopUp         common.AutoTopUpStatus  `gorm:"default:null"`
	Auto3DS           common.Auto3DSStatus    `gorm:"default:null;column:auto_3ds"`
	ATMToggle         common.ATMToggle        `gorm:"default:null;column:atm_toggle"`
	Language          common.Language         `gorm:"default:null"`
	PromotionCode     string
	CreatedAt         time.Time `gorm:"default:null"`
	UpdatedAt         time.Time `gorm:"default:null;autoUpdateTime:false"`
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
	utils.Page
}

type UserDao struct {
	db  infra.Database
	env *lib.Env
}

func NewUserDao(db infra.Database, env *lib.Env) *UserDao {
	return &UserDao{db: db, env: env}
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

func (ud *UserDao) GetBySystemEmail(ctx context.Context, systemEmail string) (*User, error) {
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				SystemEmail: systemEmail,
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

func (ud *UserDao) GetByUserIDMerchantID(ctx context.Context, userID uint64, merchantID uint64) (*User, error) {

	if userID == 0 || merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID:         userID,
				MerchantID: merchantID,
				Role:       common.ROLE_MERCHANT_USER,
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

func (ud *UserDao) GetByEmailMerchantID(ctx context.Context, email string, merchantID uint64) (*User, error) {
	if email == "" || merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Email:      email,
				MerchantID: merchantID,
				Role:       common.ROLE_MERCHANT_USER,
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

func (ud *UserDao) GetByEmailRole(ctx context.Context, email string, role common.Role) (*User, error) {
	if role == 0 || email == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
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

func (ud *UserDao) GetByEmailRoleIn(ctx context.Context, email string, roles []common.Role) (*User, error) {
	if len(roles) == 0 || email == "" || roles[0] == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Email: email,
			},
			RoleIn: roles,
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

func (ud *UserDao) GetByUserIDPinCode(ctx context.Context, userID uint64, pinCode string) (*User, error) {
	if userID == 0 || pinCode == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID:      userID,
				PinCode: pinCode,
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

func (ud *UserDao) ListByUserIDMerchantID(ctx context.Context, userID uint64, merchantID uint64) ([]*User, error) {
	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID:         userID,
				MerchantID: merchantID,
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

func (ud *UserDao) ListByMerchantUserIDIn(ctx context.Context, userIDs []uint64, merchantID uint64) ([]*User, error) {
	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				MerchantID: merchantID,
			},
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

func (ud *UserDao) ListByMerchantEmailIn(ctx context.Context, emails []string, merchantID uint64) ([]*User, error) {
	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				MerchantID: merchantID,
			},
			EmailIn: emails,
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

func (ud *UserDao) ListByUserIDInRole(ctx context.Context, userIDs []uint64, role common.Role) ([]*User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Role: role,
			},
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

func (ud *UserDao) ListByEmail(ctx context.Context, email string) ([]*User, error) {
	if email == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Email: email,
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

func (ud *UserDao) ListByEmailRole(ctx context.Context, email string, role common.Role) ([]*User, error) {
	if email == "" || role == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
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

func (ud *UserDao) PageAgentUsers(ctx context.Context, promotionCode string, userID uint64, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				PromotionCode: promotionCode,
				ID:            userID,
			},
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) PageByRoleEmailPhone(ctx context.Context, role common.Role, email string, countryCode int, phoneNumber string, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	if role == 0 || (email == "" && countryCode == 0 && phoneNumber == "") {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Role:        role,
				Email:       email,
				CountryCode: countryCode,
				PhoneNumber: phoneNumber,
			},
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) PageByMerchantIDUserIDIn(ctx context.Context, merchantID uint64, userIDs []uint64, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	if merchantID == 0 || len(userIDs) == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				MerchantID: merchantID,
			},
			IDIn: userIDs,
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_ASC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) PageByRole(ctx context.Context, role common.Role, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	if role == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Role: role,
			},
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) PageByMerchantID(ctx context.Context, merchantID uint64, direction common.OrderDirection, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	if merchantID == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				MerchantID: merchantID,
			},
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: direction,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) PageByRoleEmailPhoneIDIn(ctx context.Context, role common.Role, email string, countryCode int, phone string, idIn []uint64, pageCurrent int, pageSize int) (records []*User, current int, size int, total int, err error) {
	if role == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*User, 0)
	s := int64(0)
	db := ud.db.WithContext(ctx)

	err = db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Role:        role,
				Email:       email,
				CountryCode: countryCode,
				PhoneNumber: phone,
			},
			IDIn: idIn,
		})).
		Count(&s).
		Scopes(ud.queryChain(&UserQuery{
			Page: utils.Page{
				Current:  pageCurrent,
				PageSize: pageSize,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, 0, 0, 0, err
	}
	return result, pageCurrent, pageSize, int(s), nil
}

func (ud *UserDao) GetByUserIDAndPromotionCode(ctx context.Context, userID uint64, promotionCode string) (*User, error) {
	result := &User{}
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID:            userID,
				PromotionCode: promotionCode,
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

func (ud *UserDao) GetUserIDsByLanguage(ctx context.Context, language common.Language) ([]*User, error) {
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)
	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				Language: language,
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

func (ud *UserDao) GetReferredCountByPromotionCode(ctx context.Context, promotionCode string) (int64, error) {
	if promotionCode == "" {
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := int64(0)
	db := ud.db.WithContext(ctx)

	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				PromotionCode: promotionCode,
			},
		})).
		Count(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return result, nil
}
func (ud *UserDao) GetReferredUserByPromotionCode(ctx context.Context, promotionCode string) ([]*User, error) {
	result := make([]*User, 0)
	db := ud.db.WithContext(ctx)
	err := db.
		Model(&User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				PromotionCode: promotionCode,
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

func (ud *UserDao) UpdateStatusByUserID(ctx context.Context, userID uint64, status common.UserBlockStatus) error {
	db := ud.db.WithContext(ctx)

	err := db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID: userID,
			},
		})).
		Updates(map[string]interface{}{
			"block_status": status,
			"block_reason": nil,
		}).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err != nil {
		return err
	}

	return nil

}

func (ud *UserDao) UpdateSystemEmailByUserID(ctx context.Context, userID uint64, systemEmail string) error {
	db := ud.db.WithContext(ctx)

	err := db.
		Model(User{}).
		Scopes(ud.queryChain(&UserQuery{
			User: User{
				ID: userID,
			},
		})).
		Updates(map[string]interface{}{
			"system_email": systemEmail,
		}).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err != nil {
		return err
	}

	return nil

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
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
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
		return 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
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
