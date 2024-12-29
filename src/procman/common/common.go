package common

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func GetLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return log.Logger
}
