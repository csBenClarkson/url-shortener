package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/csBenClarkson/url-shortener/logger"
	"github.com/csBenClarkson/url-shortener/router"
	"github.com/csBenClarkson/url-shortener/store"
	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
)

func main() {
	var debug bool
	var host string
	var port string
	var logPath string
	var redisHost string
	var redisPort string
	var redisPass string
	var redisDB int
	var sqliteFile string

	flag.BoolVar(&debug, "debug", true, "Enable debug mode")
	flag.StringVar(&host, "host", "127.0.0.1", "Host address to run the shortener server.")
	flag.StringVar(&port, "port", "9008", "Port to run the shortener server.")
	flag.StringVar(&logPath, "logPath", "log", "Directory to store log files.")
	flag.StringVar(&redisHost, "redisHost", "127.0.0.1", "Redis server host address.")
	flag.StringVar(&redisPort, "redisPort", "6379", "Redis server port.")
	flag.StringVar(&redisPass, "redisPass", "", "Redis server password. (default <empty>)")
	flag.IntVar(&redisDB, "redisDB", 0, "Which redis database is used. (default 0)")
	flag.StringVar(&sqliteFile, "sqliteFile", "data.db", "Database file for sqlite3.")
	flag.Parse()

	// Setting up loggers
	coreLog := logger.CreateLogFile(logPath, "core.log")
	webLog := logger.CreateLogFile(logPath, "web.log")
	defer coreLog.Close()
	defer webLog.Close()

	jsonHandlerCore := slog.NewJSONHandler(coreLog, nil)
	jsonHandlerWeb := slog.NewJSONHandler(webLog, nil)

	var webLogger *slog.Logger
	var coreLogger *slog.Logger

	if debug {
		textHandlerCore := tint.NewHandler(os.Stdout, nil)
		textHandlerWeb := tint.NewHandler(os.Stdout, nil)

		webLogger = slog.New(slogmulti.Fanout(jsonHandlerWeb, textHandlerWeb))
		coreLogger = slog.New(slogmulti.Fanout(jsonHandlerCore, textHandlerCore))
		slog.SetDefault(coreLogger)
	} else {
		webLogger = slog.New(jsonHandlerWeb)
		coreLogger = slog.New(jsonHandlerCore)
		slog.SetDefault(coreLogger)
	}

	storage := &store.Storage{}

	err := storage.InitDB(redisHost, redisPort, redisPass, redisDB, sqliteFile)
	if err != nil {
		slog.Error("Error when initializing databases. Exit...")
		os.Exit(1)
	}
	defer storage.RedisClient.Close()
	defer storage.SqliteClient.Close()
	slog.Info("Databases are initialized successfully!")

	// Checking if ENV for password is set.
	_, setPass := os.LookupEnv("SHORTENER_SECRET")
	if !setPass {
		slog.Error("Environment variable SHORTENER_SECRET is not set. Please set a secret for login.")
		os.Exit(1)
	}

	r := router.SetupRouter(webLogger, storage)
	err = r.Run(host + ":" + port)
	if err != nil {
		slog.Error("Failed to start the web server. Error: %v", err)
		os.Exit(1)
	}
}
