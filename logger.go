package main

import (
	"github.com/samthor/daikin-go/api"

	"fmt"
	"log"
	"strconv"
	"time"
)

type basicInfo struct {
	Addr    string
	Name    string
	Group   string
	MAC     string
	Version string
}

type sensorInfo struct {
	Home float32
	Unit *float32
}

type update struct {
	MAC        string
	Error      error
	SensorInfo *sensorInfo
}

func main() {
	d, err := api.NewDiscover()
	if err != nil {
		log.Fatalf("couldn't construct discover: %v", err)
	}

	type initialPayload struct {
		basicInfo
		Power bool
	}

	seenCh := make(chan initialPayload)
	go func() {
		for {
			addr, data, err := d.Next()
			if err != nil {
				log.Fatalf("couldn't read next: %v", err)
			}
			if data.Get("type") != "aircon" {
				continue
			}
			seenCh <- initialPayload{
				Power: data.Get("pow") != "0",
				basicInfo: basicInfo{
					Addr:    addr,
					Name:    data.Get("name"),
					Group:   data.Get("grp_name"),
					MAC:     data.Get("mac"),
					Version: data.Get("ver"),
				},
			}
		}
	}()

	tickCh := make(chan bool)
	go func() {
		// probe lots to start with
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
		}
		for {
			tickCh <- true
			time.Sleep(time.Minute)
		}
	}()

	dataCh := make(chan update)

	sensorFetch := func(info basicInfo) {
		values, err := api.Get(info.Addr, "aircon/get_sensor_info")
		if err != nil {
			dataCh <- update{MAC: info.MAC, Error: err}
		} else {
			si := &sensorInfo{}
			// TODO: htemp might not exist
			htemp, _ := strconv.ParseFloat(values.Get("htemp"), 64)
			si.Home = float32(htemp)
			if otemp, err := strconv.ParseFloat(values.Get("otemp"), 64); err == nil {
				f32 := float32(otemp)
				si.Unit = &f32
			}

			dataCh <- update{MAC: info.MAC, SensorInfo: si}
		}
	}

	known := make(map[string]basicInfo)
	for {
		select {
		case <-tickCh:
			d.Announce()

			for _, info := range known {
				go sensorFetch(info)
			}

		case u := <-dataCh:
			if _, had := known[u.MAC]; !had {
				log.Printf("got update for unknown mac: %+v", u)
				continue
			}

			if u.Error != nil {
				log.Printf("got err: %+v", u)
				delete(known, u.MAC)
				continue
			}

			part := "-"
			if u.SensorInfo.Unit != nil {
				part = fmt.Sprintf("%.1f", *u.SensorInfo.Unit)
			}
			print(fmt.Sprintf("%v\t%.1f\t%s\n", u.MAC, u.SensorInfo.Home, part))

		case payload := <-seenCh:
			_, had := known[payload.MAC]
			known[payload.MAC] = payload.basicInfo

			if had {
				break
			}
			log.Printf("got new: %+v", payload)
		}
	}
}
