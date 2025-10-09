package service

import (
	"context"
	"errors"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Valkey related errors.
var (
	ErrFailedValkeyOperation = errors.New("failed valkey operation")
	ErrFailedValkeyParse     = errors.New("failed to parse valkey result")
)

type cacheHgetallConfig struct {
	withCache  bool
	expiration time.Duration
}

// WithoutLocalCache is an option to disable caching for a specific operation.
func WithoutLocalCache() Option[cacheHgetallConfig] {
	return func(c *cacheHgetallConfig) {
		c.withCache = false
	}
}

// WithExpiration sets the cache expiration duration.
func WithExpiration(duration time.Duration) Option[cacheHgetallConfig] {
	return func(c *cacheHgetallConfig) {
		c.expiration = duration
	}
}

// CacheHgetall is a helper struct for caching hgetall results.
func (s *BaseService) CacheHgetall(
	ctx context.Context,
	key string,
	opts ...Option[cacheHgetallConfig],
) (map[string]string, error) {
	cfg := &cacheHgetallConfig{
		withCache:  true,
		expiration: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	query := s.Valkey.B().Hgetall().Key(key)

	var result valkey.ValkeyResult
	if cfg.withCache {
		result = s.Valkey.DoCache(ctx, query.Cache(), cfg.expiration)
	} else {
		result = s.Valkey.Do(ctx, query.Build())
	}
	if result.Error() != nil {
		return nil, ErrFailedValkeyOperation
	}

	v, err := result.AsStrMap()
	if err != nil {
		return nil, ErrFailedValkeyParse
	}

	return v, nil
}

type cacheHsetConfig struct {
	hasTTL bool
	ttl    time.Duration
}

// WithTTL sets the TTL for the cache entry.
func WithTTL(duration time.Duration) Option[cacheHsetConfig] {
	return func(c *cacheHsetConfig) {
		c.hasTTL = true
		c.ttl = duration
	}
}

// CacheHset is a helper for storing hash maps in cache.
func (s *BaseService) CacheHset(
	ctx context.Context,
	key string,
	values map[string]string,
	opts ...Option[cacheHsetConfig],
) error {
	cfg := &cacheHsetConfig{
		hasTTL: false,
		ttl:    0,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	query := s.Valkey.B().
		Hset().
		Key(key).
		FieldValue().
		FieldValueIter(func(yield func(string, string) bool) {
			for k, v := range values {
				if !yield(k, v) {
					break
				}
			}
		}).
		Build()

	if cfg.hasTTL {
		var queries valkey.Commands
		queries = append(queries, query)
		queries = append(
			queries,
			s.Valkey.B().Expire().Key(key).Seconds(int64(cfg.ttl.Seconds())).Build(),
		)
		results := s.Valkey.DoMulti(ctx, queries...)
		for _, result := range results {
			if result.Error() != nil {
				return ErrFailedValkeyOperation
			}
		}
		return nil
	}
	result := s.Valkey.Do(ctx, query)
	if result.Error() != nil {
		return ErrFailedValkeyOperation
	}
	return nil

}

// CacheDel is a helper for deleting cache entries.
func (s *BaseService) CacheDel(ctx context.Context, key string) error {
	query := s.Valkey.B().Del().Key(key).Build()
	result := s.Valkey.Do(ctx, query)
	if result.Error() != nil {
		return ErrFailedValkeyOperation
	}
	return nil
}

// CacheExpire is a helper for setting expiration on cache entries.
func (s *BaseService) CacheExpire(ctx context.Context, key string, duration time.Duration) error {
	query := s.Valkey.B().Expire().Key(key).Seconds(int64(duration.Seconds())).Build()
	result := s.Valkey.Do(ctx, query)
	if result.Error() != nil {
		return ErrFailedValkeyOperation
	}
	return nil
}

// CacheSet is a helper for setting simple string values in cache.
func (s *BaseService) CacheSet(
	ctx context.Context,
	key, value string,
	opts ...Option[cacheHsetConfig],
) error {
	cfg := &cacheHsetConfig{
		hasTTL: false,
		ttl:    0,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	query := s.Valkey.B().Set().Key(key).Value(value).Build()

	if cfg.hasTTL {
		var queries valkey.Commands
		queries = append(queries, query)
		queries = append(
			queries,
			s.Valkey.B().Expire().Key(key).Seconds(int64(cfg.ttl.Seconds())).Build(),
		)
		results := s.Valkey.DoMulti(ctx, queries...)
		for _, result := range results {
			if result.Error() != nil {
				return ErrFailedValkeyOperation
			}
		}
		return nil
	}
	result := s.Valkey.Do(ctx, query)
	if result.Error() != nil {
		return ErrFailedValkeyOperation
	}
	return nil
}
