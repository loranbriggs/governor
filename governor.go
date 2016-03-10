package main

import (
  "fmt"
  "github.com/loranbriggs/w1thermsensor"
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
  pin  *rpi.Pin
  state State = ready
)

const buffer = 1.0

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
  initPin()
  for {
    currentTemp = w1thermsensor.TemperatureF()
    fmt.Printf("%v\nstate:%v\ncurrent:%v\nminTemp:%v\n", time.Now(), state, currentTemp, minTemp)
    switch state {
    case ready:
      if currentTemp < minTemp-buffer {
        state = heating
        fmt.Println("heat on")
        heatOn()
      } else {
        fmt.Println("heat off")
        heatOff()
      }
    case heating:
      if currentTemp < minTemp+buffer {
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

func initPin() {
  pin, err = rpi.OpenPin(21, rpi.OUT)
  if err != nil {
    panic(err)
  }
  pin.Write(rpi.HIGH)
  fmt.Println("init")
  //defer pin.Close()
}

func heatOn() {
  pin.Write(rpi.LOW)
}

func heatOff() {
  pin.Write(rpi.HIGH)
}

func main() {
  go regulate()

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/set", adjustTempHandler)

  fmt.Println("listening at http://localhost:8080/")
  http.ListenAndServe(":8080", nil)
}
