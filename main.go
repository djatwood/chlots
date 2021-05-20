package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type options struct {
	format   string
	padding  int
	averages bool
}

type parseJob struct {
	match   string
	matcher func(line string, substr string) bool
	exec    func(p *plot, line string) error
}

var (
	outputFormat   = flag.String("f", "default", "output format")
	outputAverages = flag.Bool("a", false, "display averages")
	tablePadding   = flag.Int("p", 4, "table padding for default output")
	parseJobs      = []parseJob{
		{"Starting plotting", strings.HasPrefix, parseTempDirs},
		{"Plot size", strings.HasPrefix, parseKSize},
		{"Buffer size", strings.HasPrefix, parseBufferSize},
		{"threads of stripe", strings.Contains, parseThreadCount},
		{"Starting phase 1", strings.HasPrefix, parseStartTime},
		{"Time for phase 1", strings.HasPrefix, parsePhaseTime},
		{"Time for phase 2", strings.HasPrefix, parsePhaseTime},
		{"Time for phase 3", strings.HasPrefix, parsePhaseTime},
		{"Time for phase 4", strings.HasPrefix, parsePhaseTime},
		{"Copy time", strings.HasPrefix, parseCopyTime},
		{"Renamed final file", strings.HasPrefix, parseDestDir},
	}
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		args = append(args, filepath.Join(home, ".chia", "mainnet", "plotter"))
	}

	paths := getPaths(args...)
	plots := make([]*plot, 0, len(paths))
	failed := make(map[string]error)
	for _, loc := range paths {
		file, err := os.Open(loc)
		if err != nil {
			failed[loc] = err
		}
		p, err := parseLog(file, parseJobs)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			failed[loc] = err
			continue
		}
		plots = append(plots, p)
	}

	sort.Slice(plots, func(i, j int) bool {
		return plots[i].EndTime.Before(plots[j].EndTime)
	})

	export(plots, failed, options{
		format:   *outputFormat,
		averages: *outputAverages,
		padding:  *tablePadding,
	})
}

func getPaths(locations ...string) []string {
	var paths []string
	for _, loc := range locations {
		stats, err := os.Stat(loc)
		if err != nil {
			panic(err)
		}

		if !stats.IsDir() {
			paths = append(paths, loc)
		} else {
			dir, err := os.Open(loc)
			if err != nil {
				panic(err)
			}

			paths, err = dir.Readdirnames(-1)
			if err != nil {
				panic(err)
			}

			if !strings.HasSuffix(loc, string(os.PathSeparator)) {
				loc += string(os.PathSeparator)
			}

			for i := range paths {
				paths[i] = loc + paths[i]
			}
		}
	}
	return paths
}

func humanTime(seconds float64) string {
	minutes := seconds / 60
	hours := int(minutes / 60)
	minutes -= float64(hours) * 60
	raw := fmt.Sprintf("%dh %dm", hours, int(math.Round(minutes)))
	return fmt.Sprintf("%-7s", raw)
}

func firstWord(str string) string {
	for i, r := range str {
		if r == ' ' {
			return str[:i]
		}
	}
	return str
}

func maxInt(a, b int) int {
	if b > a {
		return b
	}
	return a
}

func mean(list []float64) (avg float64) {
	for i, t := 0, 1.0; i < len(list); i, t = i+1, t+1 {
		avg += (list[i] - avg) / t
	}
	return
}
