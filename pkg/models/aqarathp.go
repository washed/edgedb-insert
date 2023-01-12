package models

import (
	"time"

	"github.com/edgedb/edgedb-go"
)

type AqaraTHP struct {
	edgedb.Optional
	Timestamp   time.Time            `edgedb:"timestamp"`
	Device      Device               `edgedb:"device"`
	Humidity    float32              `edgedb:"humidity"`
	Linkquality int16                `edgedb:"linkquality"`
	Pressure    float32              `edgedb:"pressure"`
	Temperature float32              `edgedb:"temperature"`
	Battery     edgedb.OptionalInt16 `edgedb:"battery"`
}
