package benchmarks

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

type BenchmarkResult struct {
	Name        string
	Cores       int64
	TotalOps    int64
	NsPerOp     float64
	AllocsPerOp int64
	BytesPerOp  int64
}

func ExecBenchmarks(benchPattern string, benchFlags []string) ([]BenchmarkResult, string, error) {
	args := []string{"test", "-bench", benchPattern}
	if len(benchFlags) > 0 {
		args = append(args, benchFlags...)
	}
	cmd := exec.Command("go", args...)

	fmt.Print("Running benchmarks...\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("error running benchmarks: %v", err)
	}

	results, err := parseBenchmarkOutput(string(output))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing benchmark output: %v", err)
	}

	return results, string(output), nil
}

func parseBenchmarkOutput(output string) ([]BenchmarkResult, error) {
	var results []BenchmarkResult
	re := regexp.MustCompile(`(\w+)-(\d+)\s+(\d+)\s+(\d+\.\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		cores, _ := strconv.ParseInt(match[2], 10, 64)
		totalOps, _ := strconv.ParseInt(match[3], 10, 64)
		nsPerOp, _ := strconv.ParseFloat(match[4], 64)
		bytesPerOp, _ := strconv.ParseInt(match[5], 10, 64)
		allocsPerOp, _ := strconv.ParseInt(match[6], 10, 64)

		results = append(results, BenchmarkResult{
			Name:        match[1],
			Cores:       cores,
			TotalOps:    totalOps,
			NsPerOp:     nsPerOp,
			BytesPerOp:  bytesPerOp,
			AllocsPerOp: allocsPerOp,
		})
	}

	return results, nil
}
