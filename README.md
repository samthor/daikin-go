Raw interface to Daikin AC units.
Provides discovery service (via UDP multicast) and direct API access (via HTTP GET/POST).

Method descriptions can be found [here](https://github.com/ael-code/daikin-control/wiki/API-System), or elsewhere online.

## Sample

Run `logger.go` or `broadcast.go` for simple demos.

Or, to dial a specific URL and get its sensor info:

```go
package main

import (
  "log"
  daikin "github.com/samthor/daikin-go/api"
)

func main() {
  values, err := daikin.Get("192.168.1.155", "aircon/get_sensor_info")
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("values: %+v", values)
}
```
