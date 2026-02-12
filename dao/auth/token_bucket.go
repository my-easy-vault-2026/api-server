package auth

import (
	"time"

	"github.com/my-easy-vault-2026/api-server/infra"
	"github.com/my-easy-vault-2026/api-server/lib"
)

type TokenBucketData struct {
	LastGenAt time.Time `json:"lastGenAt"`
}

type TokenBucketDao struct {
	redis infra.Redis
	env   *lib.Env
}

func NewTokenBucketDao(redis infra.Redis, env *lib.Env) *TokenBucketDao {
	return &TokenBucketDao{redis: redis, env: env}
}

// // Save stores a TokenBucket struct in Redis.
// func (pd *TokenBucketDao) Save(ctx context.Context, key string, token *TokenBucket, expiration time.Duration, loginDataExpiration time.Duration) error {
// 	// Convert the TokenBucket struct to JSON
// 	tokenJSON, err := json.Marshal(token)
// 	if err != nil {
// 		return err
// 	}

// 	redisKey := utils.GetTokenBucketRedisKey(key)

// 	// Store the JSON data in Redis using the provided key
// 	err = utils.RDB.Set(ctx, redisKey, token.UserID, expiration).Err()
// 	if err != nil {
// 		return err
// 	}

// 	redisKey = utils.GetLoginDataRedisKey(token.UserID)

// 	// Store the JSON data in Redis using the provided key
// 	err = utils.RDB.Set(ctx, redisKey, tokenJSON, loginDataExpiration).Err()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // Get retrieves a TokenBucket struct from Redis by key.
// func (pd *TokenBucketDao) Get(ctx context.Context, key string) (*TokenBucket, error) {
// 	// Retrieve the JSON data from Redis
// 	redisKey := utils.GetTokenBucketRedisKey(key)

// 	userIDStr, err := utils.RDB.Get(ctx, redisKey).Result()
// 	if errors.Is(err, redis.Nil) {
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	userID, err := strconv.ParseUint(userIDStr, 10, 64)
// 	if err != nil {
// 		return nil, err
// 	}

// 	redisKey = utils.GetLoginDataRedisKey(userID)
// 	previewJSON, err := utils.RDB.Get(ctx, redisKey).Result()
// 	if errors.Is(err, redis.Nil) {
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Unmarshal the JSON data into a TokenBucket struct
// 	preview := &TokenBucket{}
// 	err = json.Unmarshal([]byte(previewJSON), preview)
// 	if errors.Is(err, redis.Nil) {
// 		return nil, nil
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return preview, nil
// }

// // Remove deletes a TokenBucket struct from Redis by key.
// func (pd *TokenBucketDao) Remove(ctx context.Context, key string) error {
// 	// Delete the data from Redis using the provided key
// 	redisKey := utils.GetTokenBucketRedisKey(key)
// 	err := utils.RDB.Del(ctx, redisKey).Err()
// 	if errors.Is(err, redis.Nil) {
// 		return nil
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
