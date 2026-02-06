package infra

import (
	"fmt"
	"shared-modules/utils"

	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type L2CacheLevel int // 二級緩存等級
const (
	L2_CACHE_LEVEL_MEMORY L2CacheLevel = 1
	L2_CACHE_LEVEL_REDIS  L2CacheLevel = 2
)

type L2CacheConfig struct {
	Level         []L2CacheLevel
	ExpireSeconds time.Duration
}

// 僅支援select單表查詢
func L2CQuery[T any](ctx context.Context, db *gorm.DB, rd Redis, qf func(tx *gorm.DB) *gorm.DB, rf func(tx *gorm.DB) (T, error), config *L2CacheConfig) (T, error) {

	if config == nil {
		return *new(T), errors.New("missing parameter")
	}

	sql := db.ToSQL(qf)
	sql = strings.ReplaceAll(sql, "`", "")
	split := strings.Split(sql, " ")
	table := "unknown"
	for i, s := range split {
		if strings.ToLower(s) == "from" {
			if len(split) > i+1 {
				table = split[i+1]
				break
			}
		}
	}

	switch ret := rd.Get(ctx, utils.GetL2CacheKey(table, sql)); true {
	case ret.Err() != nil && !errors.Is(ret.Err(), redis.Nil):
		return *new(T), ret.Err()
	case ret.Err() == nil && ret.Val() != "":
		t := *new(T)
		err := json.Unmarshal([]byte(ret.Val()), &t)
		if err == nil {
			return t, nil
		}
		fmt.Printf("unmarshal failed [%s][%s], %#v", utils.GetL2CacheKey(table, sql), ret.Val(), err)
	case ret.Err() == nil && ret.Val() == "":
		return *new(T), nil
	case errors.Is(ret.Err(), redis.Nil):
	default:
		fmt.Printf("redis get failed [%s], %#v", utils.GetL2CacheKey(table, sql), ret.Err())
	}

	ret := qf(db)

	result, err := rf(ret)
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return *new(T), errors.New("invalid parameter")
	}
	resultCopy := deepCopy(result)
	go func(res T) {
		value := ""
		if err == nil && !reflect.ValueOf(res).IsZero() {
			resJSON, err := json.Marshal(res)
			if err != nil {
				fmt.Printf("marshal failed: [%s], ", utils.GetL2CacheKey(table, sql), err)
				return
			}
			value = string(resJSON)
		}

		if err := rd.SetEx(ctx, utils.GetL2CacheKey(table, sql), value, config.ExpireSeconds*time.Second).Err(); err != nil {
			fmt.Printf("redis cache failed: [%s], ", utils.GetL2CacheKey(table, sql), err)
			return
		}

		if err := rd.SAdd(ctx, utils.GetL2CacheKeysKey(table), utils.GetL2CacheKey(table, sql)).Err(); err != nil {
			fmt.Printf("redis sadd failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
			return
		}

		switch ret := rd.TTL(ctx, utils.GetL2CacheKeysKey(table)); true {
		case ret.Err() != nil && !errors.Is(ret.Err(), redis.Nil):
			fmt.Printf("redis ttl failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
			return
		case errors.Is(ret.Err(), redis.Nil):
			fmt.Printf("redis expire not found: [%#v]", utils.GetL2CacheKeysKey(table))
			if err := rd.SAdd(ctx, utils.GetL2CacheKeysKey(table), utils.GetL2CacheKey(table, sql)).Err(); err != nil {
				fmt.Printf("redis sadd failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
				return
			}
			if err := rd.Expire(ctx, utils.GetL2CacheKeysKey(table), config.ExpireSeconds*time.Second).Err(); err != nil {
				fmt.Printf("redis expire failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
				return
			}
		case ret.Val() <= config.ExpireSeconds*time.Second:
			if err := rd.Expire(ctx, utils.GetL2CacheKeysKey(table), config.ExpireSeconds*time.Second).Err(); err != nil {
				fmt.Printf("redis expire failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
				return
			}
		default:
			fmt.Printf("redis expire unknown condition: [%s][%#v]", utils.GetL2CacheKeysKey(table), ret)
		}

		if err := rd.SAdd(ctx, utils.GetL2CacheTablesKey(), table).Err(); err != nil {
			fmt.Printf("redis sadd failed: [%s], %v", utils.GetL2CacheKeysKey(table), err)
			return
		}

	}(resultCopy)
	return result, nil
}

func deepCopy[T any](src T) T {
	b, _ := json.Marshal(src) // 先序列化
	var dst T
	_ = json.Unmarshal(b, &dst) // 再反序列化
	return dst                  // dst 與 src 完全獨立
}
