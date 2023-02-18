package main

import (
	"edgedb-ingest/pkg/config"
	"edgedb-ingest/pkg/models"
	"os"
	"syscall"

	"os/signal"

	"github.com/edgedb/edgedb-go"

	"github.com/rs/zerolog/log"

	ks "github.com/washed/kitchen-sink-go"
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
	config := config.Config{}
	ks.ReadConfig(&config)
	ks.InitLogger(config.Log)

	dbClient := getEdgeDbClient()
	defer dbClient.Close()

	for _, trvId := range config.ShellyTRVIDs {
		trv := models.IngestShellyTRV(dbClient, trvId)
		defer trv.Close()
	}

	for _, dw2Id := range config.ShellyDW2IDs {
		dw2 := models.IngestShellyDW2(dbClient, dw2Id)
		defer dw2.Close()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info().Msg("Exiting")
}
