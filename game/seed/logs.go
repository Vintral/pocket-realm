package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogs() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	output := zerolog.ConsoleWriter{
		Out:           os.Stderr,
		FieldsExclude: []string{zerolog.TimestampFieldName},
	}

	log.Logger = log.Output(output).With().Logger()
}
