package repo

import (
	"core/app"
	"core/internal/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const CacheRootKey = "network_social"

const (
	CacheKeyFriendsByUserID = CacheRootKey + ":friends_by_user_id:%s" // friends_by_user_id:{user_id} -> list of friend IDs
)

const (
	TTLCacheFriendsByUserID = 10 * time.Minute
)

func GetFriendsByUserIDFromCache(userID string) ([]model.User, error) {
	if app.RedisClient == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf(CacheKeyFriendsByUserID, userID)
	raw, err := app.RedisClient.Get(app.Ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var friends []model.User
	if err := json.Unmarshal([]byte(raw), &friends); err != nil {
		return nil, nil
	}

	return friends, nil
}

func SetFriendsByUserIDCache(userID string, friends []model.User) error {
	if app.RedisClient == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(CacheKeyFriendsByUserID, userID)
	data, err := json.Marshal(friends)
	if err != nil {
		return err
	}

	return app.RedisClient.Set(app.Ctx, cacheKey, data, TTLCacheFriendsByUserID).Err()
}

func InvalidateFriendsByUserIDCache(userIDs ...string) error {
	if app.RedisClient == nil {
		return nil
	}

	if len(userIDs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, fmt.Sprintf(CacheKeyFriendsByUserID, userID))
	}

	return app.RedisClient.Del(app.Ctx, keys...).Err()
}
