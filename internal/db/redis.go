package db

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedis() RedisStore {
	uri, ok := os.LookupEnv("REDIS_URI")
	if !ok {
		uri = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(uri)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)

	return RedisStore{
		rdb: rdb,
	}
}

func (r *RedisStore) Close() error {
	return r.rdb.Close()
}

func (r *RedisStore) SetPlayerData(ctx context.Context, key string, value PlayerData) error {
	data, _ := json.Marshal(value)

	err := r.rdb.Set(ctx, "steamid:"+key, data, time.Hour*24*7).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisStore) GetPlayerData(ctx context.Context, key string) (*PlayerData, error) {
	val, err := r.rdb.Get(ctx, "steamid:"+key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var playerData PlayerData
	json.Unmarshal([]byte(val), &playerData)
	return &playerData, nil
}

type PlayerData struct {
	Username    string
	Avatar      string
	CustomURL   string
	RealName    string
	SteamID64   string
	Location    string
	CreatedAt   int64
	LastUpdated int64
}
