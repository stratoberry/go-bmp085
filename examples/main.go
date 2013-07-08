package main

import (
  "fmt"
  "github.com/stratoberry/go-bmp085"
)

func main() {
  var dev *bmp085.Device
  var err error
  if dev, err = bmp085.Init(0x77, 0, bmp085.MODE_STANDARD); err != nil {
    panic(fmt.Sprintf("Failed to init device", err))
  }

  if temp, pressure, alt, err := dev.GetData(); err != nil {
    panic(fmt.Sprintf("Failed to get data from the device", err))
  } else {
    fmt.Println(fmt.Sprintf("Temperature: %.2f degC", temp))
    fmt.Println(fmt.Sprintf("Pressure: %.2f hPa", float64(pressure)/100))
    fmt.Println(fmt.Sprintf("Altitude: %.2f m", alt))
  }
}
