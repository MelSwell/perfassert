package cmd

import (
	"flag"
	"log"
	"perfassert/pkg/benchmarks"
	"perfassert/pkg/thresholds"
)

var (
	benchPattern         *string
	benchmemOn           *bool
	benchTime            *string
	configFilepath       *string
	nsPerOpThreshold     *float64
	bytesPerOpThreshold  *int64
	allocsPerOpThreshold *int64
)

func Run() {
	parseFlags()
	benchFlags := getBenchFlags()
	// run the benchmarks
	results, output, err := benchmarks.ExecBenchmarks(*benchPattern, benchFlags)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(output)

	// handle configurations
	var cfgs *thresholds.ThresholdConfigs
	if *configFilepath != "" {
		// read the config file
		cfgs, err = thresholds.LoadFileConfigs(*configFilepath)
		if err != nil {
			log.Fatalf("error reading config file: %v", err)
		}
	}

	if len(cfgs.Benchmarks) == 0 {
		cfgs.GetBenchmarkConfigsFromResults(results)
	}

	// check if thresholds are provided via command line
	if scanForCmdThresholds() {
		err = cfgs.HandleCmdThresholds(*nsPerOpThreshold, *bytesPerOpThreshold, *allocsPerOpThreshold)
		if err != nil {
			log.Fatalf("error handling command line thresholds: %v", err)
		}
	}

	// assert the thresholds
	if err = thresholds.AssertThresholds(cfgs, results); err != nil {
		log.Fatal(err)
	}
	log.Println("Threshold checks passed. Exiting perfassert...")
}

func parseFlags() {
	// go benchmark flags
	benchPattern = flag.String("bench", "", "pattern to pass to 'go test -bench'")
	benchmemOn = flag.Bool("benchmem", false, "enable memory allocation statistics")
	benchTime = flag.String("benchtime", "", "run enough iterations of each benchmark to take t, default is 1s")
	// perfassert flags
	configFilepath = flag.String("config", "", "path to the config file")
	nsPerOpThreshold = flag.Float64("maxns", 0, "threshold for ns/op")
	bytesPerOpThreshold = flag.Int64("maxbytes", 0, "threshold for B/op")
	allocsPerOpThreshold = flag.Int64("maxallocs", 0, "threshold for allocs/op")
	flag.Parse()
}

func getBenchFlags() []string {
	var benchFlags []string
	if *benchPattern == "" {
		log.Fatal("you must provide a benchmark pattern to -bench (e.g.: -bench ., -bench BenchmarkDBInsert, -bench BenchmarkDB*)")
	}

	if *benchmemOn {
		benchFlags = append(benchFlags, "-benchmem")
	}

	if *benchTime != "" {
		benchFlags = append(benchFlags, "-benchtime", *benchTime)
	}
	return benchFlags
}

func scanForCmdThresholds() bool {
	if *nsPerOpThreshold != 0 || *bytesPerOpThreshold != 0 || *allocsPerOpThreshold != 0 {
		return true
	}
	return false
}
