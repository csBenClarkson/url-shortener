package store

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
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

func (s storage) InitDB() error {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.redisAddr + ":" + s.redisPort,
		Password: s.redisPass,
		DB:       s.redisDB,
	})
	ctx := context.Background()
	err := s.redisClient.Ping(ctx).Err()
	if err != nil {

	}

	return nil
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
