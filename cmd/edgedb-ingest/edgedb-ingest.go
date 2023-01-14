package main

import (
	"context"
	"edgedb-ingest/pkg/config"
	"edgedb-ingest/pkg/models"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgedb/edgedb-go"

	"github.com/washed/shelly-go"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getMQTTOpts() *MQTT.ClientOptions {
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(os.Getenv("MQTT_BROKER_URL"))
	mqttOpts.SetUsername(os.Getenv("MQTT_BROKER_USERNAME"))
	mqttOpts.SetPassword(os.Getenv("MQTT_BROKER_PASSWORD"))
	return mqttOpts
}

func getEdgeDbClient(ctx context.Context) *edgedb.Client {
	edgeDbDSN := os.Getenv("EDGEDB_DSN")
	client, err := edgedb.CreateClientDSN(ctx, edgeDbDSN, edgedb.Options{})
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to EdgeDB")
	}

	return client
}

func ingestShellyTRV(ctx context.Context, dbClient *edgedb.Client, trvId string) {
	trv := shelly.NewShellyTRV(trvId, getMQTTOpts())
	trv.Connect()
	defer trv.Close()

	infoCallback := func(status shelly.ShellyTRVInfo) {
		log.Debug().Interface("status", status).Msg("Received status")

		s := models.ShellyTRVDbModel{
			Device:            models.Device{DeviceId: trvId},
			Timestamp:         time.Now().UTC(),
			Battery:           float32(status.Bat.Value),
			Position:          status.Thermostats[0].Pos,
			TargetTemperature: status.Thermostats[0].TargetT.Value,
			Temperature:       status.Thermostats[0].Tmp.Value,
		}
		inserted, err := s.Insert(ctx, dbClient)

		if err != nil {
			log.Error().Str("DeviceName", trv.DeviceName()).Err(err).Msg("Error inserting data")
		}
		log.Info().
			Str("DeviceName", trv.DeviceName()).
			Str("id", inserted.Id.String()).
			Msg("inserted object")
	}

	trv.SubscribeInfo(infoCallback)

	for {
		time.Sleep(time.Second * 10)
	}
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
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

	ctx := context.Background()
	dbClient := getEdgeDbClient(ctx)
	defer dbClient.Close()

	for _, trvId := range conf.ShellyTRVIDs {
		go ingestShellyTRV(ctx, dbClient, trvId)
	}

	for {
		time.Sleep(time.Second * 10)
	}
}
