# edgedb-insert

## Config

Name it `config.yaml` and put in exeuctable directory.

```yaml
log_level: trace
shelly_trv_ids:
  - ids
  - ...
```

| Field              | Type            | values                                                 | Description                                                     |
| ------------------ | --------------- | ------------------------------------------------------ | --------------------------------------------------------------- |
| **log_level**      | string          | one of [trace, debug, info, warn, error, fatal, panic] | Confgures the logger to the specified level. Defaults to "info" |
| **shelly_trv_ids** | list of strings |                                                        | ShellyTRV IDs (MACs) to ingest                                  |
| **shelly_dw2_ids** | list of strings |                                                        | ShellyDW2 IDs (MACs) to ingest                                  |
