package order

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

type TransactionRecordDao struct {
	db *gorm.DB
}

func NewTransactionRecordDao() *TransactionRecordDao {
	return &TransactionRecordDao{
		db: utils.DB,
	}
}

type TransactionRecord struct {
	ID                               uint64
	Type                             common.TransactionRecordType
	TransactionNO                    string
	ParentTransactionNO              string `gorm:"default:null"`
	TransferTransactionNO            string `gorm:"default:null"`
	UserID                           uint64
	MerchantID                       uint64 `gorm:"default:null"`
	CardID                           uint64
	CentralCardID                    uint64 `gorm:"default:null"`
	DelegatedID                      uint64 `gorm:"default:null"`
	Income                           decimal.NullDecimal
	IncomeCategoryID                 uint64
	IncomeCurrency                   common.Currency
	AgainstIncome                    decimal.NullDecimal        `gorm:"default:null"`
	AgainstIncomeCategoryID          uint64                     `gorm:"default:null"`
	AgainstIncomeCurrency            common.Currency            `gorm:"default:null"`
	AgainstIncomeCurrencyOriginal    string                     `gorm:"default:null"`
	Side                             common.TransactionSide     `gorm:"default:null"`
	FromType                         common.AssetType           `gorm:"default:null"`
	FromCardID                       uint64                     `gorm:"default:null"`
	FromCategoryID                   uint64                     `gorm:"default:null"`
	FromCurrency                     common.Currency            `gorm:"default:null"`
	FromAmount                       decimal.NullDecimal        `gorm:"default:null"`
	FromDiscount                     decimal.NullDecimal        `gorm:"default:null"`
	FromUserID                       uint64                     `gorm:"default:null"`
	FromAddress                      string                     `gorm:"default:null"`
	FromEmail                        string                     `gorm:"default:null"`
	FromMerchantID                   uint64                     `gorm:"default:null"`
	FromBalanceType                  common.BalanceType         `gorm:"default:null"`
	FromAlias                        string                     `gorm:"default:null"`
	FromPANNumber                    string                     `gorm:"column:from_pan_number;default:null"`
	FromRole                         common.Role                `gorm:"default:null"`
	ToType                           common.AssetType           `gorm:"default:null"`
	ToCardID                         uint64                     `gorm:"default:null"`
	ToCategoryID                     uint64                     `gorm:"default:null"`
	ToCurrency                       common.Currency            `gorm:"default:null"`
	ToAmount                         decimal.NullDecimal        `gorm:"default:null"`
	ToBonus                          decimal.NullDecimal        `gorm:"default:null"`
	ToUserID                         uint64                     `gorm:"default:null"`
	ToAddress                        string                     `gorm:"default:null"`
	ToEmail                          string                     `gorm:"default:null"`
	ToMerchantID                     uint64                     `gorm:"default:null"`
	ToBalanceType                    common.BalanceType         `gorm:"default:null"`
	ToAlias                          string                     `gorm:"default:null"`
	ToPANNumber                      string                     `gorm:"column:to_pan_number;default:null"`
	ToRole                           common.Role                `gorm:"default:null"`
	ToProductID                      uint64                     `gorm:"default:null"`
	TransferToType                   common.AssetType           `gorm:"default:null"`
	TransferToCardID                 uint64                     `gorm:"default:null"`
	TransferToCategoryID             uint64                     `gorm:"default:null"`
	TransferToCurrency               common.Currency            `gorm:"default:null"`
	TransferToUserID                 uint64                     `gorm:"default:null"`
	TransferToAlias                  string                     `gorm:"default:null"`
	TransferToPANNumber              string                     `gorm:"column:transfer_to_pan_number;default:null"`
	TXHash                           string                     `gorm:"column:tx_hash;default:null"`
	Mainnet                          common.Mainnet             `gorm:"default:null"`
	Protocol                         common.Protocol            `gorm:"default:null"`
	TransferChannel                  common.TransferChannel     `gorm:"default:null"`
	PointSource                      common.PointSource         `gorm:"default:null"`
	SourceOrderNO                    string                     `gorm:"default:null"`
	FinancialCode                    common.FinancialCode       `gorm:"default:null"`
	ReapMerchantID                   uint64                     `gorm:"default:null"`
	MerchantMCCCode                  string                     `gorm:"default:null"`
	MerchantMCCCategory              string                     `gorm:"default:null"`
	MerchantCity                     string                     `gorm:"default:null"`
	MerchantName                     string                     `gorm:"default:null"`
	MerchantState                    string                     `gorm:"default:null"`
	MerchantCountry                  string                     `gorm:"default:null"`
	MerchantPostCode                 string                     `gorm:"default:null"`
	ReapTransactionID                string                     `gorm:"default:null"`
	ParentTransactionID              string                     `gorm:"default:null"`
	ReapChannel                      common.ReapChannel         `gorm:"default:null"`
	WhaleTransactionID               string                     `gorm:"default:null"`
	WhaleDetail                      string                     `gorm:"default:null"`
	PaycryptoTransactionID           string                     `gorm:"default:null"`
	PaycryptoOriginalTransactionID   string                     `gorm:"default:null"`
	ClearAmount                      decimal.NullDecimal        `gorm:"default:null"`
	ReversalAmount                   decimal.NullDecimal        `gorm:"default:null"`
	ReversalTransactionAmount        decimal.NullDecimal        `gorm:"default:null"`
	ReversalExchangeRate             decimal.NullDecimal        `gorm:"default:null"`
	RefundAmount                     decimal.NullDecimal        `gorm:"default:null"`
	RefundTransactionAmount          decimal.NullDecimal        `gorm:"default:null"`
	RefundExchangeRate               decimal.NullDecimal        `gorm:"default:null"`
	DisplayRate                      *decimal.Decimal           `gorm:"default:null"`
	YieldRate                        *decimal.Decimal           `gorm:"default:null"`
	ExchangeRate                     *decimal.Decimal           `gorm:"default:null"`
	ExchangeFee                      *decimal.Decimal           `gorm:"default:null"`
	ExchangeFeeCurrency              common.Currency            `gorm:"default:null"`
	DepositFee                       *decimal.Decimal           `gorm:"default:null"`
	DepositFeeCurrency               common.Currency            `gorm:"default:null"`
	TransferFee                      *decimal.Decimal           `gorm:"default:null"`
	TransferFeeCurrency              common.Currency            `gorm:"default:null"`
	WithdrawFee                      *decimal.Decimal           `gorm:"default:null"`
	WithdrawFeeCurrency              common.Currency            `gorm:"default:null"`
	CardFee                          *decimal.Decimal           `gorm:"default:null"`
	CardFeeCurrency                  common.Currency            `gorm:"default:null"`
	PhysicalCardFee                  *decimal.Decimal           `gorm:"default:null"`
	PhysicalCardFeeCurrency          common.Currency            `gorm:"default:null"`
	DeliveryFee                      *decimal.Decimal           `gorm:"default:null"`
	DeliveryFeeCurrency              common.Currency            `gorm:"default:null"`
	TopUpFee                         *decimal.Decimal           `gorm:"default:null"`
	TopUpFeeCurrency                 common.Currency            `gorm:"default:null"`
	TopDownFee                       *decimal.Decimal           `gorm:"default:null"`
	TopDownFeeCurrency               common.Currency            `gorm:"default:null"`
	CardToCardFee                    *decimal.Decimal           `gorm:"default:null"`
	CardToCardFeeCurrency            common.Currency            `gorm:"default:null"`
	MerchantAdjustBalanceFee         *decimal.Decimal           `gorm:"default:null"`
	MerchantAdjustBalanceFeeCurrency common.Currency            `gorm:"default:null"`
	CPExpressFee                     *decimal.Decimal           `gorm:"column:cp_express_fee;default:null"`
	CPExpressFeeCurrency             common.Currency            `gorm:"column:cp_express_fee_currency;default:null"`
	ATMFee                           *decimal.Decimal           `gorm:"column:atm_fee;default:null"`
	ATMFeeCurrency                   common.Currency            `gorm:"column:atm_fee_currency;default:null"`
	FXFee                            *decimal.Decimal           `gorm:"column:fx_fee;default:null"`
	FXFeeCurrency                    common.Currency            `gorm:"column:fx_fee_currency;default:null"`
	ReapAuthorizationFee             *decimal.Decimal           `gorm:"default:null"`
	ReapAuthorizationFeeCurrency     common.Currency            `gorm:"default:null"`
	WhaleFee                         *decimal.Decimal           `gorm:"default:null"`
	WhaleFeeCurrency                 common.Currency            `gorm:"default:null"`
	PaycryptoChannel                 common.PaycryptoChannel    `gorm:"default:null"`
	PaycryptoFee                     *decimal.Decimal           `gorm:"default:null"`
	PaycryptoFeeCurrency             common.Currency            `gorm:"default:null"`
	PromotionCode                    string                     `gorm:"default:null"`
	PromotionType                    common.PromotionType       `gorm:"default:null"`
	ExecutorRole                     common.Role                `gorm:"default:null"`
	ChargePurpose                    common.SystemChargePurpose `gorm:"default:null"`
	Status                           common.TransactionStatus
	FailReason                       string              `gorm:"default:null"`
	ResponseCode                     string              `gorm:"default:null"`
	ReapDataLoss                     common.ReapDataLoss `gorm:"default:null"`
	ReversalAt                       *time.Time          `gorm:"default:null"`
	ClearedAt                        *time.Time          `gorm:"default:null"`
	RefundedAt                       *time.Time          `gorm:"default:null"`
	CoinsdoRepliedAt                 *time.Time          `gorm:"default:null"`
	CreatedAt                        time.Time           `gorm:"default:null"`
	UpdatedAt                        time.Time           `gorm:"default:null;autoUpdateTime:false"`
	Remark                           *string             `gorm:"default:null"`
	UserTempAddressID                *uint64             `gorm:"default:null"`
	EtherfiTransactionID             string              `gorm:"default:null"`
}

type TransactionRecordQuery struct {
	TransactionRecord
	Attrs                   TransactionRecord
	StatusIn                []common.TransactionStatus
	ForUpdate               bool
	ForShare                bool
	OrderBy                 string
	OrderDirection          common.OrderDirection
	UserIDIn                []uint64
	TransactionNOIn         []string
	TransactionRecordTypeIn []common.TransactionRecordType
	CreatedAtFrom           time.Time
	CreatedAtTo             time.Time
	IsNull                  []string
	utils.Page
}

func (TransactionRecord) TableName() string {
	return "transaction_record"
}

func (trd *TransactionRecordDao) PageByUserIDInCardIDType(ctx context.Context, userIDs []uint64, cardID uint64, typeIn []common.TransactionRecordType, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				CardID: cardID,
			},
			UserIDIn:                userIDs,
			TransactionRecordTypeIn: typeIn,
			IsNull:                  []string{"parent_transaction_id"},
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) PageByUserIDInCategoryIDType(ctx context.Context, userIDs []uint64, categoryID uint64, typeIn []common.TransactionRecordType, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				IncomeCategoryID: categoryID,
			},
			TransactionRecordTypeIn: typeIn,
			UserIDIn:                userIDs,
			IsNull:                  []string{"parent_transaction_id"},
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) PageByUserIDCardIDCategoryID(ctx context.Context, userID uint64, cardID uint64, categoryID uint64, createdAtFrom time.Time, createdAtTo time.Time, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:           userID,
				CardID:           cardID,
				IncomeCategoryID: categoryID,
			},
			CreatedAtFrom: createdAtFrom,
			CreatedAtTo:   createdAtTo,
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) PageByUserIDCardIDCategoryIDMerchantID(ctx context.Context, userID uint64, cardID uint64, categoryID uint64, merchantID uint64, createdAtFrom time.Time, createdAtTo time.Time, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:           userID,
				CardID:           cardID,
				MerchantID:       merchantID,
				IncomeCategoryID: categoryID,
			},
			CreatedAtFrom: createdAtFrom,
			CreatedAtTo:   createdAtTo,
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) PageByUserIDAndPromotionCode(ctx context.Context, userID uint64, promotionCode string, createdAtFrom time.Time, createdAtTo time.Time, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:        userID,
				PromotionCode: promotionCode,
			},
			CreatedAtFrom: createdAtFrom,
			CreatedAtTo:   createdAtTo,
			IsNull:        []string{"parent_transaction_id"},
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) GetByUserIDAndPromotionCode(ctx context.Context, userID uint64, promotionCode string, createdAtFrom time.Time, createdAtTo time.Time) (records []*TransactionRecord, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:        userID,
				PromotionCode: promotionCode,
			},
			CreatedAtFrom: createdAtFrom,
			CreatedAtTo:   createdAtTo,
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) GetByPromotionCodeAndDate(ctx context.Context, promotionCode string, createdAtFrom time.Time, createdAtTo time.Time) (records []*TransactionRecord, err error) {
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				PromotionCode: promotionCode,
				Side:          common.TRANSACTION_SIDE_FROM,
			},
			CreatedAtFrom: createdAtFrom,
			CreatedAtTo:   createdAtTo,
			IsNull:        []string{"parent_transaction_id"},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) GetByPromotionCode(ctx context.Context, promotionCode string) (records []*TransactionRecord, err error) {
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				PromotionCode: promotionCode,
			},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) GetByTypeAndPromotionCode(ctx context.Context, userID uint64, txType common.TransactionRecordType, promotionCode string) (records []*TransactionRecord, err error) {
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:        userID,
				PromotionCode: promotionCode,
				Type:          txType,
			},
		})).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) PageByTypeAndPromotionCode(ctx context.Context, userID uint64, txType common.TransactionRecordType, promotionCode string, pageCurrent int, pageSize int) (records []*TransactionRecord, current int, size int, total int, err error) {
	result := make([]*TransactionRecord, 0)
	s := int64(0)
	db := utils.GetDB(ctx)

	err = db.
		Model(TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:        userID,
				PromotionCode: promotionCode,
				Type:          txType,
			},
			IsNull: []string{"parent_transaction_id"},
		})).
		Count(&s).
		Scopes(trd.queryChain(&TransactionRecordQuery{
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

func (trd *TransactionRecordDao) Get(ctx context.Context, query *TransactionRecordQuery) (*TransactionRecord, error) {
	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(query)).
		First(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) Gets(ctx context.Context, query *TransactionRecordQuery) ([]*TransactionRecord, error) {
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*TransactionRecord, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (trd *TransactionRecordDao) GetByReapTransactionID(ctx context.Context, reapTransactionID string) (*TransactionRecord, error) {
	if reapTransactionID == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				ReapTransactionID: reapTransactionID,
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

func (trd *TransactionRecordDao) GetByTransactionNO(ctx context.Context, transactionNO string, side common.TransactionSide, userID uint64) (*TransactionRecord, error) {
	if transactionNO == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID:        userID,
				TransactionNO: transactionNO,
				Side:          side,
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

func (trd *TransactionRecordDao) GetByTransactionNOSideCardID(ctx context.Context, transactionNO string, side common.TransactionSide, cardID uint64, userIDs []uint64) (*TransactionRecord, error) {
	if transactionNO == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				TransactionNO: transactionNO,
				Side:          side,
				CardID:        cardID,
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

func (trd *TransactionRecordDao) GetByTransactionNOSide(ctx context.Context, transactionNO string, side common.TransactionSide) (*TransactionRecord, error) {
	if transactionNO == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				TransactionNO: transactionNO,
				Side:          side,
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

func (trd *TransactionRecordDao) GetByTransactionNOUserID(ctx context.Context, transactionNO string, userID uint64) (*TransactionRecord, error) {
	if transactionNO == "" || userID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				TransactionNO: transactionNO,
				UserID:        userID,
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

func (trd *TransactionRecordDao) GetByTransactionNOMerchantID(ctx context.Context, transactionNO string, merchantID uint64) (*TransactionRecord, error) {
	if transactionNO == "" || merchantID == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := &TransactionRecord{}
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				TransactionNO: transactionNO,
				MerchantID:    merchantID,
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

func (trd *TransactionRecordDao) GetBYEtherfiTransacionId(ctx context.Context, transactionNO string) (*TransactionRecord, error) {
	if transactionNO == "" {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := &TransactionRecord{}
	db := utils.GetDB(ctx)
	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				EtherfiTransactionID: transactionNO,
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

func (trd *TransactionRecordDao) ListByReapTransactionID(ctx context.Context, reapTransactionID string) ([]*TransactionRecord, error) {
	if reapTransactionID == "" {
		return make([]*TransactionRecord, 0), utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}

	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)

	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				ReapTransactionID: reapTransactionID,
			},
		})).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*TransactionRecord, 0), nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (trd *TransactionRecordDao) ListByTransactionNos(ctx context.Context, transactionNOs []string) ([]*TransactionRecord, error) {
	if len(transactionNOs) == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)
	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{},
			TransactionNOIn:   transactionNOs,
		})).
		Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*TransactionRecord, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) ListByUserIDCardIdTypeStatus(ctx context.Context, userId uint64, cardId uint64, txType common.TransactionRecordType, status common.TransactionStatus) ([]*TransactionRecord, error) {
	if userId == 0 || txType == 0 || status == 0 {
		return nil, utils.NewBusinessError(ctx, common.CODE_INVALID_PARAMETER)
	}
	result := make([]*TransactionRecord, 0)
	db := utils.GetDB(ctx)
	err := db.
		Model(&TransactionRecord{}).
		Scopes(trd.queryChain(&TransactionRecordQuery{
			TransactionRecord: TransactionRecord{
				UserID: userId,
				CardID: cardId,
				Type:   txType,
				Status: status,
			},
			OrderBy:        "created_at",
			OrderDirection: common.ORDER_DIRECTION_DESC,
		})).
		Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make([]*TransactionRecord, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (trd *TransactionRecordDao) Save(ctx context.Context, model *TransactionRecord) (uint64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(TransactionRecord{}).
		Create(model)

	if ret.Error != nil {
		return 0, ret.Error
	}
	return model.ID, nil
}

func (trd *TransactionRecordDao) Saves(ctx context.Context, models []*TransactionRecord) (int64, error) {

	db := utils.GetDB(ctx)

	ret := db.
		Model(TransactionRecord{}).
		CreateInBatches(models, len(models))

	if ret.Error != nil {
		return ret.RowsAffected, ret.Error
	}
	return ret.RowsAffected, nil
}

func (trd *TransactionRecordDao) Update(ctx context.Context, query *TransactionRecordQuery) (int64, error) {

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
		Model(&TransactionRecord{}).
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

func (trd *TransactionRecordDao) queryChain(query *TransactionRecordQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		structType := reflect.TypeOf(query.TransactionRecord)
		structValue := reflect.ValueOf(query.TransactionRecord)
		structPtrValue := reflect.ValueOf(&query.TransactionRecord)
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

		if query.UserIDIn != nil {
			db.Scopes(trd.inScope("user_id", query.UserIDIn))
		}

		if query.TransactionNOIn != nil {
			db.Scopes(trd.inScope("transaction_no", query.TransactionNOIn))
		}

		if query.TransactionRecordTypeIn != nil {
			db.Scopes(trd.inScope("type", query.TransactionRecordTypeIn))
		}

		if !query.CreatedAtFrom.IsZero() && query.CreatedAtFrom.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtFrom, true, true))
		}

		if !query.CreatedAtTo.IsZero() && query.CreatedAtTo.Unix() != 0 {
			db.Scopes(trd.compareScope("created_at", query.CreatedAtTo, false, true))
		}

		if query.IsNull != nil {
			db.Scopes(trd.nullScope(query.IsNull, true))
		}

		return db.
			Scopes(trd.orderByScope(query.OrderBy, query.OrderDirection)).
			Scopes(trd.pageScope(query.Current, query.PageSize))
	}
}

func (trd *TransactionRecordDao) nullScope(fieldNames []string, isNull bool) func(db *gorm.DB) *gorm.DB {
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

func (trd *TransactionRecordDao) equalScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("`%s` = ? ", fieldName), field)
	}
}

func (trd *TransactionRecordDao) inScope(fieldName string, field interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if reflect.TypeOf(field).Kind() == reflect.Slice {
			return db.Where(fmt.Sprintf("`%s` IN ? ", fieldName), field)
		}
		return db
	}
}

func (trd *TransactionRecordDao) compareScope(fieldName string, field interface{}, greater bool, equal bool) func(db *gorm.DB) *gorm.DB {
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

func (trd *TransactionRecordDao) orderByScope(fieldName string, direction common.OrderDirection) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if fieldName != "" && direction != 0 {
			return db.Order(fmt.Sprintf("`%s` %s", fieldName, direction.String()))
		}
		return db
	}
}

func (trd *TransactionRecordDao) pageScope(page int, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page != 0 && size != 0 {
			offset := size * (page - 1)
			return db.Limit(size).Offset(offset)

		}
		return db
	}
}
