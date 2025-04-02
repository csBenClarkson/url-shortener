package logger

import (
	"flag"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func init() {
	var debug bool
	var logPath string
	flag.BoolVar(&debug, "debug", false, "Enable debug options.")
	flag.StringVar(&logPath, "logPath", "log/shortener.log", "Path to store the log file.")
	flag.Parse()

	if debug {
		textHandler := tint.NewHandler(os.Stdout, nil)
		slog.SetDefault(slog.New(textHandler))
	}

}
