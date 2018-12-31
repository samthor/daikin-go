package main

import (
	"github.com/samthor/daikin-go/api"

	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	flagPort     = flag.Int("port", 9000, "default port")
	flagAnnounce = flag.Duration("announce", time.Second*30, "time between announcements")
	flagUpdate   = flag.Duration("update", time.Second*10, "time between updates")

	vz = &VarzHandler{}
)

type device struct {
	addr string
	data url.Values
	seen time.Time
}

type deviceData struct {
	ci api.ControlInfo
	si api.SensorInfo
}

func (d device) get() (dd deviceData, err error) {
	var values url.Values
	values, err = api.Get(d.addr, "aircon/get_control_info")
	if err != nil {
		return
	}
	dd.ci = api.ParseControlInfo(values)

	values, err = api.Get(d.addr, "aircon/get_sensor_info")
	dd.si = api.ParseSensorInfo(values)
	return
}

func main() {
	flag.Parse()

	d, err := api.NewDiscover()
	if err != nil {
		log.Fatalf("couldn't construct discover: %v", err)
	}

	updateVarz := func(mac string, dd *deviceData) {
		u := func(part string, v interface{}) {
			key := fmt.Sprintf("aircon.%s.%s", mac, part)
			vz.Update(key, v)
		}
		if dd == nil {
			u("target", nil)
			u("temp", nil)
			u("world", nil)
		} else {
			if dd.ci.Power {
				u("target", dd.ci.Temp)
			} else {
				u("target", nil)
			}
			u("temp", dd.si.Temp)
			if dd.si.World != nil {
				u("world", *dd.si.World)
			} else {
				u("world", nil)
			}
		}
	}

	deviceCh := make(chan device)
	go func() {
		devices := make(map[string]device)
		t := time.NewTicker(*flagUpdate)
		for {
			select {
			case d := <-deviceCh:
				mac := d.data.Get("mac")
				if _, ok := devices[mac]; ok {
					continue // already got this one
				}

				dd, err := d.get()
				if err != nil {
					log.Printf("failed to fetch initial update (mac=%v): %v", mac, err)
					continue
				}

				log.Printf("got addr=%v mac=%v name=%v", d.addr, d.data.Get("mac"), d.data.Get("name"))
				devices[mac] = d
				updateVarz(mac, &dd)

			case <-t.C:
				for mac, d := range devices {
					dd, err := d.get()
					if err != nil {
						log.Printf("device err (mac=%v): %v", mac, err)
						delete(devices, mac)
						updateVarz(mac, nil)
						continue
					}
					updateVarz(mac, &dd)
				}

			}
		}

	}()

	go func() {
		for {
			addr, data, err := d.Next()
			if err != nil {
				log.Fatalf("couldn't read next: %v", err)
			} else if data.Get("mac") == "" || data.Get("type") != "aircon" {
				log.Printf("unexpected device: %+v", data)
			}
			deviceCh <- device{addr: addr, data: data}
		}
	}()

	go func() {
		for {
			err := d.Announce()
			if err != nil {
				log.Fatalf("got err sending to probe: %v", err)
			}
			time.Sleep(*flagAnnounce)
		}
	}()

	http.Handle("/varz", vz)
	log.Printf("listening on :%d...", *flagPort)
	http.ListenAndServe(fmt.Sprintf(":%d", *flagPort), nil)
}
