package bmp085

import (
  "bitbucket.org/gmcbay/i2c"
  "fmt"
  "math"
  "time"
)

type Mode int

const (
  MODE_ULTRALOWPOWER Mode = 0
  MODE_STANDARD      Mode = 1
  MODE_HIGHRES       Mode = 2
  MODE_ULTRAHIGHRES  Mode = 3
)

type Register byte

const (
  REG_AC1 Register = 0xAA
  REG_AC2 Register = 0xAC
  REG_AC3 Register = 0xAE
  REG_AC4 Register = 0xB0
  REG_AC5 Register = 0xB2
  REG_AC6 Register = 0xB4
  REG_B1  Register = 0xB6
  REG_B2  Register = 0xB8
  REG_MB  Register = 0xBA
  REG_MC  Register = 0xBC
  REG_MD  Register = 0xBE
)

const (
  CONTROL           = 0xF4
  TEMP_DATA         = 0xF6
  PRESSURE_DATA     = 0xF6
  READ_TEMP_CMD     = 0x2E
  READ_PRESSURE_CMD = 0x34
)

type Device struct {
  bus                                                       *i2c.I2CBus
  Addr, BusNum                                              byte
  Mode                                                      Mode
  RegAC1, RegAC2, RegAC3, RegB1, RegB2, RegMB, RegMC, RegMD int16
  RegAC4, RegAC5, RegAC6                                    uint16
}

func Init(addr, busNum byte, mode Mode) (dev *Device, err error) {
  dev = new(Device)
  err = dev.Init(addr, busNum, mode)
  return
}

func (dev *Device) Init(addr, busNum byte, mode Mode) (err error) {
  if dev.bus, err = i2c.Bus(busNum); err != nil {
    return
  }
  dev.Addr = addr
  dev.BusNum = busNum
  dev.Mode = mode
  dev.readCalibrationData()

  return
}

func (dev *Device) readCalibrationData() {
  dev.RegAC1, _ = dev.readInt16(REG_AC1)
  dev.RegAC2, _ = dev.readInt16(REG_AC2)
  dev.RegAC3, _ = dev.readInt16(REG_AC3)
  dev.RegAC4, _ = dev.readUint16(REG_AC4)
  dev.RegAC5, _ = dev.readUint16(REG_AC5)
  dev.RegAC6, _ = dev.readUint16(REG_AC6)
  dev.RegB1, _ = dev.readInt16(REG_B1)
  dev.RegB2, _ = dev.readInt16(REG_B2)
  dev.RegMB, _ = dev.readInt16(REG_MB)
  dev.RegMC, _ = dev.readInt16(REG_MC)
  dev.RegMD, _ = dev.readInt16(REG_MD)
}

func (dev *Device) PrintCalibrationData() {
  fmt.Println("AC1", dev.RegAC1)
  fmt.Println("AC2", dev.RegAC2)
  fmt.Println("AC3", dev.RegAC3)
  fmt.Println("AC4", dev.RegAC4)
  fmt.Println("AC5", dev.RegAC5)
  fmt.Println("AC6", dev.RegAC6)

  fmt.Println("B1", dev.RegB1)
  fmt.Println("B2", dev.RegB2)

  fmt.Println("MB", dev.RegMB)
  fmt.Println("MC", dev.RegMC)
  fmt.Println("MD", dev.RegMD)
}

func (dev *Device) readRawTemperature() (temperature uint, err error) {
  if err = dev.bus.WriteByte(dev.Addr, CONTROL, READ_TEMP_CMD); err != nil {
    return
  }

  time.Sleep(5 * time.Millisecond)
  var temp uint16
  if temp, err = dev.readUint16(TEMP_DATA); err != nil {
    return
  }

  temperature = uint(temp)

  return
}

func (dev *Device) readRawPressure() (pressure uint, err error) {
  modeModifier := byte(dev.Mode) << 6
  if err = dev.bus.WriteByte(dev.Addr, CONTROL, READ_PRESSURE_CMD+modeModifier); err != nil {
    return
  }

  switch dev.Mode {
  default:
    time.Sleep(8 * time.Millisecond)
  case MODE_ULTRALOWPOWER:
    time.Sleep(5 * time.Millisecond)
  case MODE_HIGHRES:
    time.Sleep(14 * time.Millisecond)
  case MODE_ULTRAHIGHRES:
    time.Sleep(26 * time.Millisecond)
  }

  var list []byte
  if list, err = dev.bus.ReadByteBlock(dev.Addr, byte(PRESSURE_DATA), 3); err != nil {
    return
  }

  msb, lsb, xlsb := uint(list[0]), uint(list[1]), uint(list[2])
  pressure = ((msb << 16) + (lsb << 8) + xlsb) >> (8 - uint(dev.Mode))

  return
}

func (dev *Device) GetData() (temperature float64, pressure uint, altitude float64, err error) {
  var ut, up uint
  if ut, err = dev.readRawTemperature(); err != nil {
    return
  }
  if up, err = dev.readRawPressure(); err != nil {
    return
  }

  // temperature calculaton
  x1 := ((ut - uint(dev.RegAC6)) * uint(dev.RegAC5)) >> 15
  x2 := (int(dev.RegMC) << 11) / (int(x1) + int(dev.RegMD))
  b5 := int(x1) + x2
  temperature = float64((b5+8)>>4) / 10.0

  // pressure calculations
  b6 := b5 - 4000
  x1_2 := (int(dev.RegB2) * (b6 * b6) >> 12) >> 11
  x2_2 := (int(dev.RegAC2) * b6) >> 11
  x3_2 := x1_2 + x2_2
  b3 := (((int(dev.RegAC1)*4 + x3_2) << uint(dev.Mode)) + 2) / 4

  x1_2 = (int(dev.RegAC3) * b6) >> 13
  x2_2 = (int(dev.RegB1) * ((b6 * b6) >> 12)) >> 16
  x3_2 = (x1_2 + x2_2 + 2) >> 2
  b4 := (int(dev.RegAC4) * (x3_2 + 32768)) >> 15
  b7 := (int(up) - b3) * (50000 >> uint(dev.Mode))

  if uint(b7) < uint(0x80000000) {
    pressure = uint(b7*2) / uint(b4)
  } else {
    pressure = (uint(b7) / uint(b4)) * 2
  }

  x1 = (pressure >> 8) * (pressure >> 8)
  x1 = (x1 * 3038) >> 16
  x2_2 = (-7357 * int(pressure)) >> 16

  pressure = pressure + uint((int(x1)+x2_2+3791)>>4)

  altitude = 44330.0 * (1.0 - math.Pow(float64(pressure)/101325, 0.1903))

  return
}

func (dev *Device) GetTemperature() (temperature float64, err error) {
  temperature, _, _, err = dev.GetData()
  return
}

func (dev *Device) GetPressure() (pressure uint, err error) {
  _, pressure, _, err = dev.GetData()
  return
}

func (dev *Device) GetAltitude() (altitude float64, err error) {
  _, _, altitude, err = dev.GetData()
  return
}

func (dev *Device) readByte(reg Register) (ret byte, err error) {
  list, err := dev.bus.ReadByteBlock(dev.Addr, byte(reg), 1)
  if err != nil {
    return
  }

  ret = list[0]
  return
}

func (dev *Device) readInt16(reg Register) (ret int16, err error) {
  list, err := dev.bus.ReadByteBlock(dev.Addr, byte(reg), 2)

  hi := signByte(int16(list[0]))
  lo := int16(list[1])

  ret = (hi << 8) + lo

  return
}

func (dev *Device) readUint16(reg Register) (ret uint16, err error) {
  list, err := dev.bus.ReadByteBlock(dev.Addr, byte(reg), 2)

  hi := uint16(list[0])
  lo := uint16(list[1])

  ret = (hi << 8) + lo

  return
}

func signByte(n int16) (signedN int16) {
  if n > 127 {
    signedN = n - 256
  } else {
    signedN = n
  }

  return
}
