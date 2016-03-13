package main

import (
  "fmt"
  "github.com/loranbriggs/go-w1sensor"
  "github.com/nathan-osman/go-rpigpio"
  "html/template"
  "net/http"
  "strconv"
  "time"
)

var (
  currentTemp float64
  minTemp float64 = 70.0
  maxTemp float64
  err error
  heaterPin  *rpi.Pin
  tempSensor *w1sensor.Sensor
  state State = ready
)

const buffer = 2.0

type State int

const (
  cooling = iota
  heating = iota
  ready = iota
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
  data := struct {
    CurrentTemp float64
    DesiredTemp float64
  } {
    currentTemp,
    minTemp,
  }

  t, _ := template.ParseFiles("views/home.html")
  t.Execute(w, data)
}

func adjustTempHandler(w http.ResponseWriter, r *http.Request) {
  newTemp, err := strconv.ParseFloat(r.FormValue("newTemp"), 64)
  if err == nil {
    if newTemp < 90 || newTemp > 40 {
      minTemp = newTemp
    }
  }
  http.Redirect(w, r, "/", http.StatusFound)
}

func regulate() {
  initHeaterPin()
  initTempSensor()
  for {
    currentTemp, err = tempSensor.ReadF()
    if err != nil {
      fmt.Print(err)
      continue
    }
    fmt.Printf("%v\nstate:%v\ncurrent:%v\nminTemp:%v\n", time.Now(), state, currentTemp, minTemp)
    switch state {
    case ready:
      if currentTemp < minTemp {
        state = heating
        fmt.Println("heat on")
        heatOn()
      } else {
        fmt.Println("heat off")
        heatOff()
      }
    case heating:
      if currentTemp < minTemp + buffer {
        fmt.Println("heat on")
        heatOn()
      } else {
        state = ready
        fmt.Println("heat off")
        heatOff()
      }
    case cooling:
      // TODO: implement cooling functionality before summer
      state = ready
    }
    fmt.Println()
    time.Sleep(10*time.Second)
  }
}

func initTempSensor() {
  tempSensor, err = w1sensor.FirstAvailableSensor()
  if err != nil {
    panic(err)
  }
  fmt.Println("init temp sensor")
}

func initHeaterPin() {
  heaterPin, err = rpi.OpenPin(21, rpi.OUT)
  if err != nil {
    panic(err)
  }
  heaterPin.Write(rpi.HIGH)
  fmt.Println("init heater pin")
}

func heatOn() {
  heaterPin.Write(rpi.LOW)
}

func heatOff() {
  heaterPin.Write(rpi.HIGH)
}

func main() {
  go regulate()

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/set", adjustTempHandler)

  fmt.Println("listening at http://localhost:8080/")
  http.ListenAndServe(":8080", nil)
}
