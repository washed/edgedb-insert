package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	models "edgedb-ingest/pkg/models"

	"github.com/edgedb/edgedb-go"
)

func main() {
	ctx := context.Background()
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
}
