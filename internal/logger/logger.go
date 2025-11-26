package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New initialises a zerlog.Logger based on env and also sets the global loggers.
func New(env string) zerolog.Logger {
	var l zerolog.Logger

	if env == "dev" {
		// Human-friendly logs for local dev
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
		l = zerolog.New(consoleWriter).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// JSON logs for production (possible ELK)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		l = zerolog.New(os.Stdout).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	l = l.With().Str("env", env).Logger()

	zerolog.DefaultContextLogger = &l // set global default logger used when calling Ctx()
	log.Logger = l                    // set global logger used by log package
	return l
}
