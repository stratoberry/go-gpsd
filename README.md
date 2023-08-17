# go-gpsd

*GPSD client for Go.*

[![Go Reference](https://pkg.go.dev/badge/github.com/stratoberry/go-gpsd.svg)](https://pkg.go.dev/github.com/stratoberry/go-gpsd)

## Installation

```
# go get github.com/stratoberry/go-gpsd
```

go-gpsd has no external dependencies.

## Usage

go-gpsd is a streaming client for GPSD's JSON service and as such can be used only in async manner unlike clients for other languages which support both async and sync modes.

```go
import ("github.com/stratoberry/go-gpsd")

func main() {
	gps := gpsd.Dial("localhost:2947")
}
```

After `Dial`ing the server, you should install stream filters. Stream filters allow you to capture only certain types of GPSD reports.

```go
gps.AddFilter("TPV", tpvFilter)
```

Filter functions have a type of `gps.Filter` and should receive one argument of type `interface{}`.

```go
tpvFilter := func(r interface{}) {
	report := r.(*gpsd.TPVReport)
	fmt.Println("Location updated", report.Lat, report.Lon)
}
```

Due to the nature of GPSD reports your filter will manually have to cast the type of the argument it received to a proper `*gpsd.Report` struct pointer.

After installing all needed filters, call the `Watch` method to start observing reports. Please note that at this time installed filters can't be removed.

```go
done := gps.Watch()
<-done
// ...some time later...
gps.Close()
```

`Watch()` spans a new goroutine in which all data processing will happen, `done` doesn't send anything.

### Currently supported GPSD report types

* [`VERSION`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_version) (`gpsd.VERSIONReport`)
* [`TPV`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_tpv) (`gpsd.TPVReport`)
* [`SKY`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_sky) (`gpsd.SKYReport`)
* [`ATT`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_att) (`gpsd.ATTReport`)
* [`GST`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_gst) (`gpsd.GSTReport`)
* [`PPS`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_pps) (`gpsd.PPSReport`)
* [`DEVICES`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_devices) (`gpsd.DEVICESReport`)
* [`DEVICE`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_device_device) (`gpsd.DEVICEReport`)
* [`ERROR`](https://gpsd.gitlab.io/gpsd/gpsd_json.html#_error) (`gpsd.ERRORReport`)

## Documentation

For complete library documentation visit [Go Reference](https://pkg.go.dev/github.com/stratoberry/go-gpsd).

Documentation of GPSD's JSON protocol is available at [https://gpsd.gitlab.io/gpsd/gpsd_json.html](https://gpsd.gitlab.io/gpsd/gpsd_json.html).

## References

This library was originally developed as a part of a student project at [FOI/University of Zagreb](https://www.foi.unizg.hr/en).
