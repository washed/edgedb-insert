package main

import (
	"context"
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

func ingestShellyTRV(ctx context.Context) {
	dbClient := getEdgeDbClient(ctx)
	defer dbClient.Close()

	trv := shelly.NewShellyTRV("60A423DAE8DE", getMQTTOpts())
	trv.Connect()
	defer trv.Close()

	infoCallback := func(status shelly.ShellyTRVInfo) {
		log.Debug().Interface("status", status).Msg("Received status")

		s := models.ShellyTRVDbModel{
			Device:            models.Device{DeviceId: "60A423DAE8DE"},
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
		log.Info().Str("id", inserted.Id.String()).Msg("inserted object")
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

	ctx := context.Background()

	/*

		edgeDbDSN := os.Getenv("EDGEDB_DSN")
		client, err := edgedb.CreateClientDSN(ctx, edgeDbDSN, edgedb.Options{})
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		s := models.ShellyTRV{
			Device:            models.Device{DeviceId: "fake-test-id-trv"},
			Timestamp:         time.Now().UTC(),
			Battery:           47,
			Position:          32,
			TargetTemperature: 21,
			Temperature:       20,
		}
		s.Insert(ctx, client)

		query := `select ShellyTRV {
			position,
			temperature,
			battery,
			timestamp,
			device: {
					name,
					device_id,
					device_type
				}
			}
			filter .device.device_id = "fake-test-id-trv"
			limit 1;`

		var result models.ShellyTRV
		err = client.QuerySingle(ctx, query, &result)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Got: %+v\n", result)
	*/

	go ingestShellyTRV(ctx)

	for {
		time.Sleep(time.Second * 10)
	}
}
