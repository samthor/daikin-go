Raw interface to Daikin AC units.
Provides discovery service (via UDP multicast) and direct API access (via HTTP GET/POST).

Run `logger.go` or `broadcast.go` for simple demos.

## Sample

```go
package main

import (
  "log"
  daikin "github.com/samthor/daikin-go/api"
)

func main() {
	values, err := api.Get("192.168.1.155", "aircon/get_sensor_info")
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("values: %+v", values)
}
```
