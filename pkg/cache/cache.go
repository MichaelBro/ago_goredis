package cache

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

const cacheTimeout = 50 * time.Millisecond

type Service struct {
	pool *redis.Pool
}

func NewService(redisPool *redis.Pool) *Service {
	return &Service{
		pool: redisPool,
	}
}

func (s *Service) Get(ctx context.Context, key string) (bytes []byte, err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	reply, err := redis.DoWithTimeout(conn, cacheTimeout, "GET", key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	value, err := redis.Bytes(reply, err)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return value, nil
}

func (s *Service) Set(ctx context.Context, key string, bytes []byte) (err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	_, err = redis.DoWithTimeout(conn, cacheTimeout, "SET", key, bytes)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Service) DeleteAllCache(ctx context.Context) (err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	_, err = redis.DoWithTimeout(conn, cacheTimeout, "FLUSHDB")
	if err != nil {
		log.Println(err)
	}
	return err
}
