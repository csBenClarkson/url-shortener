package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/csBenClarkson/url-shortener/logger"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

func main() {
	var debug bool
	var logPath string

	flag.BoolVar(&debug, "debug", true, "Enable debug mode.")
	flag.StringVar(&logPath, "logPath", "log", "Directory to store log files.")
	flag.Parse()

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
