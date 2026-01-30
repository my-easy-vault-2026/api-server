package wallet

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

type WalletAddress struct {
	ID         uint64                     `gorm:"default:null"`
	CardID     uint64                     `gorm:"default:null"`
	UserID     uint64                     `gorm:"default:null"`
	Address    string                     `gorm:"default:null"`
	CategoryID uint64                     `gorm:"default:null"`
	Mainnet    common.Mainnet             `gorm:"default:null"`
	Protocol   common.Protocol            `gorm:"default:null"`
	Currency   common.Currency            `gorm:"default:null"`
	Status     common.WalletAddressStatus `gorm:"default:null"`
	CreatedAt  time.Time                  `gorm:"default:null"`
	UpdatedAt  time.Time                  `gorm:"default:null;autoUpdateTime:false"`
}

type WalletAddressQuery struct {
	WalletAddress
	Attrs     WalletAddress
	ForUpdate bool
	ForShare  bool
	IsNull    []string
	MainnetIn []common.Mainnet
	utils.Page
}

type WalletAddressDao struct {
}

func NewWalletAddressDao() *WalletAddressDao {
	return &WalletAddressDao{}
}

func (WalletAddress) TableName() string {
	return "wallet_address"
}

func (wd *WalletAddressDao) GetByID(
	ctx context.Context,
	id uint64,
) (*WalletAddress, error) {

	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
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

func (wd *WalletAddressDao) GetByUserIDMainnetCardID(
	ctx context.Context,
	userID uint64,
	mainnet common.Mainnet,
	cardID uint64,
) (*WalletAddress, error) {

	if userID == 0 || mainnet == 0 || cardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID:  userID,
				Mainnet: mainnet,
				CardID:  cardID,
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

func (wd *WalletAddressDao) GetByUserIDMainnet(
	ctx context.Context,
	userID uint64,
	mainnet common.Mainnet,
) (*WalletAddress, error) {

	if userID == 0 || mainnet == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID:  userID,
				Mainnet: mainnet,
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

func (wd *WalletAddressDao) GetByUserIDMainnetIn(
	ctx context.Context,
	userID uint64,
	mainnets []common.Mainnet,
) (*WalletAddress, error) {

	if userID == 0 || len(mainnets) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID: userID,
			},
			MainnetIn: mainnets,
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

func (wd *WalletAddressDao) ListByAddressMainnetIn(
	ctx context.Context,
	address string,
	mainnets []common.Mainnet,
) ([]*WalletAddress, error) {

	if address == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	if len(mainnets) == 0 {
		return make([]*WalletAddress, 0), nil
	}

	result := make([]*WalletAddress, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				Address: address,
			},
			MainnetIn: mainnets,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*WalletAddress, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (wd *WalletAddressDao) GetByUserIDMainnetProtocolCurrency(
	ctx context.Context,
	userID uint64,
	mainnet common.Mainnet,
	protocol common.Protocol,
	currency common.Currency,
) (*WalletAddress, error) {

	if userID == 0 || mainnet == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	isNull := make([]string, 0, 10)
	if protocol == 0 {
		isNull = append(isNull, "protocol")
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID:   userID,
				Mainnet:  mainnet,
				Protocol: protocol,
				Currency: currency,
			},
			IsNull: isNull,
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

func (wd *WalletAddressDao) GetByUserIDMainnetProtocolCurrencyCardID(
	ctx context.Context,
	userID uint64,
	mainnet common.Mainnet,
	protocol common.Protocol,
	currency common.Currency,
	cardID uint64,
) (*WalletAddress, error) {

	if userID == 0 || mainnet == 0 || cardID == 0 || currency == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	isNull := make([]string, 0, 10)
	if protocol == 0 {
		isNull = append(isNull, "protocol")
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID:   userID,
				Mainnet:  mainnet,
				Protocol: protocol,
				Currency: currency,
				CardID:   cardID,
			},
			IsNull: isNull,
		})).
		Take(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wd *WalletAddressDao) GetByUserIDMainnetCurrency(
	ctx context.Context,
	userID uint64,
	mainnet common.Mainnet,
	currency common.Currency,
	cardId uint64,
) (*WalletAddress, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				UserID:   userID,
				Mainnet:  mainnet,
				Currency: currency,
				CardID:   cardId,
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

func (wd *WalletAddressDao) GetByMainnetAddress(
	ctx context.Context,
	mainnet common.Mainnet,
	address string,
) (*WalletAddress, error) {

	if mainnet == 0 || address == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				Mainnet: mainnet,
				Address: address,
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

func (wd *WalletAddressDao) GetByMainnetUserID(
	ctx context.Context,
	mainnet common.Mainnet,
	userID uint64,
) (*WalletAddress, error) {

	if mainnet == 0 || userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(&WalletAddressQuery{
			WalletAddress: WalletAddress{
				Mainnet: mainnet,
				UserID:  userID,
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

func (wd *WalletAddressDao) Get(ctx context.Context, query *WalletAddressQuery) (*WalletAddress, error) {
	result := &WalletAddress{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wd *WalletAddressDao) Gets(ctx context.Context, query *WalletAddressQuery) ([]*WalletAddress, error) {
	result := make([]*WalletAddress, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&WalletAddress{}).
		Scopes(wd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*WalletAddress, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (wd *WalletAddressDao) Save(ctx context.Context, wallet *WalletAddress) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.Create(wallet)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return wallet.ID, nil
}

func (wd *WalletAddressDao) Update(ctx context.Context, query *WalletAddressQuery) (int64, error) {

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
		Model(&WalletAddress{}).
		Where("id = ?", query.ID).
		Scopes(wd.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (wd *WalletAddressDao) queryChain(query *WalletAddressQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.WalletAddress)
		structValue := reflect.ValueOf(query.WalletAddress)
		structPtrValue := reflect.ValueOf(&query.WalletAddress)
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
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(wd.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(wd.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		if query.MainnetIn != nil {
			db.Scopes(wd.inScope("mainnet", query.MainnetIn))
		}

		return db.
			Scopes(wd.nullScope(query.IsNull, true))
	}
}

func (wd *WalletAddressDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (wd *WalletAddressDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (wd *WalletAddressDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}
