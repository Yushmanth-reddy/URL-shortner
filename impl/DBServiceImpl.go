package impl

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"../base62"
	"../storage"

	"github.com/gomodule/redigo/redis"
)

type redisClient struct{ pool *redis.Pool }

func NewPool(host, port, password string) (*redisClient, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host+":"+port)
		},
	}
	return &(redisClient{pool}), nil

}

func (r *redisClient) IsAvailable(id uint64) bool {
	conn := r.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", "Shortener:"+strconv.FormatUint(id, 10)))

	if err != nil {
		return false
	}

	return exists

}

func (r *redisClient) Save(url string, expiration time.Time) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	var id uint64

	for used := true; used; used = r.IsAvailable(id) {
		id = rand.Uint64()
	}

	shortLink := storage.Item{
		Id:         id,
		URL:        url,
		Expiration: expiration.Format("2006-01-02 15:04:05"),
		Visits:     0,
	}

	_, err := conn.Do("HSET", redis.Args{"Shortener:" + strconv.FormatUint(id, 10)}.AddFlat(shortLink)...)
	if err != nil {
		return "", err
	}

	_, err = conn.Do("EXPIREAT", "Shortener:"+strconv.FormatUint(id, 10), expiration.Unix())
	if err != nil {
		return "", err
	}

	return base62.Encode(id), err
}

func (r *redisClient) Load(code string) (string, error) {
	id, err := base62.Decode(code)

	if err != nil {
		return "", err
	}

	conn := r.pool.Get()
	defer conn.Close()

	id, err = redis.Uint64(conn.Do("HGET", "Shortener:"+strconv.FormatUint(id, 10), "id"))

	urlString, err := redis.String(conn.Do("HGET", "Shortener:"+strconv.FormatUint(id, 10), "url"))

	if err != nil {
		return "", err
	} else if len(urlString) == 0 {
		return "", errors.New("Empty url")
	}

	_, err = conn.Do("HINCRBY", "Shortener:"+strconv.FormatUint(id, 10), "visits", 1)

	return urlString, nil

}

func (r *redisClient) LoadInfo(code string) (*storage.Item, error) {
	id, err := base62.Decode(code)
	if err != nil {
		return nil, err
	}

	conn := r.pool.Get()
	defer conn.Close()

	values, err := redis.Values(conn.Do("HGETALL", "Shortener:"+strconv.FormatUint(id, 10)))

	if err != nil {
		return nil, err
	} else if len(values) == 0 {
		return nil, errors.New("0 values")
	}

	var item storage.Item

	err = redis.ScanStruct(values, &item)

	return &item, err

}

func (r *redisClient) Close() error {
	return r.pool.Close()
}
