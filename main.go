package main

import (
  "flag"
  "fmt"
  "strings"
  "strconv"
  "io/ioutil"

  mp "github.com/mackerelio/go-mackerel-plugin-helper"

  "github.com/kidoman/embd"
  _ "github.com/kidoman/embd/host/rpi"
  "github.com/taiyoh/go-embd-bme280"
)

type SensorPlugin struct {
  Prefix string
  BME280 *bme280.BME280
}

func (s SensorPlugin) GraphDefinition() map[string](mp.Graphs) {
  return map[string](mp.Graphs){
    s.Prefix + ".temperature": mp.Graphs{
      Label: "Sensor Temperature(℃)",
      Unit:  "float",
      Metrics: [](mp.Metrics){
        mp.Metrics{Name: "cpu_temperature", Label: "CPU"},
        mp.Metrics{Name: "bme280_temperature", Label: "BME280"},
      },
    },
    s.Prefix + ".pressure": mp.Graphs{
      Label: "Sensor Pressure(hPa)",
      Unit:  "float",
      Metrics: [](mp.Metrics){
        mp.Metrics{Name: "bme280_pressure", Label: "BME280"},
      },
    },
    s.Prefix + ".humidity": mp.Graphs{
      Label: "Sensor Humidity(%)",
      Unit:  "float",
      Metrics: [](mp.Metrics){
        mp.Metrics{Name: "bme280_humidity", Label: "BME280"},
      },
    },
  }
}

func (s SensorPlugin) FetchMetrics() (map[string]interface{}, error) {
    result := map[string]interface{}{"cpu_temperature": nil, "bme280_temperature": nil, "bme280_pressure": nil, "bme280_humidity": nil}

    if cpuData, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err != nil {
      return nil, fmt.Errorf("Failed to read CPU temperature: %s", err)
    } else {
      cpuTemp, err := strconv.ParseFloat(strings.TrimSpace(string(cpuData)), 64)
      if err != nil {
        return nil, fmt.Errorf("Failed to parse CPU temperature: %s", err)
      } else {
        result["cpu_temperature"] = cpuTemp / 1000.0
      }
    }

		bme280Data, err := s.BME280.Read()
		if err != nil {
			return nil, fmt.Errorf("Failed to read BME280 data: %s", err)
		} else {
		  result["bme280_temperature"] = bme280Data[0]
		  result["bme280_pressure"] = bme280Data[1] / 100
		  result["bme280_humidity"] = bme280Data[2]
		}

    return result, nil
}

func main() {
    // Parse arguments
    optPrefix := flag.String("metric-key-prefix", "sensor", "Metric key prefix")
    optTempfile := flag.String("tempfile", "", "Temp file name")
    flag.Parse()

    // Initialize BME280
    if err := embd.InitI2C(); err != nil {
  		panic(err)
  	}
  	defer embd.CloseI2C()

  	bus := embd.NewI2CBus(1)

  	opt := bme280.NewOpt()
  	bme280, err := bme280.New(bus, opt)
  	if err != nil {
  		panic(err)
  	}

    // Initialize helper
    s := SensorPlugin{
        Prefix: *optPrefix,
        BME280: bme280,
    }
    helper := mp.NewMackerelPlugin(s)
    helper.Tempfile = *optTempfile
    if helper.Tempfile == "" {
        helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-%s", *optPrefix)
    }
    helper.Run()
}
