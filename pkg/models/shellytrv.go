package models

import (
	"context"
	"fmt"
	"time"

	"github.com/edgedb/edgedb-go"
	"github.com/rs/zerolog/log"
	"github.com/washed/shelly-go"
)

type ShellyTRVDbModel struct {
	edgedb.Optional
	Timestamp         time.Time `edgedb:"timestamp"`
	Device            Device    `edgedb:"device"`
	Battery           float32   `edgedb:"battery"`
	Position          float32   `edgedb:"position"`
	TargetTemperature float32   `edgedb:"target_temperature"`
	Temperature       float32   `edgedb:"temperature"`
}

func (s ShellyTRVDbModel) Insert(client *edgedb.Client) (*Inserted, error) {
	insertQuery := fmt.Sprintf(`
	INSERT ShellyTRV {
		timestamp := <datetime>$0,
		device := (insert Device { device_id := "%s" } unless conflict on .device_id else (select Device)),
		battery := <float32>$1,
		position := <float32>$2,
		target_temperature := <float32>$3,
		temperature := <float32>$4
	}`, s.Device.DeviceId)

	var inserted Inserted
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := client.QuerySingle(
		ctx,
		insertQuery,
		&inserted,
		s.Timestamp,
		s.Battery,
		s.Position,
		s.TargetTemperature,
		s.Temperature)

	if err != nil {
		return nil, err
	}

	return &inserted, nil
}

func IngestShellyTRV(dbClient *edgedb.Client, trvId string) shelly.ShellyTRV {
	trv := shelly.NewShellyTRV(trvId, getMQTTOpts())
	trv.Connect()

	infoCallback := func(status shelly.ShellyTRVInfo) {
		log.Debug().Interface("status", status).Msg("Received status")

		s := ShellyTRVDbModel{
			Device:            Device{DeviceId: trvId},
			Timestamp:         time.Now().UTC(),
			Battery:           float32(status.Bat.Value),
			Position:          status.Thermostats[0].Pos,
			TargetTemperature: status.Thermostats[0].TargetT.Value,
			Temperature:       status.Thermostats[0].Tmp.Value,
		}
		inserted, err := s.Insert(dbClient)

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
