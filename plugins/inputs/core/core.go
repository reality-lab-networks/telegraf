package core

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Core struct {
	Thermal []string
	Emc     string
	Avp     string
	NvDec   string
	MsEnc   string
	Gpu     string
	SdCard  string
}

func NewCore() *Core {
	return &Core{
		Thermal: []string{
			"/dev/ql2",
			"/sys/devices/virtual/thermal/thermal_zone0/temp",
			"/sys/devices/virtual/thermal/thermal_zone1/temp",
			"/sys/devices/virtual/thermal/thermal_zone2/temp",
			"/sys/devices/virtual/thermal/thermal_zone3/temp",
			"/sys/devices/virtual/thermal/thermal_zone4/temp",
			"/sys/devices/virtual/thermal/thermal_zone5/temp",
			"/sys/devices/virtual/thermal/thermal_zone6/temp",
			//"/sys/devices/virtual/thermal/thermal_zone7/temp",
			"/sys/devices/virtual/thermal/thermal_zone8/temp",
		},
		Emc:    "/sys/kernel/debug/clock/emc/rate",
		Avp:    "/sys/kernel/debug/clock/avp.sclk/rate",
		NvDec:  "/sys/kernel/debug/clock/nvdec/rate",
		MsEnc:  "/sys/kernel/debug/clock/msenc/rate",
		Gpu:    "/sys/devices/platform/host1x/gpu.0/load",
		SdCard: "/run/user/1001/sd_stats",
	}
}

const sampleConfig = `no config needed`

func (t *Core) SampleConfig() string {
	return sampleConfig
}

func (t *Core) Description() string {
	return sampleConfig
}

func (t *Core) Read(filepath string) (int, error) {
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

func (t *Core) ReadSys(filepath string) (int, error) {
	var fd, numread int
	var err error

	fd, err = syscall.Open(filepath, syscall.O_RDONLY, 0777)

	if err == nil {
		defer syscall.Close(fd)

		buffer := make([]byte, 10, 100)

		numread, err = syscall.Read(fd, buffer)

		if err != nil {
			fmt.Print(err.Error(), "\n")
		}

		fmt.Printf("Numbytes read: %d\n", numread)
		fmt.Printf("Buffer: %b\n", buffer)
	}
	return -1, err
}

func (t *Core) ReadSdCardInfo(filepath string) (string, int, int, error) {
	file, err := os.Open(filepath)
	if err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan()

		line := scanner.Text()

		if err = scanner.Err(); err != nil {
			return "", -1, -1, err
		} else {
			values := strings.Split(line, " ")

			if len(values) < 3 {
				return "", -1, -1, errors.New("not enough data to read")
			}

			used := -1
			total := -1

			names := make([]string, len(values)-2)
			copy(names, values)

			name := strings.Join(names, " ")

			if i, err := strconv.Atoi(values[len(values)-2]); err == nil {
				used = i
			}

			if err != nil {
				return "", -1, -1, err
			}

			if i, err := strconv.Atoi(values[len(values)-1]); err == nil {
				total = i
			}

			if err != nil {
				return "", -1, -1, err
			}

			return name, total, used, nil
		}
	}
	return "", -1, -1, err
}

func (t *Core) Gather(acc telegraf.Accumulator) error {
	// fpgaTemperature, err := t.Read(t.Thermal[0])
	// if err != nil {
	// 	acc.AddError(err)
	// }

	thermo := -1
	for _, filepath := range t.Thermal[1:] {
		val, err := t.Read(filepath)
		if err != nil {
			acc.AddError(err)
		} else if thermo < val {
			thermo = val
		}
	}

	// emc, err := t.Read(t.Emc)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	// avp, err := t.Read(t.Avp)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	// nvdec, err := t.Read(t.NvDec)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	// msenc, err := t.Read(t.MsEnc)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	// gpu, err := t.Read(t.Gpu)
	// if err != nil {
	// 	acc.AddError(err)
	// }

	name, total, used, err := t.ReadSdCardInfo(t.SdCard)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields(
		"core",
		map[string]interface{}{
			// "fpga_temperature": byte(fpgaTemperature),
			"thermo": thermo,
			// "emc":           emc,
			// "avp":           avp,
			// "nvdec":         nvdec,
			// "msenc":         msenc,
			// "gpu":           gpu,
			"sd_card_name":  name,
			"sd_card_used":  used,
			"sd_card_total": total,
		},
		nil)

	return nil
}

func init() {
	inputs.Add("core", func() telegraf.Input {
		return NewCore()
	})
}
