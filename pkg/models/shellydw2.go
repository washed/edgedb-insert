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

func (s ShellyDW2DbModel) Insert(client *edgedb.Client) (*Inserted, error) {
	insertQuery := fmt.Sprintf(`
	INSERT ShellyDW2 {
		timestamp := <datetime>$0,
		device := (insert Device { device_id := "%s" } unless conflict on .device_id else (select Device)),
		battery := <float32>$1,
		lux := <float32>$2,
		open := <bool>$3,
		temperature := <float32>$4,
		tilt := <int16>$5
	}`, s.Device.DeviceId)

	log.Trace().
		Str("DeviceId", s.Device.DeviceId).
		Str("insertQuery", insertQuery).
		Time("s.Timestamp", s.Timestamp).
		Float32("s.Battery", s.Battery).
		Float32("s.Lux", s.Lux).
		Bool("s.Open", s.Open).
		Float32("s.Temperature", s.Temperature).
		Int16("s.Tilt", s.Tilt).
		Msg("insertQuery")

	var inserted Inserted
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
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

func IngestShellyDW2(dbClient *edgedb.Client, dw2Id string) shelly.ShellyDW2 {
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

		inserted, err := s.Insert(dbClient)

		if err != nil {
			log.Error().Str("DeviceName", dw2.DeviceName()).Err(err).Msg("Error inserting data")
			return
		}
		log.Info().
			Str("DeviceName", dw2.DeviceName()).
			Str("id", inserted.Id.String()).
			Msg("inserted object")
	}

	go dw2.SubscribeInfo(infoCallback)

	return dw2
}
