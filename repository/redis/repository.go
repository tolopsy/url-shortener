package redis

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/tolopsy/url-shortener/shortener"
)

type redisRepository struct {
	client *redis.Client
}

func newRedisClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if _, err = client.Ping().Result(); err != nil {
		return nil, err
	}
	return client, nil
}

func NewRedisRepository(redisURL string) (shortener.RedirectRepository, error) {
	client, err := newRedisClient(redisURL)
	if err != nil {
		return nil, errors.Wrap(err, "repository.NewRedisRepository")
	}

	return &redisRepository{client: client}, nil
}

func (r *redisRepository) generateKey(code string) string {
	return fmt.Sprintf("redirect:%s", code)
}

func (r *redisRepository) Find(code string) (*shortener.Redirect, error) {
	key := r.generateKey(code)
	data, err := r.client.HGetAll(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "repository.redis.Find")
	}
	if len(data) == 0 {
		return nil, errors.Wrap(shortener.ErrRedirectNotFound, "repository.redis.Find")
	}

	createdAt, err := strconv.ParseInt(data["createdAt"], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "repository.redis.Find")
	}

	redirect := &shortener.Redirect{
		Code: data["code"],
		CreatedAt: createdAt,
		URL: data["url"],
	}
	return redirect, err
}

func (r *redisRepository) Store(redirect *shortener.Redirect) error {
	key := r.generateKey(redirect.Code)
	data := map[string]interface{}{
		"code": redirect.Code,
		"createdAt": redirect.CreatedAt,
		"url": redirect.URL,
	}
	_, err := r.client.HMSet(key, data).Result()
	if err != nil {
		return errors.Wrap(err, "redirect.redis.Store")
	}
	return nil
}
