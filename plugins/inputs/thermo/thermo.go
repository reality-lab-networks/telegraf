package thermo

import (
	"bufio"
	"os"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Thermo struct {
	Files []string
}

func NewThermo() *Thermo {
	return &Thermo{
		Files: []string{
			"/sys/devices/virtual/thermal/thermal_zone1/temp",
			"/sys/devices/virtual/thermal/thermal_zone2/temp",
			"/sys/devices/virtual/thermal/thermal_zone3/temp",
			"/sys/devices/virtual/thermal/thermal_zone4/temp",
			"/sys/devices/virtual/thermal/thermal_zone5/temp",
			"/sys/devices/virtual/thermal/thermal_zone6/temp",
		},
	}
}

const sampleConfig = `
  ## specify files=[] if you want to override standard location
`

func (t *Thermo) SampleConfig() string {
	return sampleConfig
}

func (t *Thermo) Description() string {
	return sampleConfig
}

func (t *Thermo) ReadTemp(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan()

		val := scanner.Text()

		if err = scanner.Err(); err != nil {
			return -1, err
		} else {
			if i, err := strconv.Atoi(val); err == nil {
				return i, nil
			}
			return -1, err
		}
	}
	return -1, err
}

func (t *Thermo) Gather(acc telegraf.Accumulator) error {
	fpgaTemp, err := t.ReadTemp(t.Files[0])
	if err != nil {
		acc.AddError(err)
	}

	jetsonTemp := -1
	for _, filepath := range t.Files[1:] {
		val, err := t.ReadTemp(filepath)
		if err != nil {
			acc.AddError(err)
		} else if jetsonTemp < val {
			jetsonTemp = val
		}
	}

	acc.AddFields("temp", map[string]interface{}{"fpga": fpgaTemp, "jetson": jetsonTemp}, nil)

	return nil
}

func init() {
	inputs.Add("temp", func() telegraf.Input {
		return NewThermo()
	})
}
