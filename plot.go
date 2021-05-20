package main

import (
	"bufio"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

type plot struct {
	KSize     int
	Buffer    int
	Threads   int
	Stripe    int
	Phases    [5]float64
	TotalTime float64
	StartTime time.Time
	EndTime   time.Time
	TempDirs  [2]string
	DestDir   string
}

func parseLog(reader io.Reader, jobs []parseJob) (*plot, error) {
	p := new(plot)
	scanner := bufio.NewScanner(reader)
	for len(jobs) > 0 {
		if ok := scanner.Scan(); !ok {
			err := scanner.Err()
			if err == nil {
				return p, io.EOF
			}
			return p, err
		}

		j, line := jobs[0], scanner.Text()
		if j.matcher(line, j.match) {
			err := j.exec(p, line)
			if err != nil {
				return p, err
			}
			jobs = jobs[1:]
		}
	}

	return p, nil
}

func parseTempDirs(p *plot, line string) error {
	parsed := strings.Split(line[48:], " and ")
	if len(parsed) != 2 {
		return io.ErrUnexpectedEOF
	}
	p.TempDirs = [2]string{parsed[0], parsed[1]}
	return nil
}

func parseKSize(p *plot, line string) error {
	if len(line) < 14 {
		return io.ErrUnexpectedEOF
	}
	var err error
	p.KSize, err = strconv.Atoi(line[14:])
	return err
}

func parseBufferSize(p *plot, line string) error {
	if len(line) < 19 {
		return io.ErrUnexpectedEOF
	}
	var err error
	p.Buffer, err = strconv.Atoi(line[16 : len(line)-3])
	return err
}

func parseThreadCount(p *plot, line string) error {
	if len(line) < 6 {
		return io.ErrUnexpectedEOF
	}
	var err error
	var wordCount, wordStart int
	// Rather than adding a check after the loop, add a space to force a final check
	for i, r := range line + " " {
		if r != ' ' {
			continue
		}

		wordCount++
		switch wordCount {
		case 2:
			p.Threads, err = strconv.Atoi(line[wordStart:i])
			if err != nil {
				return err
			}
		case 7:
			p.Stripe, err = strconv.Atoi(line[wordStart:i])
			if err != nil {
				return err
			}
		}
		wordStart = i + 1
	}
	return nil
}

func parseStartTime(p *plot, line string) error {
	start := strings.LastIndex(line, "...") + 3
	if len(line) < start {
		return io.ErrUnexpectedEOF
	}
	var err error
	p.StartTime, err = time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
	return err
}

func parsePhaseTime(p *plot, line string) error {
	if len(line) < 19 {
		return io.ErrUnexpectedEOF
	}
	phase, err := strconv.Atoi(firstWord(line[15:]))
	if err != nil {
		return err
	}
	seconds, err := strconv.ParseFloat(firstWord(line[19:]), 64)
	if err != nil {
		return err
	}
	p.Phases[phase-1] = seconds
	return nil
}

func parseCopyTime(p *plot, line string) error {
	if len(line) < 12 {
		return io.ErrUnexpectedEOF
	}
	var err error
	p.Phases[4], err = strconv.ParseFloat(firstWord(line[12:]), 64)
	if err != nil {
		return err
	}
	// Parse End Time
	start := strings.LastIndex(line, ")") + 1
	if len(line) < start {
		return io.ErrUnexpectedEOF
	}
	p.EndTime, err = time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
	if err != nil {
		return err
	}
	p.TotalTime = p.EndTime.Sub(p.StartTime).Seconds()
	return nil
}

func parseDestDir(p *plot, line string) error {
	start, end := 25, -1
	for i := range line[start:] {
		if line[i] == '"' {
			end = i + start
			break
		}
	}
	if end == -1 {
		return io.ErrUnexpectedEOF
	}
	p.DestDir = path.Dir(line[start:end])
	return nil
}
