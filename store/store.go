package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"log/slog"
	"net"
	"time"

	"github.com/avast/retry-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	RedisClient  *redis.Client
	SqliteClient *sql.DB
}

var ErrURLExists = errors.New("the URL had existed")
var ErrNoSuchRecord = errors.New("cannot find the mapping")
var ErrDBFails = errors.New("database connection error")
var ErrTooLucky = errors.New("duplicate keys after hash with salt")

const redisNamespace = "urlshortener:"

// InitDB initialize Redis and Sqlite, try connecting to them and return error if something fatal happened.
func (s Storage) InitDB(redisHost string, redisPort string, redisPass string, redisDB int, sqliteFile string) error {
	slog.Info("Initializing databases...")
	s.RedisClient = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	})
	ctx := context.Background()

	slog.Debug("Check if Redis server is reachable.")
	err := checkRedisReachable(ctx, s.RedisClient)
	if err != nil {
		slog.Error("Failed to connect to Redis. All attempts failed.")
		return err
	}

	slog.Debug("Check if Sqlite3 is reachable.")
	s.SqliteClient, err = sql.Open("sqlite3", sqliteFile)
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
	sqlStmt := `CREATE TABLE IF NOT EXISTS Shortener (url TEXT PRIMARY KEY,
													  digest TEXT NOT NULL,
													  date DATETIME NOT NULL,
													  collide BOOLEAN);
				CREATE UNIQUE INDEX idx_digest ON TABLE Shortener (digest);`
	_, err := client.Exec(sqlStmt)
	return err
}

// GetOriginalURL return the corresponding URL according to the short digest
// It first finds if the digest appears as a key in Redis.
// Then it retrive URL from redis if it exists, otherwise from MySQL.
// It returns error only when record does not exist on either databases.
func (s Storage) GetOriginalURL(ctx context.Context, digest string) (string, error) {
	val, err := s.RedisClient.Get(ctx, redisNamespace+digest).Result()
	if err == nil {
		return val, nil
	}
	return s.getFromSqlite(ctx, digest)
}

func (s Storage) getFromSqlite(ctx context.Context, digest string) (string, error) {
	url := ""
	row := s.SqliteClient.QueryRowContext(ctx, "SELECT url FROM shortener WHERE digest = ? LIMIT 1", digest)
	err := row.Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoSuchRecord
		}
		return "", ErrDBFails
	}
	return url, nil
}

// StoreURL sets up a mapping from the original URL to a generated digest, and store it into both databases.
// It returns the digest and error.
func (s Storage) StoreURL(ctx context.Context, url string) (string, error) {
	urlExist, err := s.checkURLExist(ctx, url)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDBFails, err)
	}

	var digest string
	if !urlExist {
		digest = GenerateDigest(url)
		digestExist, err := s.checkDigestExist(ctx, digest)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrDBFails, err)
		}

		var now time.Time
		if digestExist {
			now = time.Now()
			digest = GenerateDigest(url + now.String())
		}

		err = s.storeToSqlite(ctx, url, digest, now, digestExist)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				// Break unique key constrain
				if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					return "", ErrTooLucky
				}
			}
			return "", fmt.Errorf("%w: %v", ErrDBFails, err)
		}

		err = s.storeToRedis(ctx, url, digest)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrDBFails, err)
		}

	} else {
		return "", ErrURLExists
	}

	return digest, nil
}

// checkURLExist check if a URL is in Sqlite3 database, return a boolean and an error.
// Error occurs only when database connection goes wrong.
// The value of the boolean should be omitted if error occurs.
func (s Storage) checkURLExist(ctx context.Context, url string) (bool, error) {
	urlBFExist, err := s.RedisClient.BFExists(ctx, "url_filter", url).Result()
	if err != nil {
		return true, err
	}

	if urlBFExist {
		err = s.SqliteClient.QueryRowContext(ctx, "SELECT url FROM shortener WHERE url = ?", url).Scan()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return true, err
		} else {
			return true, nil
		}
	} else {
		return false, nil
	}
}

// checkDigestExist check if a digest is in Sqlite3 database, return a boolean and an error.
// Error occurs only when database connection goes wrong.
// The value of the boolean should be omitted if error occurs.
func (s Storage) checkDigestExist(ctx context.Context, digest string) (bool, error) {
	digestBFExist, err := s.RedisClient.BFExists(ctx, "digest_filter", digest).Result()
	if err != nil {
		return false, err
	}

	if digestBFExist {
		err = s.SqliteClient.QueryRowContext(ctx, "SELECT digest FROM shortener WHERE digest = ?", digest).Scan()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return true, err
		} else {
			return true, nil
		}
	} else {
		return false, nil
	}
}

func (s Storage) storeToSqlite(ctx context.Context, url string, digest string, date time.Time, collide bool) error {
	_, err := s.SqliteClient.ExecContext(ctx, "INSERT INTO shortener (url, digest, date) VALUES (?, ?, ?, ?)", url, digest, date, collide)
	return err
}

func (s Storage) storeToRedis(ctx context.Context, url string, digest string) error {
	err := s.RedisClient.Set(ctx, redisNamespace+digest, url, 2*time.Hour).Err()
	return err
}
