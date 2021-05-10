package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

func (p *plot) parsePlotSize() (int, error) {
	err := p.scanLine(7)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	if len(line) < 14 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.Atoi(line[14:])
}

func (p *plot) parseBufferSize() (int, error) {
	err := p.scanLine(8)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	if len(line) < 19 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.Atoi(line[16 : len(line)-3])
}

func (p *plot) parseThreadCount() (int, error) {
	err := p.scanLine(10)
	if err != nil {
		return 0, err
	}
	line := p.scanner.Text()
	if len(line) < 6 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.Atoi(firstWord(line[6:]))
}

func (p *plot) parseStartTime() (time.Time, error) {
	err := p.scanLine(12)
	if err != nil {
		return time.Time{}, err
	}
	line := p.scanner.Text()
	start := strings.LastIndex(line, "...") + 3
	if len(line) < start {
		return time.Time{}, err
	}
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}

func (p *plot) parsePhaseTimes() (phases [5]float64, err error) {
	for i, n := range [4]int{801, 834, 2474, 2620} {
		err = p.scanLine(n)
		if err != nil {
			return
		}
		line := p.scanner.Text()
		if len(line) < 19 {
			err = io.ErrUnexpectedEOF
			return
		}
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
	if len(line) < 12 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.ParseFloat(firstWord(line[12:]), 64)
}

func (p *plot) parseEndTime() (time.Time, error) {
	err := p.scanLine(2625)
	if err != nil {
		return time.Time{}, err
	}
	line := p.scanner.Text()
	start := strings.LastIndex(line, ")") + 1
	if len(line) < start {
		fmt.Println("LENGTH", line)
		return time.Time{}, io.ErrUnexpectedEOF
	}
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}
