package main

import "fmt"
import "go-gpsd"

func main() {
	var gps *gpsd.Session
	var err error

	if gps, err = gpsd.Dial(gpsd.DefaultAddress); err != nil {
		panic(fmt.Sprintf("Failed to connect to GPSD: %s", err))
	}

	gps.AddFilter("TPV", func(r interface{}) {
		tpv := r.(*gpsd.TPVReport)
		fmt.Println("TPV", tpv.Mode, tpv.Time)
	})

	skyfilter := func(r interface{}) {
		sky := r.(*gpsd.SKYReport)

		fmt.Println("SKY", len(sky.Satellites), "satellites")
	}

	gps.AddFilter("SKY", skyfilter)

	//Handle errors
	errChan := gps.Watch()
	for {
		if err := <-errChan; err.Error != nil {
			fmt.Println(err.Message)
		} else {
			fmt.Println("OK")
		}
	}
}
