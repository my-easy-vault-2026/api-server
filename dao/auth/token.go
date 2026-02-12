package auth

import (
	"context"
	"encoding/json"
	"errors"
	"shared-modules/common"
	"shared-modules/utils"
	"strconv"
	"time"

	"api-server/infra"
	"api-server/lib"

	"github.com/redis/go-redis/v9"
)

type Token struct {
	UserID    uint64      `json:"userId"`
	GroupIDs  []uint64    `json:"groupIds,omitempty"`
	Role      common.Role `json:"role,omitempty"`
	WsToken   string      `json:"wsToken,omitempty"`
	IssuedAt  time.Time   `json:"issuedAt"`
	ExpiredAt time.Time   `json:"expiredAt"`
}

type TokenDao struct {
	redis infra.Redis
	env   *lib.Env
}

func NewTokenDao(redis infra.Redis, env *lib.Env) *TokenDao {
	return &TokenDao{redis: redis, env: env}
}

// Save stores a Token struct in Redis.
func (pd *TokenDao) Save(ctx context.Context, key string, token *Token, expiration time.Duration, loginDataExpiration time.Duration) error {
	// Convert the Token struct to JSON
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	redisKey := utils.GetTokenRedisKey(key)

	// Store the JSON data in Redis using the provided key
	err = pd.redis.Set(ctx, redisKey, token.UserID, expiration).Err()
	if err != nil {
		return err
	}

	redisKey = utils.GetLoginDataRedisKey(token.UserID)

	// Store the JSON data in Redis using the provided key
	err = pd.redis.Set(ctx, redisKey, tokenJSON, loginDataExpiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a Token struct from Redis by key.
func (pd *TokenDao) Get(ctx context.Context, key string) (*Token, error) {
	// Retrieve the JSON data from Redis
	redisKey := utils.GetTokenRedisKey(key)

	userIDStr, err := pd.redis.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, err
	}

	redisKey = utils.GetLoginDataRedisKey(userID)
	previewJSON, err := pd.redis.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a Token struct
	preview := &Token{}
	err = json.Unmarshal([]byte(previewJSON), preview)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return preview, nil
}

// Remove deletes a Token struct from Redis by key.
func (pd *TokenDao) Remove(ctx context.Context, key string) error {
	// Delete the data from Redis using the provided key
	redisKey := utils.GetTokenRedisKey(key)
	err := pd.redis.Del(ctx, redisKey).Err()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}
