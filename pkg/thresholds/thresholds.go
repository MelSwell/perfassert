package thresholds

import (
	"encoding/json"
	"fmt"
	"os"
	"perfassert/pkg/benchmarks"
	"strings"

	"gopkg.in/yaml.v2"
)

type Threshold struct {
	MaxNsPerOp     float64 `yaml:"max_ns_per_op" json:"max_ns_per_op"`
	MaxBytesPerOp  int64   `yaml:"max_bytes_per_op" json:"max_bytes_per_op"`
	MaxAllocsPerOp int64   `yaml:"max_allocs_per_op" json:"max_allocs_per_op"`
}

type ThresholdConfigs struct {
	Thresholds map[string]Threshold `yaml:"thresholds" json:"thresholds"`
	Benchmarks map[string]string    `yaml:"benchmarks" json:"benchmarks"`
}

func LoadFileConfigs(filepath string) (*ThresholdConfigs, error) {
	var cfgs ThresholdConfigs
	data, err := os.ReadFile(filepath)
	if err != nil {
		return &cfgs, err
	}

	if strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml") {
		if err = yaml.Unmarshal(data, &cfgs); err != nil {
			return &cfgs, fmt.Errorf("error unmarshalling config file: %v", err)
		}
	}

	if strings.HasSuffix(filepath, ".json") {
		if err = json.Unmarshal(data, &cfgs); err != nil {
			return &cfgs, fmt.Errorf("error unmarshalling config file: %v", err)
		}
	}

	return &cfgs, err
}

func AssertThresholds(cfgs *ThresholdConfigs, br []benchmarks.BenchmarkResult) error {
	for _, r := range br {
		if cfgs.Benchmarks[r.Name] == "" {
			continue
		}

		t := cfgs.Thresholds[cfgs.Benchmarks[r.Name]]
		if r.NsPerOp > t.MaxNsPerOp {
			return fmt.Errorf("benchmark %s exceeded ns/op threshold: got %v, wanted %v", r.Name, r.NsPerOp, t.MaxNsPerOp)
		}
		if r.BytesPerOp > t.MaxBytesPerOp {
			return fmt.Errorf("benchmark %s exceeded bytes/op threshold: got %v, wanted %v", r.Name, r.BytesPerOp, t.MaxBytesPerOp)
		}
		if r.AllocsPerOp > t.MaxAllocsPerOp {
			return fmt.Errorf("benchmark %s exceeded allocs/op threshold: got %v, wanted %v", r.Name, r.AllocsPerOp, t.MaxAllocsPerOp)
		}
	}
	return nil
}

func (cfgs *ThresholdConfigs) HandleCmdThresholds(nsPerOp float64, bPerOp int64, aPerOp int64) error {
	var err error
	// if there are already thresholds from the config file, override them accordingly
	if len(cfgs.Thresholds) != 0 {
		if nsPerOp != 0 {
			err = cfgs.handleNsPerOpThreshold(nsPerOp)
		}
		if bPerOp != 0 {
			err = cfgs.handleBytesPerOpThreshold(bPerOp)
		}
		if aPerOp != 0 {
			err = cfgs.handleAllocsPerOpThreshold(aPerOp)
		}

		if err != nil {
			return err
		}
	} else {
		cfgs.Thresholds["global"] = Threshold{
			MaxNsPerOp:     nsPerOp,
			MaxBytesPerOp:  bPerOp,
			MaxAllocsPerOp: aPerOp,
		}
	}

	return nil
}

func (cfgs *ThresholdConfigs) handleNsPerOpThreshold(nsPerOp float64) error {
	if nsPerOp < 0 {
		return fmt.Errorf("maxns must be a positive number")
	}

	for key, t := range cfgs.Thresholds {
		t.MaxNsPerOp = nsPerOp
		cfgs.Thresholds[key] = t
	}

	return nil
}

func (cfgs *ThresholdConfigs) handleBytesPerOpThreshold(bPerOp int64) error {
	if bPerOp < 0 {
		return fmt.Errorf("maxbytes must be a positive number")
	}

	for key, t := range cfgs.Thresholds {
		t.MaxBytesPerOp = bPerOp
		cfgs.Thresholds[key] = t
	}

	return nil
}

func (cfgs *ThresholdConfigs) handleAllocsPerOpThreshold(aPerOp int64) error {
	if aPerOp < 0 {
		return fmt.Errorf("maxallocs must be a positive number")
	}

	for key, t := range cfgs.Thresholds {
		t.MaxAllocsPerOp = aPerOp
		cfgs.Thresholds[key] = t
	}

	return nil
}

func (cfgs *ThresholdConfigs) GetBenchmarkConfigsFromResults(br []benchmarks.BenchmarkResult) {
	for _, r := range br {
		cfgs.Benchmarks[r.Name] = "global"
	}
}
