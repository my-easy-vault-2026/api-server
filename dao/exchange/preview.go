package topUp

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

type Preview struct {
	FromAmount     decimal.Decimal        `json:"depositAmount"`
	UserID         uint64                 `json:"userId"`
	FromCardID     uint64                 `json:"fromCardId"`
	FromCategoryID uint64                 `json:"fromCategoryId"`
	FromCurrency   common.Currency        `json:"fromCurrency"`
	ToAmount       decimal.Decimal        `json:"receiveAmount"`
	ToCardID       uint64                 `json:"toCardId"`
	ToCategoryID   uint64                 `json:"toCategoryId"`
	ToCurrency     common.Currency        `json:"toCurrency"`
	ExchangeFee    decimal.Decimal        `json:"exchangeFee"`
	Rate           []*common.ExchangeRate `json:"rate,omitempty"`
	DisplayRate    *decimal.Decimal       `json:"displayRate"`
	ExpiredAt      time.Time              `json:"expiredAt"`
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

	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_EXCHANGE, key)

	// Store the JSON data in Redis using the provided key
	err = utils.RDB.Set(ctx, redisKey, previewJSON, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a Preview struct from Redis by key.
func (pd *PreviewDao) Get(ctx context.Context, key string) (*Preview, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_EXCHANGE, key)

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

// Remove deletes a Preview struct from Redis by key.
func (pd *PreviewDao) Remove(ctx context.Context, key string) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetPreviewRedisKey(common.PREVIEW_PURPOSE_EXCHANGE, key)
	err := utils.RDB.Del(ctx, redisKey).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}
