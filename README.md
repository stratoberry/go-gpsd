[![GoDoc](https://godoc.org/github.com/stratoberry/go-gpsd?status.svg)](https://godoc.org/github.com/stratoberry/go-gpsd)

# go-gpsd

*GPSD client for Go.*

## Installation

<pre><code># go get github.com/stratoberry/go-gpsd</code></pre>

go-gpsd has no external dependencies.

## Usage

go-gpsd is a streaming client for GPSD's JSON service and as such can be used only in async manner unlike clients for other languages which support both async and sync modes.
```golang
import ("github.com/stratoberry/go-gpsd")

func main() {
	gps := gpsd.Dial("localhost:2947")
}
```

After `Dial`ing the server, you should install stream filters. Stream filters allow you to capture only certain types of GPSD reports.
```golang
gps.AddFilter("TPV", tpvFilter)
```

Filter functions have a type of `gps.Filter` and should receive one argument of type `interface{}`.
```golang
tpvFilter := func(r interface{}) {
	report := r.(*gpsd.TPVReport)
	fmt.Println("Location updated", report.Lat, report.Lon)
}
```

Due to the nature of GPSD reports your filter will manually have to cast the type of the argument it received to a proper `*gpsd.Report` struct pointer.

After installing all needed filters, call the `Watch` method to start observing reports. Please note that at this time installed filters can't be removed.
```golang
errChan := gps.Watch()
```

`Watch()` will span a new goroutine in which all data processing will happen, `errChan` channel will send errors or `nil`. Error handling:

```golang
for {
	if err := <-errChan; err.Error != nil {
		fmt.Println(err.Message)
	} else {
		fmt.Println("OK")
	}
}
```

### Currently supported GPSD report types

* `VERSION` (`gpsd.VERSIONReport`)
* `TPV` (`gpsd.TPVReport`)
* `SKY` (`gpsd.SKYReport`)
* `ATT` (`gpsd.ATTReport`)
* `GST` (`gpsd.GSTReport`)
* `PPS` (`gpsd.PPSReport`)
* `Devices` (`gpsd.DEVICESReport`)
* `DEVICE` (`gpsd.DEVICEReport`)
* `ERROR` (`gpsd.ERRORReport`)

## Documentation

For complete library docs, visit [GoDoc.org](http://godoc.org/github.com/stratoberry/go-gpsd) or take a look at the `gpsd.go` file in this repository.

GPSD's documentation on their JSON protocol can be found at [http://catb.org/gpsd/gpsd_json.html](http://catb.org/gpsd/gpsd_json.html)

To learn more about the Stratoberry Pi project, visit our website at [stratoberry.foi.hr](http://stratoberry.foi.hr).


## License

