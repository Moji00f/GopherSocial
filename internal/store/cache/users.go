package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Moji00f/GopherSocial/internal/store"
	"github.com/go-redis/redis/v8"
	"time"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (u *UserStore) Get(ctx context.Context, userId int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userId)

	data, err := u.rdb.Get(ctx, cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return u.rdb.Set(ctx, cacheKey, json, UserExpTime).Err()
}

func (u *UserStore) Delete(ctx context.Context, userId int64) {
	cacheKey := fmt.Sprintf("user-%v", userId)
	u.rdb.Del(ctx, cacheKey)
}
