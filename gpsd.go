package gpsd

import (
  "bufio"
  "encoding/json"
  "fmt"
  "net"
  "time"
)

const DefaultAddress = "localhost:2947"

type Filter func(interface{})

type Session struct {
  socket  net.Conn
  reader  *bufio.Reader
  filters map[string][]Filter
}

type Mode byte

const (
  NoValueSeen Mode = 0
  NoFix       Mode = 1
  Mode2D      Mode = 2
  Mode3D      Mode = 3
)

type GPSDReport struct {
  Class string `json:"class"`
}

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

type VERSIONReport struct {
  Class      string `json:"class"`
  Release    string `json:"release"`
  Rev        string `json:"rev"`
  ProtoMajor int    `json:"proto_major"`
  ProtoMinor int    `json:"proto_minor"`
  Remote     string `json:"remote"`
}

type DEVICESReport struct {
  Class   string         `json:"class"`
  Devices []DEVICEReport `json:"devices"`
  Remote  string         `json:"remote"`
}

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

type PPSReport struct {
  Class      string  `json:"class"`
  Device     string  `json:"device"`
  RealSec    float64 `json:"real_sec"`
  RealMusec  float64 `json:"real_musec"`
  ClockSec   float64 `json:"clock_sec"`
  ClockMusec float64 `json:"clock_musec"`
}

type ERRORReport struct {
  Class   string `json:"class"`
  Message string `json:"message"`
}

type Satellite struct {
  PRN  float64 `json:"PRN"`
  Az   float64 `json:"az"`
  El   float64 `json:"el"`
  Ss   float64 `json:"ss"`
  Used bool    `json:"used"`
}

// Dial opens a new connection to GPSD.
func Dial(address string) (session *Session, err error) {
  session = new(Session)
  session.socket, err = net.Dial("tcp4", address)

  session.reader = bufio.NewReader(session.socket)
  line, _ := session.reader.ReadString('\n')
  _ = line
  session.filters = make(map[string][]Filter)

  return
}

// Starts watching GPSD reports in a new goroutine.
//
// Example
//    gps := gpsd.Dial(gpsd.DEFAULT_ADDRESS)
//    done := gpsd.Watch()
//    <- done
func (s *Session) Watch() (done chan bool) {
  fmt.Fprintf(s.socket, "?WATCH={\"enable\":true,\"json\":true}")
  done = make(chan bool)

  go watch(done, s)

  return
}

func (s *Session) SendCommand(command string) {
  fmt.Fprintf(s.socket, "?"+command+";")
}

// AddFilter attaches a function which will be called for all
// GPSD reports with the given class. Callback functions have type Filter.
//
// Example:
//    gps := gpsd.Init(gpsd.DEFAULT_ADDRESS)
//    gps.AddFilter("TPV", func (r interface{}) {
//      report := r.(*gpsd.TPVReport)
//      fmt.Println(report.Time, report.Lat, report.Lon)
//    })
//    done := gps.Watch()
//    <- done
func (s *Session) AddFilter(class string, f Filter) {
  s.filters[class] = append(s.filters[class], f)
}

func (s *Session) deliverReport(class string, report interface{}) {
  for _, f := range s.filters[class] {
    f(report)
  }
}

func watch(done chan bool, s *Session) {
  fmt.Println("WATCHING")

  // We're not using a JSON decoder because we first need to inspect
  // the JSON string to determine it's "class"
  for {
    if line, err := s.reader.ReadString('\n'); err == nil {
      var reportPeek GPSDReport
      lineBytes := []byte(line)
      if err = json.Unmarshal(lineBytes, &reportPeek); err == nil {
        if len(s.filters[reportPeek.Class]) == 0 {
          continue
        }

        if report, err := unmarshalReport(reportPeek.Class, lineBytes); err == nil {
          s.deliverReport(reportPeek.Class, report)
        } else {
          fmt.Println("JSON parsing error 2:", err)
        }
      } else {
        fmt.Println("JSON parsing error:", err)
      }
    } else {
      fmt.Println("Stream reader error:", err)
    }
  }
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
  case "ERROR":
    var r *ERRORReport
    err = json.Unmarshal(bytes, &r)
    return r, err
  }

  return nil, err
}
