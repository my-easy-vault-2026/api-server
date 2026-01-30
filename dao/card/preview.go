package card

import (
	"context"
	"encoding/json"
	"errors"
	"shared-modules/common"
	"shared-modules/utils"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shopspring/decimal"
)

type ExchangeRate struct {
	BaseCurrency  common.Currency `json:"baseCurrency"`
	QuoteCurrency common.Currency `json:"quoteCurrency"`
	Rate          decimal.Decimal `json:"rate"`
	Timestamp     time.Time       `json:"timestamp"`
}

type Preview struct {
	UserID                uint64            `json:"userID"`
	DepositAmount         decimal.Decimal   `json:"depositAmount"`
	DepositAmountShortage *decimal.Decimal  `json:"depositAmountShortage,omitempty"`
	FromCardID            uint64            `json:"fromCardID"`
	FromCategoryID        uint64            `json:"fromCategoryID"`
	FromCurrency          common.Currency   `json:"fromCurrency"`
	ToCategoryID          uint64            `json:"toCategoryID"`
	ToCurrency            common.Currency   `json:"toCurrency"`
	ReceiveAmount         decimal.Decimal   `json:"receiveAmount"`
	FromDiscount          *decimal.Decimal  `json:"fromDiscount,omitempty"` // 幣種為來源幣種
	ToDiscount            *decimal.Decimal  `json:"toDiscount,omitempty"`   // 幣種為到帳幣種
	Bonus                 *decimal.Decimal  `json:"bonus,omitempty"`        // 幣種為目的幣種
	PromotionCode         string            `json:"promotionCode,omitempty"`
	Fee                   *decimal.Decimal  `json:"fee"`
	CardFee               *decimal.Decimal  `json:"cardFee"`
	Rate                  []*ExchangeRate   `json:"rate"`
	DisplayRate           *decimal.Decimal  `json:"displayRate"`
	AddressLine1          string            `json:"addressLine1"`
	AddressLine2          string            `json:"addressLine2"`
	AddressLine3          string            `json:"addressLine3"`
	AddressLine4          string            `json:"addressLine4"`
	AddressLine5          string            `json:"addressLine5"`
	State                 string            `json:"state"`
	NationCode            common.NationCode `json:"NationCode"`
	PostalCode            string            `json:"postalCode"`
	City                  string            `json:"city"`
	ExpiredAt             time.Time         `json:"expiredAt"`
	UserTempAddressID     *uint64           `json:"userTempAddressID"` //用戶臨時地址
}

type PreviewDao struct {
}

func NewPreviewDao() *PreviewDao {
	return &PreviewDao{}
}

// Save stores a Preview struct in Redis.
func (pd *PreviewDao) Save(ctx context.Context, key string, preview *Preview, expiration time.Duration) error {
	// Convert the Preview struct to JSON
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return err
	}

	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)

	// Store the JSON data in Redis using the provided key
	err = utils.RDB.Set(ctx, redisKey, previewJSON, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Save stores a Preview struct in Redis.
func (pd *PreviewDao) SaveDapp(ctx context.Context, key, address string, preview *Preview, expiration time.Duration) error {
	// Convert the Preview struct to JSON
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return err
	}

	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)

	// Store the JSON data in Redis using the provided key
	err = utils.RDB.HSet(ctx, redisKey, address, previewJSON).Err()
	if err != nil {
		return err
	}
	utils.RDB.Expire(ctx, redisKey, expiration)

	return nil
}

// Get retrieves a Preview struct from Redis by key.
func (pd *PreviewDao) Get(ctx context.Context, key string) (*Preview, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)

	previewJSON, err := utils.RDB.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a Preview struct
	preview := &Preview{}
	err = json.Unmarshal([]byte(previewJSON), preview)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return preview, nil
}

// Get retrieves a Preview struct from Redis by key.
func (pd *PreviewDao) GetDapp(ctx context.Context, key, address string) (*Preview, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)

	previewJSON, err := utils.RDB.HGet(ctx, redisKey, address).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a Preview struct
	preview := &Preview{}
	err = json.Unmarshal([]byte(previewJSON), preview)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return preview, nil
}

// Remove deletes a Preview struct from Redis by key.
func (pd *PreviewDao) Remove(ctx context.Context, key string) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)
	err := utils.RDB.Del(ctx, redisKey).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

// Remove deletes a Preview struct from Redis by key.
func (pd *PreviewDao) RemoveDapp(ctx context.Context, key, address string) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_APPLY, key)
	err := utils.RDB.HDel(ctx, redisKey, address).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}
