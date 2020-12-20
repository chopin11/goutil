package load

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	loadAvgProcFile = "/proc/loadavg"
)

type LoadStat struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

func (l LoadStat) String() string {
	s, _ := json.Marshal(l)
	return string(s)
}

func LoadAvg() (*LoadStat, error) {
	values, err := readLoadAvg()
	if err != nil {
		return nil, err
	}

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	ret := &LoadStat{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}

func readLoadAvg() ([]string, error) {
	line, err := ioutil.ReadFile(loadAvgProcFile)
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(line))

	return values, nil
}
