package infra

import (
	"context"
	"crypto/tls"

	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*redis.Client
}

func NewRedis(logger lib.Logger, env *lib.Env) (Redis, error) {

	tlsConfig := (*tls.Config)(nil)
	switch env.RedisTls {
	case "SSLv3":
		tlsConfig = &tls.Config{MinVersion: tls.VersionSSL30}
	case "TLS 1.0":
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS10}
	case "TLS 1.1":
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS11}
	case "TLS 1.2":
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	case "TLS 1.3":
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS13}
	}

	rc := redis.NewClient(&redis.Options{
		Addr:      env.RedisAddr,
		Password:  env.RedisPwd,
		DB:        env.RedisDb,
		PoolSize:  env.RedisPool,
		TLSConfig: tlsConfig,
	})

	return Redis{Client: rc}, nil
}

func (r *Redis) PushMessage(ctx context.Context, channel string, message string) error {
	ret := r.Client.Publish(ctx, channel, message)

	if ret.Err() != nil {
		return ret.Err()
	}
	return nil
}
