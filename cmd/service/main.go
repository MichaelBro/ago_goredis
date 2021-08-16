package main

import (
	"ago_goredis/cmd/service/app"
	"ago_goredis/pkg/cache"
	"ago_goredis/pkg/news"
	"context"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort     = "9999"
	defaultHost     = "0.0.0.0"
	defaultDbDSN    = "postgres://app:pass@localhost:5432/db"
	defaultCacheDSN = "redis://localhost:6379/0"
)

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	dbDSN, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dbDSN = defaultDbDSN
	}

	cacheDSN, ok := os.LookupEnv("APP_CACHE_DSN")
	if !ok {
		cacheDSN = defaultCacheDSN
	}

	if err := execute(net.JoinHostPort(host, port), dbDSN, cacheDSN); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, dbDSN string, cacheDSN string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	redisPool := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURL(cacheDSN)
		},
	}

	defer func() {
		if cerr := redisPool.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	newsSvc := news.NewService(pool)
	router := chi.NewRouter()
	cacheSvc := cache.NewService(redisPool)

	application := app.NewServer(newsSvc, router, cacheSvc)

	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}
