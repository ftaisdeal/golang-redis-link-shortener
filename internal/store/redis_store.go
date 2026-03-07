package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeAlreadyExists = errors.New("code already exists")
	ErrCodeNotFound      = errors.New("code not found")
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string, db int) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStore{client: client}
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}

func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisStore) CreateShortURL(ctx context.Context, code, originalURL string, ttl time.Duration) error {
	key := shortKey(code)
	args := redis.SetArgs{Mode: "NX"}
	if ttl > 0 {
		args.TTL = ttl
	}

	ok, err := r.client.SetArgs(ctx, key, originalURL, args).Result()
	if err != nil {
		return fmt.Errorf("create short url: %w", err)
	}
	if ok != "OK" {
		return ErrCodeAlreadyExists
	}

	if err := r.client.SetNX(ctx, clicksKey(code), 0, 0).Err(); err != nil {
		return fmt.Errorf("create clicks counter: %w", err)
	}
	if ttl > 0 {
		_ = r.client.Expire(ctx, clicksKey(code), ttl).Err()
	}
	return nil
}

func (r *RedisStore) GetOriginalURL(ctx context.Context, code string) (string, error) {
	val, err := r.client.Get(ctx, shortKey(code)).Result()
	if err == redis.Nil {
		return "", ErrCodeNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get original url: %w", err)
	}
	return val, nil
}

func (r *RedisStore) IncrementClicks(ctx context.Context, code string) error {
	if err := r.client.Incr(ctx, clicksKey(code)).Err(); err != nil {
		return fmt.Errorf("increment clicks: %w", err)
	}
	return nil
}

func (r *RedisStore) AllowRequest(ctx context.Context, ip string, maxPerMinute int) (bool, error) {
	now := time.Now().UTC()
	key := fmt.Sprintf("rl:%s:%s", ip, now.Format("200601021504"))

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limit increment: %w", err)
	}
	if count == 1 {
		_ = r.client.Expire(ctx, key, time.Minute).Err()
	}
	return int(count) <= maxPerMinute, nil
}

func shortKey(code string) string {
	return "short:" + code
}

func clicksKey(code string) string {
	return "clicks:" + code
}
