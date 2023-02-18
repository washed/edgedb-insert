package config

import ks "github.com/washed/kitchen-sink-go"

type Config struct {
	Log          ks.LogConfig `yaml:"log"`
	ShellyTRVIDs []string     `yaml:"shelly_trv_ids"`
	ShellyDW2IDs []string     `yaml:"shelly_dw2_ids"`
}
