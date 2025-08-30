package cache

import (
	"context"
	"fmt"
	"gold_portal/config"
	"sync"
	"time"
)

// Простая in-memory реализация для демонстрации
// В продакшене заменить на реальный Redis клиент
type RedisCache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Close() error
}

type redisCache struct {
	store map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	value      string
	expiration time.Time
}

func NewRedisCache(cfg *config.Config) (RedisCache, error) {
	if cfg.Redis.Host == "" {
		return nil, fmt.Errorf("redis host is required")
	}

	cache := &redisCache{
		store: make(map[string]cacheItem),
	}

	// Запускаем очистку устаревших записей
	go cache.cleanupExpired()

	return cache, nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	expTime := time.Now().Add(expiration)
	r.store[key] = cacheItem{
		value:      fmt.Sprintf("%v", value),
		expiration: expTime,
	}
	return nil
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	item, exists := r.store[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if time.Now().After(item.expiration) {
		delete(r.store, key)
		return "", fmt.Errorf("key expired")
	}

	return item.value, nil
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.store, key)
	return nil
}

func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	item, exists := r.store[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(item.expiration) {
		delete(r.store, key)
		return false, nil
	}

	return true, nil
}

func (r *redisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.store[key]; exists {
		return false, nil
	}

	expTime := time.Now().Add(expiration)
	r.store[key] = cacheItem{
		value:      fmt.Sprintf("%v", value),
		expiration: expTime,
	}
	return true, nil
}

func (r *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	item, exists := r.store[key]
	var current int64 = 0
	if exists {
		fmt.Sscanf(item.value, "%d", &current)
	}

	current++
	expTime := time.Now().Add(time.Hour * 24) // По умолчанию 24 часа
	r.store[key] = cacheItem{
		value:      fmt.Sprintf("%d", current),
		expiration: expTime,
	}
	return current, nil
}

func (r *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	item, exists := r.store[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	item.expiration = time.Now().Add(expiration)
	r.store[key] = item
	return nil
}

func (r *redisCache) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.store = make(map[string]cacheItem)
	return nil
}

func (r *redisCache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		now := time.Now()
		for key, item := range r.store {
			if now.After(item.expiration) {
				delete(r.store, key)
			}
		}
		r.mutex.Unlock()
	}
}
