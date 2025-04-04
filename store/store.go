package store

import (
	"context"
	"database/sql"
	"fmt"

	"log/slog"
	"net"

	"github.com/avast/retry-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	RedisHost   string
	RedisPort   string
	RedisPass   string
	RedisDB     int
	RedisClient *redis.Client

	SqliteFile   string
	SqliteClient *sql.DB
}

// InitDB initialize Redis and Sqlite, try connecting to them and return error if something fatal happened.
func (s Storage) InitDB() error {
	slog.Info("Initializing databases...")
	s.RedisClient = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(s.RedisHost, s.RedisPort),
		Password: s.RedisPass,
		DB:       s.RedisDB,
	})
	ctx := context.Background()

	slog.Debug("Check if Redis server is reachable.")
	err := checkRedisReachable(ctx, s.RedisClient)
	if err != nil {
		slog.Error("Failed to connect to Redis. All attempts failed.")
		return err
	}

	slog.Debug("Check if Sqlite3 is reachable.")
	s.SqliteClient, err = sql.Open("sqlite3", s.SqliteFile)
	if err != nil {
		slog.Error("Failed to open or create Sqlite3 database file.")
		return err
	}
	err = checkSQLiteReachable(ctx, s.SqliteClient)
	if err != nil {
		slog.Error("Cannot connect to Sqlite3. All attempts failed.")
		return err
	}
	slog.Debug("Databases are alive!")

	err = createSqliteTableIndex(s.SqliteClient)
	if err != nil {
		slog.Error("Failed to create table or index in Sqlite3.")
		return err
	}

	slog.Info("Create table and index in Sqlite3 sucessfully!")

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
			slog.Warn(fmt.Sprintf("Cannot connect to Redis. Retry: %d Error: %v", n, err))
		}),
	)
	return err
}

func checkSQLiteReachable(ctx context.Context, client *sql.DB) error {
	err := retry.Do(
		func() error {
			return client.PingContext(ctx)
		},
		retry.Attempts(8),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn(fmt.Sprintf("Cannot connect to Sqlite3. Retry: %d Error: %v", n, err))
		}),
	)
	return err
}

func createSqliteTableIndex(client *sql.DB) error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS Shortener (id INT PRIMARY KEY AUTOINCREMENT, 
													  url TEXT NOT NULL,
													  digest TEXT NOT NULL,
													  date DATETIME NOT NULL);
				CREATE UNIQUE INDEX idx_digest ON TABLE Shortener (digest);`
	_, err := client.Exec(sqlStmt)
	return err
}

// GetOriginalURL return the corresponding URL according to the short digest
// It first finds if the digest appears as a key in Redis.
// Then it retrive URL from redis if it exists, otherwise from MySQL.
// It returns error only when record does not exist on either databases.
func (s Storage) GetOriginalURL(ctx context.Context, digest string) (string, error) {
	val, err := s.RedisClient.Get(ctx, digest).Result()
	if err == nil {
		return val, nil
	}
	return s.getFromSqlite(ctx, digest)
}

func (s Storage) getFromSqlite(ctx context.Context, digest string) (string, error) {
	url := ""
	row := s.SqliteClient.QueryRowContext(ctx, "SELECT url FROM shortener LIMIT 1 WHERE digest = ?", digest)
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

// StoreURL sets up a mapping from the original URL to a generated digest,
// store it into databases, and return the digest.
// MurMur -> number -> 62-radix -> string
// Bloom Fliter on Redis
func StoreURL(ctx context.Context, url string) (string, error) {

}
