package models

import (
	"context"
	"fmt"
	"time"

	"github.com/edgedb/edgedb-go"
	"github.com/rs/zerolog/log"
	"github.com/washed/shelly-go"
)

type ShellyDW2DbModel struct {
	edgedb.Optional
	Timestamp   time.Time `edgedb:"timestamp"`
	Device      Device    `edgedb:"device"`
	Battery     float32   `edgedb:"battery"`
	Lux         float32   `edgedb:"lux"`
	Open        bool      `edgedb:"open"`
	Temperature float32   `edgedb:"temperature"`
	Tilt        int16     `edgedb:"tilt"`
}

func (s ShellyDW2DbModel) Insert(ctx context.Context, client *edgedb.Client) (*Inserted, error) {
	insertQuery := fmt.Sprintf(`
	INSERT ShellyTRV {
		timestamp := <datetime>$0,
		device := (insert Device { device_id := "%s" } unless conflict on .device_id else (select Device)),
		battery := <float32>$1,
		lux := <float32>$2,
		open := <bool>$3,
		temperature := <float32>$4,
		tilt := <int16>$5
	}`, s.Device.DeviceId)

	var inserted Inserted
	err := client.QuerySingle(
		ctx,
		insertQuery,
		&inserted,
		s.Timestamp,
		s.Battery,
		s.Lux,
		s.Open,
		s.Temperature,
		s.Tilt)

	if err != nil {
		return nil, err
	}

	return &inserted, nil
}

func IngestShellyDW2(ctx context.Context, dbClient *edgedb.Client, dw2Id string) shelly.ShellyDW2 {
	dw2 := shelly.NewShellyDW2(dw2Id, getMQTTOpts())
	dw2.Connect()

	infoCallback := func(info shelly.ShellyDW2Info) {
		log.Debug().Interface("info", info).Msg("received info")

		s := ShellyDW2DbModel{
			Device:      Device{DeviceId: dw2Id},
			Timestamp:   time.Now().UTC(),
			Battery:     float32(info.Bat.Value),
			Lux:         info.Lux.Value,
			Open:        info.Sensor.IsOpen(),
			Temperature: info.Tmp.Value,
			Tilt:        int16(info.Accel.Tilt)}

		inserted, err := s.Insert(ctx, dbClient)

		if err != nil {
			log.Error().Str("DeviceName", dw2.DeviceName()).Err(err).Msg("Error inserting data")
		}
		log.Info().
			Str("DeviceName", dw2.DeviceName()).
			Str("id", inserted.Id.String()).
			Msg("inserted object")
	}

	go dw2.SubscribeInfo(infoCallback)

	return dw2
}
