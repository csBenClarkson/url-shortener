package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avast/retry-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net"
)

type storage struct {
	redisAddr   string
	redisPort   string
	redisPass   string
	redisDB     int
	redisClient *redis.Client

	sqlitePath   string
	sqliteClient *sql.DB
}

// InitDB initialize Redis and Sqlite, try connecting to them and return error if something fatal happened.
func (s storage) InitDB() error {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(s.redisAddr, s.redisPort),
		Password: s.redisPass,
		DB:       s.redisDB,
	})
	ctx := context.Background()
	err := checkRedisReachable(ctx, s.redisClient)
	if err != nil {
		slog.Error("Failed to connect to Redis. All attempts failed.")
		return err
	}

	s.sqliteClient, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		slog.Error("Failed to open or create Sqlite3 database file.")
		return err
	}
	return nil
}

func checkRedisReachable(ctx context.Context, client *redis.Client) error {
	err := retry.Do(
		func() error {
			return client.Ping(ctx).Err()
		},
		retry.Attempts(8),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Info(fmt.Sprintf("Cannot connect to Redis. Retry: %d Error: %v", n, err))
		}),
	)
	return err
}

func checkSQLiteReachable(ctx context.Context, client *sql.DB) error {
	client.Ping()
}

// GetOriginalLink return the corresponding URL according to the short digest
// It first find if the digest appears as a key in Redis.
// Then it retrive URL from redis if it exists, otherwise from MySQL.
func (s storage) GetOriginalLink(ctx context.Context, short string) string {
	val, err := s.redisClient.Get(ctx, short).Result()
	if err == nil {
		return val
	}
}

func (s storage) getFromSqlite(ctx context.Context, digest string) (val string, err error) {
	link := ""
	row := s.sqliteClient.QueryRow("SELECT link FROM shortener LIMIT 1 WHERE digest = ?", digest)
	err := row.Scan(&link)
}
