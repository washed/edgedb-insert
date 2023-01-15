package main

import (
	"context"
	"edgedb-ingest/pkg/config"
	"edgedb-ingest/pkg/models"
	"os"
	"syscall"
	"time"

	"os/signal"

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

	mqttOpts.SetOrderMatters(false)
	mqttOpts.ConnectTimeout = time.Second
	mqttOpts.WriteTimeout = time.Second
	mqttOpts.KeepAlive = 10
	mqttOpts.PingTimeout = time.Second
	mqttOpts.ConnectRetry = true
	mqttOpts.AutoReconnect = true

	mqttOpts.DefaultPublishHandler = func(_ MQTT.Client, msg MQTT.Message) {
		log.Warn().Interface("msg", msg).Msg("unexpected message")
	}

	mqttOpts.OnConnectionLost = func(cl MQTT.Client, err error) {
		log.Err(err).Msg("MQTT connection lost")
	}

	mqttOpts.OnReconnecting = func(MQTT.Client, *MQTT.ClientOptions) {
		log.Warn().Msg("MQTT attempting to reconnect")
	}

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

func ingestShellyTRV(ctx context.Context, dbClient *edgedb.Client, trvId string) shelly.ShellyTRV {
	trv := shelly.NewShellyTRV(trvId, getMQTTOpts())
	trv.Connect()

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

	go trv.SubscribeInfo(infoCallback)

	return trv
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
		trv := ingestShellyTRV(ctx, dbClient, trvId)
		defer trv.Close()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info().Msg("Exiting")
}
