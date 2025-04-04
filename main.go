package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/csBenClarkson/url-shortener/logger"
	"github.com/csBenClarkson/url-shortener/store"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

func main() {
	var debug bool
	var logPath string
	var redisHost string
	var redisPort string
	var redisPass string
	var redisDB int
	var sqliteFile string

	flag.BoolVar(&debug, "debug", true, "Enable debug mode. Default: true")
	flag.StringVar(&logPath, "logPath", "log", "Directory to store log files. Default: ./log")
	flag.StringVar(&redisHost, "redisHost", "127.0.0.1", "Redis server host address. Default: 127.0.0.1")
	flag.StringVar(&redisPort, "redisPort", "6379", "Redis server port. Default: 6379")
	flag.StringVar(&redisPass, "redisPass", "", "Redis server password. Default: <empty>")
	flag.IntVar(&redisDB, "redisDB", 0, "Which redis database is used. Default: 0")
	flag.StringVar(&sqliteFile, "sqliteFile", "data.db", "Database file for sqlite3. Default: ./data.db")
	flag.Parse()

	// Setting up loggers
	if debug {
		file := logger.CreateLogFile(logPath)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(file)
		jsonHandler := slog.NewJSONHandler(file, nil)
		slog.SetDefault(slog.New(jsonHandler))
	}
	textHandler := tint.NewHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(textHandler))

	storage := store.Storage{
		RedisHost:   redisHost,
		RedisPort:   redisPort,
		RedisPass:   redisPass,
		RedisDB:     redisDB,
		RedisClient: nil,

		SqliteFile:   sqliteFile,
		SqliteClient: nil,
	}

	err := storage.InitDB()
	if err != nil {
		slog.Error("Error when initializing databases. Exit...")
		os.Exit(1)
	}
	defer storage.RedisClient.Close()
	defer storage.SqliteClient.Close()
	slog.Info("Databases are initialized successfully!")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello Go URL-shortener!",
		})
	})

	err := r.Run(":9008")
	if err != nil {
		panic(fmt.Sprint("Failed to start web server. Error: %v", err))
	}
}
