package gpsd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// DefaultAddress of gpsd (localhost:2947)
const DefaultAddress = "localhost:2947"

// Filter is a gpsd entry filter function
type Filter func(interface{})

// Session represents a connection to gpsd
type Session struct {
	socket  net.Conn
	reader  *bufio.Reader
	filters map[string][]Filter
}

// Mode describes status of a TPV report
type Mode byte

const (
	// NoValueSeen indicates no data has been received yet
	NoValueSeen Mode = 0
	// NoFix indicates fix has not been required yet
	NoFix Mode = 1
	// Mode2D represents quality of the fix
	Mode2D Mode = 2
	// Mode3D represents quality of the fix
	Mode3D Mode = 3
)

type gpsdReport struct {
	Class string `json:"class"`
}

// TPVReport is a Time-Position-Velocity report
type TPVReport struct {
	Class  string    `json:"class"`
	Tag    string    `json:"tag"`
	Device string    `json:"device"`
	Mode   Mode      `json:"mode"`
	Time   time.Time `json:"time"`
	Ept    float64   `json:"ept"`
	Lat    float64   `json:"lat"`
	Lon    float64   `json:"lon"`
	Alt    float64   `json:"alt"`
	Epx    float64   `json:"epx"`
	Epy    float64   `json:"epy"`
	Epv    float64   `json:"epv"`
	Track  float64   `json:"track"`
	Speed  float64   `json:"speed"`
	Climb  float64   `json:"climb"`
	Epd    float64   `json:"epd"`
	Eps    float64   `json:"eps"`
	Epc    float64   `json:"epc"`
}

// SKYReport reports sky view of GPS satellites
type SKYReport struct {
	Class      string      `json:"class"`
	Tag        string      `json:"tag"`
	Device     string      `json:"device"`
	Time       time.Time   `json:"time"`
	Xdop       float64     `json:"xdop"`
	Ydop       float64     `json:"ydop"`
	Vdop       float64     `json:"vdop"`
	Tdop       float64     `json:"tdop"`
	Hdop       float64     `json:"hdop"`
	Pdop       float64     `json:"pdop"`
	Gdop       float64     `json:"gdop"`
	Satellites []Satellite `json:"satellites"`
}

// GSTReport is pseudorange noise report
type GSTReport struct {
	Class  string    `json:"class"`
	Tag    string    `json:"tag"`
	Device string    `json:"device"`
	Time   time.Time `json:"time"`
	Rms    float64   `json:"rms"`
	Major  float64   `json:"major"`
	Minor  float64   `json:"minor"`
	Orient float64   `json:"orient"`
	Lat    float64   `json:"lat"`
	Lon    float64   `json:"lon"`
	Alt    float64   `json:"alt"`
}

// ATTReport reports vehicle-attitude from the digital compass or the gyroscope
type ATTReport struct {
	Class       string    `json:"class"`
	Tag         string    `json:"tag"`
	Device      string    `json:"device"`
	Time        time.Time `json:"time"`
	Heading     float64   `json:"heading"`
	MagSt       string    `json:"mag_st"`
	Pitch       float64   `json:"pitch"`
	PitchSt     string    `json:"pitch_st"`
	Yaw         float64   `json:"yaw"`
	YawSt       string    `json:"yaw_st"`
	Roll        float64   `json:"roll"`
	RollSt      string    `json:"roll_st"`
	Dip         float64   `json:"dip"`
	MagLen      float64   `json:"mag_len"`
	MagX        float64   `json:"mag_x"`
	MagY        float64   `json:"mag_y"`
	MagZ        float64   `json:"mag_z"`
	AccLen      float64   `json:"acc_len"`
	AccX        float64   `json:"acc_x"`
	AccY        float64   `json:"acc_y"`
	AccZ        float64   `json:"acc_z"`
	GyroX       float64   `json:"gyro_x"`
	GyroY       float64   `json:"gyro_y"`
	Depth       float64   `json:"depth"`
	Temperature float64   `json:"temperature"`
}

// VERSIONReport returns version details of gpsd client
type VERSIONReport struct {
	Class      string `json:"class"`
	Release    string `json:"release"`
	Rev        string `json:"rev"`
	ProtoMajor int    `json:"proto_major"`
	ProtoMinor int    `json:"proto_minor"`
	Remote     string `json:"remote"`
}

// DEVICESReport lists all devices connected to the system
type DEVICESReport struct {
	Class   string         `json:"class"`
	Devices []DEVICEReport `json:"devices"`
	Remote  string         `json:"remote"`
}

// DEVICEReport reports a state of a particular device
type DEVICEReport struct {
	Class     string  `json:"class"`
	Path      string  `json:"path"`
	Activated string  `json:"activated"`
	Flags     int     `json:"flags"`
	Driver    string  `json:"driver"`
	Subtype   string  `json:"subtype"`
	Bps       int     `json:"bps"`
	Parity    string  `json:"parity"`
	Stopbits  string  `json:"stopbits"`
	Native    int     `json:"native"`
	Cycle     float64 `json:"cycle"`
	Mincycle  float64 `json:"mincycle"`
}

// PPSReport is triggered on each pulse-per-second strobe from a device
type PPSReport struct {
	Class      string  `json:"class"`
	Device     string  `json:"device"`
	RealSec    float64 `json:"real_sec"`
	RealMusec  float64 `json:"real_musec"`
	ClockSec   float64 `json:"clock_sec"`
	ClockMusec float64 `json:"clock_musec"`
}

// TOFFReport is triggered on each PPS strobe from a device
type TOFFReport struct {
	Class     string  `json:"class"`
	Device    string  `json:"device"`
	RealSec   float64 `json:"real_sec"`
	RealNSec  float64 `json:"real_nsec"`
	ClockSec  float64 `json:"clock_sec"`
	ClockNSec float64 `json:"clock_nsec"`
}

// ERRORReport is an error response
type ERRORReport struct {
	Class   string `json:"class"`
	Message string `json:"message"`
}

// Satellite describes a location of a GPS satellite
type Satellite struct {
	PRN  float64 `json:"PRN"`
	Az   float64 `json:"az"`
	El   float64 `json:"el"`
	Ss   float64 `json:"ss"`
	Used bool    `json:"used"`
}

// Dial opens a new connection to GPSD.
func Dial(address string) (*Session, error) {
	return dialCommon(net.Dial("tcp4", address))
}

// DialTimeout opens a new connection to GPSD with a timeout.
func DialTimeout(address string, to time.Duration) (*Session, error) {
	return dialCommon(net.DialTimeout("tcp4", address, to))
}

func dialCommon(c net.Conn, err error) (session *Session, e error) {
	session = new(Session)
	session.socket = c
	if err != nil {
		return nil, err
	}

	session.reader = bufio.NewReader(session.socket)
	session.reader.ReadString('\n')
	session.filters = make(map[string][]Filter)

	return
}

// Watch starts watching GPSD reports in a new goroutine.
//
// Example:
//
//	gps := gpsd.Dial(gpsd.DEFAULT_ADDRESS)
//	done := gpsd.Watch()
//	<- done
func (s *Session) Watch() (done chan bool) {
	fmt.Fprintf(s.socket, "?WATCH={\"enable\":true,\"json\":true}")
	done = make(chan bool)

	go watch(done, s)

	return
}

// SendCommand sends a command to GPSD
func (s *Session) SendCommand(command string) {
	fmt.Fprintf(s.socket, "?"+command+";")
}

// AddFilter attaches a function which will be called for all
// GPSD reports with the given class. Callback functions have type Filter.
//
// Example:
//
//	gps := gpsd.Init(gpsd.DEFAULT_ADDRESS)
//	gps.AddFilter("TPV", func (r interface{}) {
//	  report := r.(*gpsd.TPVReport)
//	  fmt.Println(report.Time, report.Lat, report.Lon)
//	})
//	done := gps.Watch()
//	<- done
func (s *Session) AddFilter(class string, f Filter) {
	s.filters[class] = append(s.filters[class], f)
}

func (s *Session) deliverReport(class string, report interface{}) {
	for _, f := range s.filters[class] {
		f(report)
	}
}

// Close closes the connection to GPSD
func (s *Session) Close() error {
	if s.socket == nil {
		return errors.New("gpsd socket is alerady closed")
	}

	if err := s.socket.Close(); err != nil {
		return err
	}

	s.socket = nil
	return nil
}

func watch(done chan bool, s *Session) {
	// We're not using a JSON decoder because we first need to inspect
	// the JSON string to determine it's "class"
	for {
		if line, err := s.reader.ReadString('\n'); err == nil {
			var reportPeek gpsdReport
			lineBytes := []byte(line)
			if err = json.Unmarshal(lineBytes, &reportPeek); err == nil {
				if len(s.filters[reportPeek.Class]) == 0 {
					continue
				}

				if report, err2 := unmarshalReport(reportPeek.Class, lineBytes); err2 == nil {
					s.deliverReport(reportPeek.Class, report)
				} else {
					fmt.Println("JSON parsing error 2:", err)
				}
			} else {
				fmt.Println("JSON parsing error:", err)
			}
		} else {
			if !errors.Is(err, net.ErrClosed) {
				fmt.Println("Stream reader error (is gpsd running?):", err)
			}
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				break
			}
		}
	}
	done <- true
}

func unmarshalReport(class string, bytes []byte) (interface{}, error) {
	var err error

	switch class {
	case "TPV":
		var r *TPVReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "SKY":
		var r *SKYReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "GST":
		var r *GSTReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "ATT":
		var r *ATTReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "VERSION":
		var r *VERSIONReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "DEVICES":
		var r *DEVICESReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "PPS":
		var r *PPSReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "TOFF":
		var r *TOFFReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	case "ERROR":
		var r *ERRORReport
		err = json.Unmarshal(bytes, &r)
		return r, err
	}

	return nil, err
}
