package models

import (
	"context"
	"fmt"
	"time"

	"github.com/edgedb/edgedb-go"
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

func (s ShellyTRVDbModel) Insert(ctx context.Context, client *edgedb.Client) (*Inserted, error) {
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
