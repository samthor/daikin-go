package main

import (
	"github.com/samthor/daikin-go/api"

	"log"
	"time"
)

func main() {
	d, err := api.NewDiscover()
	if err != nil {
		log.Fatalf("couldn't construct discover: %v", err)
	}

	go func() {
		for {
			addr, data, err := d.Next()
			if err != nil {
				log.Fatalf("couldn't read next: %v", err)
			}
			log.Printf("got addr=%v data=%+v", addr, data)
		}
	}()

	for {
		log.Printf("sending probe")
		err := d.Announce()
		if err != nil {
			log.Printf("got err sending to probe: %v", err)
		}
		time.Sleep(time.Second * 2)
	}
}
