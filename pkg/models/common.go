package models

import (
	"github.com/edgedb/edgedb-go"
)

type Inserted struct {
	Id edgedb.UUID `edgedb:"id"`
}

type Device struct {
	Name       edgedb.OptionalStr `edgedb:"name"`
	DeviceId   string             `edgedb:"device_id"`
	DeviceType edgedb.OptionalStr `edgedb:"device_type"`
	// ? multi link instances := (.<device[is default::HasDeviceLink]);
}
