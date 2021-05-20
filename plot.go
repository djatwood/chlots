package main

import (
	"bufio"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type plot struct {
	KSize     int
	RAM       int
	Threads   int
	Stripe    int
	Phases    [5]float64
	TotalTime float64
	StartTime time.Time
	EndTime   time.Time
	TempDirs  [2]string
	DestDir   string
}

var lineMatches = []string{
	"^Starting plotting",
	"^Plot size",
	"^Buffer size",
	"*threads of stripe",
	"^Starting phase 1",
	"^Time for phase 1",
	"^Time for phase 2",
	"^Time for phase 3",
	"^Time for phase 4",
	"^Copy time",
	"^Renamed final file",
}

func parseLogFile(loc string) (*plot, error) {
	file, err := os.Open(loc)
	if err != nil {
		return nil, err
	}

	p := new(plot)
	scanner := bufio.NewScanner(file)

	matches := lineMatches[:]
	matcher := getMatcher(matches[0])
	lines := make([]string, 0, len(matches))
	for {
		if ok := scanner.Scan(); !ok {
			err := scanner.Err()
			if err == nil {
				break
			}
			return p, err
		}

		line := scanner.Text()
		if matcher(line, matches[0][1:]) {
			lines = append(lines, line)
			matches = matches[1:]
			if len(matches) < 1 {
				break
			}
			matcher = getMatcher(matches[0])
		}
	}

	if len(lineMatches) != len(lines) {
		return nil, io.EOF
	}

	// Parse temp dirs
	p.TempDirs, err = parseTempDirs(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Plot Size
	lines = lines[1:]
	p.KSize, err = parsePlotSize(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Buffer Size
	lines = lines[1:]
	p.RAM, err = parseBufferSize(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Thread Count
	lines = lines[1:]
	p.Threads, p.Stripe, err = parseThreadCount(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Start Time
	lines = lines[1:]
	p.StartTime, err = parseStartTime(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Time for Phase 1
	lines = lines[1:]
	p.Phases[0], err = parsePhaseTime(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Time for Phase 2
	lines = lines[1:]
	p.Phases[1], err = parsePhaseTime(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Time for Phase 3
	lines = lines[1:]
	p.Phases[2], err = parsePhaseTime(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Time for Phase 4
	lines = lines[1:]
	p.Phases[3], err = parsePhaseTime(lines[0])
	if err != nil {
		return p, err
	}

	// Parse Copy Time & End Time
	lines = lines[1:]
	p.Phases[4], err = parseCopyTime(lines[0])
	if err != nil {
		return p, err
	}
	p.EndTime, err = parseEndTime(lines[0])
	if err != nil {
		return p, err
	}

	p.TotalTime = p.EndTime.Sub(p.StartTime).Seconds()

	// Parse Dest Dir
	lines = lines[1:]
	p.DestDir, err = parseDestDir(lines[0])
	if err != nil {
		return p, err
	}

	return p, nil
}

func getMatcher(match string) func(string, string) bool {
	switch match[0] {
	case '^':
		return strings.HasPrefix
	case '$':
		return strings.HasSuffix
	default:
		return strings.Contains
	}
}

func parseTempDirs(line string) ([2]string, error) {
	parsed := strings.Split(line[48:], "and")
	if len(parsed) != 2 {
		return [2]string{}, io.ErrUnexpectedEOF
	}
	return [2]string{
		parsed[0],
		parsed[1],
	}, nil
}

func parsePlotSize(line string) (int, error) {
	if len(line) < 14 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.Atoi(line[14:])
}

func parseBufferSize(line string) (int, error) {
	if len(line) < 19 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.Atoi(line[16 : len(line)-3])
}

func parseThreadCount(line string) (threads int, stripe int, err error) {
	if len(line) < 6 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	var wordCount, wordStart int
	// Rather than adding a check after the loop, add a space to force a final check
	for i, r := range line + " " {
		if r != ' ' {
			continue
		}

		wordCount++
		switch wordCount {
		case 2:
			threads, err = strconv.Atoi(line[wordStart:i])
			if err != nil {
				return
			}
		case 7:
			stripe, err = strconv.Atoi(line[wordStart:i])
			if err != nil {
				return
			}
		}
		wordStart = i + 1
	}
	return
}

func parseStartTime(line string) (time.Time, error) {
	start := strings.LastIndex(line, "...") + 3
	if len(line) < start {
		return time.Time{}, io.ErrUnexpectedEOF
	}
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}

func parsePhaseTime(line string) (float64, error) {
	if len(line) < 19 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.ParseFloat(firstWord(line[19:]), 64)
}

func parseCopyTime(line string) (float64, error) {
	if len(line) < 12 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.ParseFloat(firstWord(line[12:]), 64)
}

func parseEndTime(line string) (time.Time, error) {
	start := strings.LastIndex(line, ")") + 1
	if len(line) < start {
		return time.Time{}, io.ErrUnexpectedEOF
	}
	return time.Parse("Mon Jan  2 15:04:05 2006", strings.TrimSpace(line[start:]))
}

func parseDestDir(line string) (string, error) {
	start, end := 25, -1
	for i := range line[start:] {
		if line[i] == '"' {
			end = i + start
			break
		}
	}
	if end == -1 {
		return "", io.ErrUnexpectedEOF
	}
	return path.Dir(line[start:end]), nil
}
