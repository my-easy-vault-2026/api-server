package financial

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

type SnapshotBalance struct {
	UserID         uint64               `json:"userId"`
	UserRole       common.Role          `json:"userRole"`
	UserKYCLevel   string               `json:"userKycLevel"`
	CardType       common.AssetType     `json:"cardType"`
	CardID         uint64               `json:"cardId"`
	CardCategoryID uint64               `json:"cardCategoryId"`
	CardCurrency   common.Currency      `json:"cardCurrency"`
	Balance        decimal.Decimal      `json:"balance"`
	FinancialCode  common.FinancialCode `json:"financialCode"`
	TakenAt        time.Time            `json:"takenAt"`
	EarningStatus  string               `json:"earningStatus"`
}

type SnapshotBalanceDao struct {
}

func NewSnapshotBalanceDao() *SnapshotBalanceDao {
	return &SnapshotBalanceDao{}
}

// Save stores a Preview struct in Redis.
func (sbd *SnapshotBalanceDao) Save(ctx context.Context, snapshotBalance *SnapshotBalance, expiration time.Duration) error {
	// Convert the Preview struct to JSON
	snapshotBalanceJSON, err := json.Marshal(snapshotBalance)
	if err != nil {
		return err
	}

	redisKey := utils.GetSnapshotRedisKey(snapshotBalance.FinancialCode, snapshotBalance.CardID, snapshotBalance.TakenAt)

	// Store the JSON data in Redis using the provided key
	err = utils.RDB.Set(ctx, redisKey, snapshotBalanceJSON, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a Preview struct from Redis by key.
func (sbd *SnapshotBalanceDao) Get(ctx context.Context, snapshotBalance *SnapshotBalance) (*SnapshotBalance, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetSnapshotRedisKey(snapshotBalance.FinancialCode, snapshotBalance.CardID, snapshotBalance.TakenAt)

	snapshotBalanceJSON, err := utils.RDB.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a Preview struct
	res := &SnapshotBalance{}
	err = json.Unmarshal([]byte(snapshotBalanceJSON), res)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Remove deletes a Preview struct from Redis by key.
func (sbd *SnapshotBalanceDao) Remove(ctx context.Context, snapshotBalance *SnapshotBalance) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetSnapshotRedisKey(snapshotBalance.FinancialCode, snapshotBalance.CardID, snapshotBalance.TakenAt)
	err := utils.RDB.Del(ctx, redisKey).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}
