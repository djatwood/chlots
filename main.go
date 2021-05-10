package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type plot struct {
	KSize     int
	RAM       int
	Threads   int
	Phases    [5]float64
	StartTime time.Time
	EndTime   time.Time

	scanner     *bufio.Scanner
	skipLines   int
	currentLine int
}

var errorMatches = []string{
	"Only wrote",
	"Could not copy",
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		args = append(args, filepath.Join(home, ".chia", "mainnet", "plotter"))
	}

	var paths []string
	for _, loc := range args {
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

	plots := []*plot{}
	failed := [][2]string{}
	for _, loc := range paths {
		p, err := parseLogFile(loc)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			failed = append(failed, [2]string{loc, err.Error()})
			continue
		}
		plots = append(plots, p)
	}

	sort.Slice(plots, func(i, j int) bool {
		return plots[i].EndTime.Before(plots[j].EndTime)
	})

	prevDate := [3]int{}
	for _, p := range plots {
		year, month, day := p.EndTime.Date()
		if prevDate[0] != year || prevDate[1] != int(month) || prevDate[2] != day {
			fmt.Printf("\n%s %d, %d\n", month, day, year)
			fmt.Println("KSize    RAM(MiB)    Threads    Phase 1    Phase 2    Phase 3    Phase 4    Copy    Total    Start End")
		}

		fmt.Println(p, p.StartTime.Format("15:04"), p.EndTime.Format("15:04"))

		prevDate = [3]int{year, int(month), day}
	}

	if len(failed) > 0 {
		fmt.Println("Failed to parse the following plots")
		for _, f := range failed {
			fmt.Println(f[0], f[1])
		}
	}
}

func parseLogFile(loc string) (*plot, error) {
	file, err := os.Open(loc)
	if err != nil {
		return nil, err
	}

	p := new(plot)
	p.scanner = bufio.NewScanner(file)

	// calculate variable length introduction
	p.skipLines = 3
	for {
		err = p.scanLine(p.currentLine + 1)
		if err != nil {
			return nil, err
		}
		if len(p.scanner.Text()) == 0 {
			p.skipLines = p.currentLine - 1
			break
		}
	}

	// Parse Plot Size
	p.KSize, err = p.parsePlotSize()
	if err != nil {
		return p, err
	}

	// Parse Buffer Size
	p.RAM, err = p.parseBufferSize()
	if err != nil {
		return p, err
	}

	// Parse Thread Count
	p.Threads, err = p.parseThreadCount()
	if err != nil {
		return p, err
	}

	// Parse Start Time
	p.StartTime, err = p.parseStartTime()
	if err != nil {
		return p, err
	}

	// Parse Time for Phases 1 - 4
	p.Phases, err = p.parsePhaseTimes()
	if err != nil {
		return p, err
	}

	// Parse Copy Time
	p.Phases[4], err = p.parseCopyTime()
	if err != nil {
		return p, err
	}

	// Parse End Time
	p.EndTime, err = p.parseEndTime()
	if err != nil {
		return p, err
	}

	return p, nil
}

func humanTime(seconds float64) string {
	minutes := seconds / 60
	hours := int(minutes / 60)
	minutes -= float64(hours) * 60

	return fmt.Sprintf("%dh %dm", hours, int(math.Round(minutes)))
}

func firstWord(str string) string {
	for i, r := range str {
		if r == ' ' {
			return str[:i]
		}
	}
	return str
}

func (p *plot) scanLine(n int) error {
	n = p.skipLines + (n - 3)
	for ; p.currentLine < n; p.currentLine++ {
		if ok := p.scanner.Scan(); !ok {
			err := p.scanner.Err()
			if err == nil {
				return io.EOF
			}
			return err
		}

		line := p.scanner.Text()
		for _, start := range errorMatches {
			if strings.HasPrefix(line, start) {
				p.skipLines++
				n++
			}
		}
	}
	return nil
}

func (p plot) String() string {
	return fmt.Sprintf("%-8d %-11d %-10d %-10s %-10s %-10s %-10s %-7s %-8s",
		p.KSize,
		p.RAM,
		p.Threads,
		humanTime(p.Phases[0]),
		humanTime(p.Phases[1]),
		humanTime(p.Phases[2]),
		humanTime(p.Phases[3]),
		humanTime(p.Phases[4]),
		humanTime(p.EndTime.Sub(p.StartTime).Seconds()),
	)
}

func (p *plot) skipCopyErrors() error {
	err := p.scanLine(2624)
	if err != nil {
		return err
	}
	for strings.HasPrefix(p.scanner.Text(), "Could not copy") {
		err = p.scanLine(p.currentLine + 1)
		if err != nil {
			return err
		}
		p.skipLines++
	}
	return nil
}

func (p *plot) parsePlotSize() (int, error) {
	err := p.scanLine(7)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	return strconv.Atoi(line[14:])
}

func (p *plot) parseBufferSize() (int, error) {
	err := p.scanLine(8)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	return strconv.Atoi(line[16 : len(line)-3])
}

func (p *plot) parseThreadCount() (int, error) {
	err := p.scanLine(10)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	return strconv.Atoi(firstWord(line[6:]))
}

func (p *plot) parseStartTime() (time.Time, error) {
	err := p.scanLine(12)
	if err != nil {
		return time.Time{}, err
	}
	line := p.scanner.Text()
	start := strings.LastIndex(line, "...") + 3
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}

func (p *plot) parsePhaseTimes() (phases [5]float64, err error) {
	for i, n := range [4]int{801, 834, 2474, 2620} {
		err = p.scanLine(n)
		if err != nil {
			return
		}
		line := p.scanner.Text()
		phases[i], err = strconv.ParseFloat(firstWord(line[19:]), 64)
		if err != nil {
			return
		}
	}

	return
}

func (p *plot) parseCopyTime() (float64, error) {
	err := p.scanLine(2625)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	return strconv.ParseFloat(firstWord(line[12:]), 64)
}

func (p *plot) parseEndTime() (time.Time, error) {
	err := p.scanLine(2625)
	if err != nil {
		return time.Time{}, err
	}
	line := p.scanner.Text()
	start := strings.LastIndex(line, ")") + 1
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}
