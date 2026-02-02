package card

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

type Card struct {
	ID           uint64              `gorm:"column:id"`
	UserID       uint64              `gorm:"default:null"`
	Type         common.AssetType    `gorm:"default:null"`
	CategoryID   uint64              `gorm:"default:null"`
	IssueID      string              `gorm:"default:null"`
	Currency     common.Currency     `gorm:"default:null"`
	CurrencyType common.CurrencyType `gorm:"default:null"`
	NationCode   common.NationCode   `gorm:"default:null"`
	Nation       string              `gorm:"default:null"`
	Status       common.CardStatus
	FrozenAt     time.Time `gorm:"default:null"`
	BlockedAt    time.Time `gorm:"default:null"`
	DeletedAt    time.Time `gorm:"default:null"`
	CreatedAt    time.Time `gorm:"default:null"`
	UpdatedAt    time.Time `gorm:"default:null;autoUpdateTime:false"`
}

type CardQuery struct {
	Card
	Attrs             Card
	ForUpdate         bool
	ForShare          bool
	TypeIn            []common.AssetType
	IDIn              []uint64
	UserIDIn          []uint64
	IssueIdIn         []string
	AssetTypeIn       []common.AssetType
	OrderBy           string
	OrderDirection    common.OrderDirection
	BlockedAtLessThan time.Time
	Deleted           bool
	NotEqual          Card
	utils.Page
}

type CardDao struct {
	db  infra.Database
	env *lib.Env
}

func NewCardDao(db infra.Database, env *lib.Env) *CardDao {
	return &CardDao{
		db:  db,
		env: env,
	}
}

func (ad *CardDao) WithTx(tx *gorm.DB) *CardDao {
	newDao := *ad
	newDao.db = infra.Database{DB: tx}
	return &newDao
}

func (Card) TableName() string {
	return "card"
}

func (ad *CardDao) GetByIssueID(ctx context.Context, issueID string) (*Card, error) {

	if issueID == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				IssueID: issueID,
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

func (ad *CardDao) GetByIssueIDs(ctx context.Context, issueIDs []string) ([]*Card, error) {
	if len(issueIDs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)
	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card:      Card{},
			IssueIdIn: issueIDs,
		})).
		Scan(result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *CardDao) GetByIDIssueIDMerchantID(ctx context.Context, id uint64, issueID string, merchantID uint64) (*Card, error) {

	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:         id,
				IssueID:    issueID,
				MerchantID: merchantID,
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

func (ad *CardDao) GetByIDIssueID(ctx context.Context, id uint64, issueID string) (*Card, error) {

	if id == 0 || issueID == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:      id,
				IssueID: issueID,
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

func (ad *CardDao) GetByIssueIDDeleted(ctx context.Context, issueID string) (*Card, error) {

	if issueID == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				IssueID: issueID,
				Status:  common.CARD_STATUS_DELETED,
			},
			Deleted: true,
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

func (ad *CardDao) GetByWhaleUserID(ctx context.Context, whaleUserID uint64) (*Card, error) {

	if whaleUserID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				WhaleUserID: whaleUserID,
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

func (ad *CardDao) GetIncludeDeletedByWhaleUserID(ctx context.Context, whaleUserID uint64) (*Card, error) {

	if whaleUserID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				WhaleUserID: whaleUserID,
			},
			Deleted: true,
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

func (ad *CardDao) GetByWhaleCardID(ctx context.Context, whaleCardID uint64) (*Card, error) {

	if whaleCardID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				WhaleCardID: whaleCardID,
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

func (ad *CardDao) GetByPaycryptoCardNO(ctx context.Context, cardNO string) (*Card, error) {

	if cardNO == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				PaycryptoCardNO: cardNO,
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

func (ad *CardDao) GetByUserIDCategoryIDForUpdate(ctx context.Context, userID uint64, categoryID uint64) (*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				CategoryID: categoryID,
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

func (ad *CardDao) GetByUserIDCurrencyType(ctx context.Context, userID uint64, currency common.Currency, t common.AssetType) (*Card, error) {

	if userID == 0 || currency == 0 || t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:   userID,
				Currency: currency,
				Type:     t,
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

func (ad *CardDao) GetByUserIDCategoryID(ctx context.Context, userID uint64, categoryID uint64) (*Card, error) {

	if userID == 0 || categoryID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				CategoryID: categoryID,
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

func (ad *CardDao) GetByUserIDVendor(ctx context.Context, userID uint64, vendor common.CardProductVendor) (*Card, error) {

	if userID == 0 || vendor == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID: userID,
				Vendor: vendor,
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

func (ad *CardDao) GetByUserIDCategoryIDPaycryptoTypeID(ctx context.Context, userID uint64, categoryID uint64, paycryptoTypeID string) (*Card, error) {

	if userID == 0 || categoryID == 0 || paycryptoTypeID == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:          userID,
				CategoryID:      categoryID,
				PaycryptoTypeID: paycryptoTypeID,
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

func (ad *CardDao) GetByUserIDInCategoryID(ctx context.Context, userIDs []uint64, categoryID uint64) (*Card, error) {

	if len(userIDs) == 0 || categoryID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				CategoryID: categoryID,
			},
			UserIDIn: userIDs,
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

func (ad *CardDao) ListByMerchantID(ctx context.Context, merchantID uint64, assetType common.AssetType) ([]*Card, error) {

	if merchantID == 0 || assetType == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				MerchantID: merchantID,
				Type:       assetType,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByTypeUserID(ctx context.Context, t common.AssetType, userID uint64) ([]*Card, error) {

	if t == 0 || userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type:   t,
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByTypeUserIDIn(ctx context.Context, t common.AssetType, userIDs []uint64) ([]*Card, error) {

	if len(userIDs) == 0 {
		return []*Card{}, nil
	}

	if t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type: t,
			},
			UserIDIn: userIDs,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByIDInMerchantID(ctx context.Context, merchantID uint64, cardIDs []uint64) ([]*Card, error) {
	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				MerchantID: merchantID,
			},
			IDIn: cardIDs,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByIDInUserIDInMerchantID(ctx context.Context, merchantID uint64, cardIDs []uint64, userIDs []uint64) ([]*Card, error) {
	if merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				MerchantID: merchantID,
			},
			IDIn:     cardIDs,
			UserIDIn: userIDs,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByUserID(ctx context.Context, userID uint64) ([]*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID: userID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByUserIDTypeFormat(ctx context.Context, userID uint64, t common.AssetType, format common.CardFormat) ([]*Card, error) {

	if userID == 0 || t == 0 || format == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID: userID,
				Type:   t,
				Format: format,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByUserIDCurrencyType(ctx context.Context, userID uint64, currency common.Currency, types []common.AssetType) ([]*Card, error) {

	if userID == 0 || currency == 0 || len(types) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:   userID,
				Currency: currency,
			},
			TypeIn: types,
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByUserIDFromAutoTopUpType(ctx context.Context, userID uint64, fromAutoTopUp common.AutoTopUpStatus, t common.AssetType) ([]*Card, error) {

	if userID == 0 || fromAutoTopUp == 0 || t == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:        userID,
				FromAutoTopUp: fromAutoTopUp,
				Type:          t,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListByUserIDCatogoryID(ctx context.Context, userID uint64, categoryID uint64) ([]*Card, error) {

	if userID == 0 || categoryID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				CategoryID: categoryID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) GetByIDTypeUserIDCatogoryID(ctx context.Context, id uint64, t common.AssetType, userID uint64, categoryID uint64) (*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type:       t,
				ID:         id,
				UserID:     userID,
				CategoryID: categoryID,
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

func (ad *CardDao) GetByIDTypeUserIDsCatogoryID(ctx context.Context, id uint64, t common.AssetType, userIDs []uint64, categoryID uint64) (*Card, error) {

	if len(userIDs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type:       t,
				ID:         id,
				CategoryID: categoryID,
			},
			UserIDIn: userIDs,
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

func (ad *CardDao) GetByIDUserID(ctx context.Context, id uint64, userID uint64) (*Card, error) {
	if id == 0 || userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:     id,
				UserID: userID,
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

func (ad *CardDao) GetByIDUserIDIn(ctx context.Context, id uint64, userIDs []uint64) (*Card, error) {
	if id == 0 || len(userIDs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: id,
			},
			UserIDIn: userIDs,
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

func (ad *CardDao) GetByID(ctx context.Context, id uint64) (*Card, error) {
	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
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

func (ad *CardDao) GetByIDForUpdate(ctx context.Context, id uint64) (*Card, error) {
	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: id,
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

func (ad *CardDao) GetByIDMerchantID(ctx context.Context, id uint64, merchantID uint64) (*Card, error) {
	if id == 0 || merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:         id,
				MerchantID: merchantID,
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

func (ad *CardDao) GetByIDUserIDMerchantID(ctx context.Context, id uint64, userID uint64, merchantID uint64) (*Card, error) {
	if id == 0 || userID == 0 || merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:         id,
				UserID:     userID,
				MerchantID: merchantID,
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

func (ad *CardDao) GetsByFreezeStatusIDForShare(ctx context.Context, freezeStatus common.CardFreezeStatus, id uint64) (*Card, error) {
	if id == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID:           id,
				FreezeStatus: freezeStatus,
			},
			ForShare: true,
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

func (ad *CardDao) GetByUserIDCurrencyTypes(ctx context.Context, userID uint64, currency common.Currency, types []common.AssetType) (*Card, error) {
	if userID == 0 || currency == 0 || len(types) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:   userID,
				Currency: currency,
			},
			TypeIn: types,
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

func (ad *CardDao) GetByUserIDCurrencyTypesVendor(ctx context.Context, userID uint64, currency common.Currency, types []common.AssetType, vendor common.CardProductVendor) (*Card, error) {
	if userID == 0 || currency == 0 || len(types) == 0 || vendor == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:   userID,
				Currency: currency,
				Vendor:   vendor,
			},
			TypeIn: types,
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

func (ad *CardDao) ListByStatusBlockReasonBlockedAtBefore(ctx context.Context, status common.CardStatus, reason common.CardBlockReason, blockedAt time.Time) ([]*Card, error) {
	if reason == "" || blockedAt.IsZero() || blockedAt.UnixMilli() == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Status:      status,
				BlockReason: &reason,
			},
			BlockedAtLessThan: utils.DBQueryTime(blockedAt),
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) PageByUserIDMerchantIDCardIDCategoryID(ctx context.Context, userID uint64, cardID uint64, merchantID uint64, categoryID uint64, pageCurrent int, pageSize int) (records []*Card, current int, size int, total int, err error) {

	result := make([]*Card, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				ID:         cardID,
				MerchantID: merchantID,
				CategoryID: categoryID,
			},
		})).
		Count(&s).
		Scopes(ad.queryChain(&CardQuery{
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

func (ad *CardDao) PageByUserIDMerchantIDCardIDCurrency(ctx context.Context, userID uint64, cardID uint64, merchantID uint64, currency common.Currency, pageCurrent int, pageSize int) (records []*Card, current int, size int, total int, err error) {
	if userID == 0 && cardID == 0 && merchantID == 0 && currency == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				UserID:     userID,
				ID:         cardID,
				MerchantID: merchantID,
				Currency:   currency,
			},
		})).
		Count(&s).
		Scopes(ad.queryChain(&CardQuery{
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

func (ad *CardDao) PageByUserIDInType(ctx context.Context, userIDs []uint64, t common.AssetType, pageCurrent int, pageSize int) (records []*Card, current int, size int, total int, err error) {
	if len(userIDs) == 0 && t == 0 {
		return nil, 0, 0, 0, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	s := int64(0)
	db := ad.db.WithContext(ctx)

	err = db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type: t,
			},
			UserIDIn: userIDs,
		})).
		Count(&s).
		Scopes(ad.queryChain(&CardQuery{
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

func (ad *CardDao) ListReapCard(ctx context.Context) ([]*Card, error) {

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type:   common.ASSET_TYPE_CARD_PRODUCT,
				Vendor: common.CARD_PRODUCT_VENDOR_REAP,
				Status: common.CARD_STATUS_ACTIVATED,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListReapPhysicalCards(ctx context.Context) ([]*Card, error) {

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Type:   common.ASSET_TYPE_CARD_PRODUCT,
				Vendor: common.CARD_PRODUCT_VENDOR_REAP,
				Status: common.CARD_STATUS_NOT_ACTIVATED,
				Format: common.CARD_FORMAT_PHYSICAL,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) Get(ctx context.Context, query *CardQuery) (*Card, error) {
	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *CardDao) Gets(ctx context.Context, query *CardQuery) ([]*Card, error) {
	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListReapBlockCardByUserID(ctx context.Context, userID uint64) ([]*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	query := db.Model(Card{}).Where("user_id = ? AND status = ? AND vendor = ?", userID, common.CARD_STATUS_BLOCKED, common.CARD_PRODUCT_VENDOR_REAP)

	err := query.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListWhaleAndPayCryptoBlockCardByUserID(ctx context.Context, userID uint64) ([]*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	vendors := []common.CardProductVendor{
		common.CARD_PRODUCT_VENDOR_WHALE,
		common.CARD_PRODUCT_VENDOR_PAYCRYPTO,
	}

	query := db.Model(Card{}).Where("user_id = ? AND status = ? AND vendor in (?)", userID, common.CARD_STATUS_BLOCKED, vendors)

	err := query.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListCryptoBlockCardByUserID(ctx context.Context, userID uint64) ([]*Card, error) {

	if userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)

	query := db.Model(Card{}).Where("user_id = ? AND status = ? AND type = ?", userID, common.CARD_STATUS_BLOCKED, common.ASSET_TYPE_CRYPTO)

	err := query.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ad *CardDao) ListCards(ctx context.Context, ids []uint64) ([]*Card, error) {
	result := make([]*Card, 0)
	db := ad.db.WithContext(ctx)
	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{},
			IDIn: ids,
		})).
		Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*Card{}, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ad *CardDao) UpdateStatusByCardID(ctx context.Context, cardID uint64, status common.CardStatus) error {
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: cardID,
			},
		})).
		Updates(map[string]interface{}{
			"status":       status,
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

func (ad *CardDao) GetByCardIDDeleted(ctx context.Context, cardID uint64) (*Card, error) {

	result := &Card{}
	db := ad.db.WithContext(ctx)

	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				ID: cardID,
			},
			Deleted: true,
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

func (ad *CardDao) Save(ctx context.Context, model *Card) (uint64, error) {

	db := ad.db.WithContext(ctx)
	if model.ID == 0 {
		cardID, err := utils.RandomIDWithPrefix(model.CategoryID)
		if err != nil {
			return 0, err
		}
		model.ID = cardID
	}

	ret := db.Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (ad *CardDao) Update(ctx context.Context, query *CardQuery) (int64, error) {

	db := ad.db.WithContext(ctx)

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
			case reflect.TypeOf((*time.Time)(nil)):
				t := ptr.(*time.Time)
				if t.IsZero() {
					continue
				}
				attrs[fieldName] = t
			}
		default:
			continue
		}
	}

	ret := db.
		Model(Card{}).
		Scopes(ad.queryChain(query)).
		Updates(attrs)

	if ret.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	if ret.Error != nil {
		return 0, ret.Error
	}

	return ret.RowsAffected, nil
}

func (ad *CardDao) queryChain(query *CardQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.Card)
		structValue := reflect.ValueOf(query.Card)
		structPtrValue := reflect.ValueOf(&query.Card)
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
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Uint()))
			case reflect.Float32, reflect.Float64:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Float()))
			case reflect.Bool:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Bool()))
			case reflect.Pointer:
				db.Scopes(ad.equalScope(fieldName, structValue.Field(i).Interface()))
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
					db.Scopes(ad.equalScope(fieldName, value))
				}
			default:
				continue
			}
		}

		notEqualType := reflect.TypeOf(query.NotEqual)
		notEqualValue := reflect.ValueOf(query.NotEqual)
		for i := 0; i < notEqualType.NumField(); i++ {
			if notEqualValue.Field(i).IsZero() {
				continue
			}
			fieldName := stringy.New(notEqualType.Field(i).Name).SnakeCase().ToLower()
			settings := schema.ParseTagSetting(notEqualType.Field(i).Tag.Get("gorm"), ";")
			if f, ok := settings["COLUMN"]; ok {
				fieldName = f
			}
			db.Scopes(ad.notEqualScope(fieldName, notEqualValue.Field(i).Interface()))
		}

		if query.TypeIn != nil {
			db.Scopes(ad.inScope("type", query.TypeIn))
		}

		if query.IDIn != nil {
			db.Scopes(ad.inScope("id", query.IDIn))
		}

		if query.UserIDIn != nil {
			db.Scopes(ad.inScope("user_id", query.UserIDIn))
		}

		if query.AssetTypeIn != nil {
			db.Scopes(ad.inScope("type", query.AssetTypeIn))
		}

		if query.IssueIdIn != nil {
			db.Scopes(ad.inScope("issue_id", query.IssueIdIn))
		}

		if query.ForUpdate {
			db.Scopes(ad.forScope("UPDATE"))
		}

		if query.ForShare {
			db.Scopes(ad.forScope("SHARE"))
		}

		if !query.Deleted {
			db.Scopes(ad.nullScope([]string{"deleted_at"}, true))
		}

		if !query.BlockedAtLessThan.IsZero() && query.BlockedAtLessThan.UnixMilli() != 0 {
			db.Scopes(ad.compareScope("blocked_at", query.BlockedAtLessThan, false, true))
		}

		return db.
			Scopes(ad.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(ad.pageScope(query.Current, query.PageSize))
	}
}

func (ad *CardDao) GetCountByUserIdAndCurrency(ctx context.Context, userID uint64, currency common.Currency) (int64, error) {
	db := ad.db.WithContext(ctx)
	var count int64
	err := db.
		Model(Card{}).
		Scopes(ad.queryChain(&CardQuery{
			Card: Card{
				Currency: currency,
				UserID:   userID,
			},
		})).
		Count(&count).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ad *CardDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (ad *CardDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
	}
}

func (ad *CardDao) forScope(lock string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lock != "" {
			return db.Clauses(clause.Locking{Strength: lock})
		}
		return db
	}
}

func (ad *CardDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *CardDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (ad *CardDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (ad *CardDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && page != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}

func (ad *CardDao) notEqualScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` != ?", fieldName), field)
	}
}
