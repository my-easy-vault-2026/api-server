package order

import (
	"api-server/infra"
	"api-server/lib"
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
	ExchangeFee    *decimal.Decimal       `json:"exchangeFee,omitempty"`
	TransferFee    *decimal.Decimal       `json:"transferFee,omitempty"`
	TopUpFee       *decimal.Decimal       `json:"topUpFee,omitempty"`
	TopDownFee     *decimal.Decimal       `json:"topDownFee,omitempty"`
	Rate           []*common.ExchangeRate `json:"rate,omitempty"`
	DisplayRate    *decimal.Decimal       `json:"displayRate"`
	ExpiredAt      time.Time              `json:"expiredAt"`
}

type PreviewDao struct {
	redis infra.Redis
	env   *lib.Env
}

func NewPreviewDao(redis infra.Redis, env *lib.Env) *PreviewDao {
	return &PreviewDao{redis: redis, env: env}
}

// Save stores a Preview struct in Redis.
func (pd *PreviewDao) Save(ctx context.Context, purpose common.PreviewPurpose, key string, preview *Preview, expiration time.Duration) error {
	// Convert the Preview struct to JSON
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return err
	}

	redisKey := utils.GetPreviewRedisKey(purpose, key)

	// Store the JSON data in Redis using the provided key
	err = pd.redis.Set(ctx, redisKey, previewJSON, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a Preview struct from Redis by key.
func (pd *PreviewDao) Get(ctx context.Context, purpose common.PreviewPurpose, key string) (*Preview, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetPreviewRedisKey(purpose, key)

	previewJSON, err := pd.redis.Get(ctx, redisKey).Result()
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
func (pd *PreviewDao) Remove(ctx context.Context, purpose common.PreviewPurpose, key string) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetPreviewRedisKey(purpose, key)
	err := pd.redis.Del(ctx, redisKey).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}
