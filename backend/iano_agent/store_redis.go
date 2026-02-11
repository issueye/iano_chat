package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client     *redis.Client
	keyPrefix  string
	expiration time.Duration
}

type RedisStoreConfig struct {
	KeyPrefix  string
	Expiration time.Duration
}

func DefaultRedisStoreConfig() *RedisStoreConfig {
	return &RedisStoreConfig{
		KeyPrefix:  "iano:conversation:",
		Expiration: 24 * time.Hour,
	}
}

func NewRedisStore(client *redis.Client, cfg *RedisStoreConfig) *RedisStore {
	if cfg == nil {
		cfg = DefaultRedisStoreConfig()
	}
	return &RedisStore{
		client:     client,
		keyPrefix:  cfg.KeyPrefix,
		expiration: cfg.Expiration,
	}
}

func (s *RedisStore) makeKey(sessionID string) string {
	return s.keyPrefix + sessionID
}

func (s *RedisStore) Save(ctx context.Context, sessionID string, layer *ConversationLayer) error {
	data := LayerToData(layer)
	jsonData, err := data.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化对话数据失败: %w", err)
	}

	key := s.makeKey(sessionID)
	if err := s.client.Set(ctx, key, jsonData, s.expiration).Err(); err != nil {
		return fmt.Errorf("保存对话数据到 Redis 失败: %w", err)
	}

	return nil
}

func (s *RedisStore) Load(ctx context.Context, sessionID string) (*ConversationLayer, error) {
	key := s.makeKey(sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("从 Redis 加载对话数据失败: %w", err)
	}

	convData, err := ConversationDataFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("解析对话数据失败: %w", err)
	}

	return DataToLayer(convData), nil
}

func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	key := s.makeKey(sessionID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("删除对话数据失败: %w", err)
	}
	return nil
}

func (s *RedisStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := s.makeKey(sessionID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("检查对话数据是否存在失败: %w", err)
	}
	return exists > 0, nil
}

func (s *RedisStore) SetExpiration(ctx context.Context, sessionID string, expiration time.Duration) error {
	key := s.makeKey(sessionID)
	return s.client.Expire(ctx, key, expiration).Err()
}

func (s *RedisStore) GetTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	key := s.makeKey(sessionID)
	return s.client.TTL(ctx, key).Result()
}
