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

func main() {
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
		p, err := parseLogFile(loc)
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

	export(plots, failed)
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

	return fmt.Sprintf("%2dh %2dm", hours, int(math.Round(minutes)))
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
