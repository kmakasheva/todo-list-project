package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func InitLogger(env string) {
	//TODO: передать конфиг а не просто локал (попозже создам config для себя)
	Log = SetupLogger("local")

	//TODO: передать конфиг а не просто локал (попозже создам config для себя)
	Log = Log.With("env", "local")

	//Log.Debug("debug messages are enabled")
	//Log.Info("starting todo-list project")
	//Log.Warn("warning messages are enabled")
	//Log.Error("error messages are enabled")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
