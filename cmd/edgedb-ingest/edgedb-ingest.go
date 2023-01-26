package main

import (
	"edgedb-ingest/pkg/config"
	"edgedb-ingest/pkg/models"
	"os"
	"syscall"
	"time"

	"os/signal"

	"github.com/edgedb/edgedb-go"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getEdgeDbClient() *edgedb.Client {
	edgeDbDSN := os.Getenv("EDGEDB_DSN")
	client, err := edgedb.CreateClientDSN(nil, edgeDbDSN, edgedb.Options{})
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to EdgeDB")
		os.Exit(1)
	}

	return client
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano},
	)

	configFilePath := "config.yaml"
	conf := config.Config{}

	err := config.ReadConfig(configFilePath, &conf)
	if err != nil {
		log.Error().Err(err).Str("configFilePath", configFilePath).Msg("error reading config file")
		os.Exit(1)
	}
	log.Info().
		Interface("conf", conf).
		Str("configFilePath", configFilePath).
		Msg("read config file")

	dbClient := getEdgeDbClient()
	defer dbClient.Close()

	for _, trvId := range conf.ShellyTRVIDs {
		trv := models.IngestShellyTRV(dbClient, trvId)
		defer trv.Close()
	}

	for _, dw2Id := range conf.ShellyDW2IDs {
		dw2 := models.IngestShellyDW2(dbClient, dw2Id)
		defer dw2.Close()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info().Msg("Exiting")
}
