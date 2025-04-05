mapcamera-checker
=================

MapCamera Checker

### Build

```bash
$ make build
```

### Run

```
$ mc keyword
```

`keyword` starts with `keyword=` and is escaped. For example use `keyword=leica%20m11` from the following url: https://www.mapcamera.com/search?keyword=leica%20m11&igngkeyword=1

### Notify

Use `NTFY_TOPIC` environment to enable push notification

### Healthcheck

Use `HC_SLUG` environment to enable healthcheck.io ping.
