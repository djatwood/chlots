package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var outputFormat = flag.String("f", "default", "output format")

var cols = []string{"K", "RAM", "Threads", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Copy", "Total", "Start", "End"}
var colWidths = []int{2, 4, 7, 7, 7, 7, 7, 7, 7, 5, 5}

func export(plots []*plot, failed map[string]error) {
	flag.Parse()

	switch *outputFormat {
	case "default":
		defaultOutput(plots, failed)
	case "csv":
		csvFormat(plots)
	default:
		fmt.Fprintln(os.Stderr, "Unknown output extension")
	}
}

func defaultOutput(plots []*plot, failed map[string]error) {
	var padding = 4
	var format string
	for i := range cols {
		format += fmt.Sprintf(" %%-%dv", colWidths[i]+padding-1)
	}
	format = format[1:]

	prevDate := [3]int{}
	for _, p := range plots {
		year, month, day := p.EndTime.Date()
		if prevDate[0] != year || prevDate[1] != int(month) || prevDate[2] != day {
			fmt.Printf("\n%s %d, %d\n", month, day, year)
			for i, c := range cols {
				fmt.Printf("%s%s", c, strings.Repeat(" ", padding+colWidths[i]-len(c)))
			}
			fmt.Println()
		}

		fmt.Printf(format+"\n", p.KSize,
			p.RAM,
			fmt.Sprintf("%d:%d", p.Threads, p.Stripe),
			humanTime(p.Phases[0]),
			humanTime(p.Phases[1]),
			humanTime(p.Phases[2]),
			humanTime(p.Phases[3]),
			humanTime(p.Phases[4]),
			humanTime(p.EndTime.Sub(p.StartTime).Seconds()),
			p.StartTime.Format("15:04"),
			p.EndTime.Format("15:04"),
		)

		prevDate = [3]int{year, int(month), day}
	}

	if len(failed) > 0 {
		fmt.Fprintln(os.Stderr, "\nFailed to parse the following plots")
		for loc, err := range failed {
			fmt.Fprintln(os.Stderr, loc, err)
		}
	}
}

func csvFormat(plots []*plot) {
	w := csv.NewWriter(os.Stdout)
	w.Write(cols)

	for _, p := range plots {
		record := []string{
			strconv.Itoa(p.KSize),
			strconv.Itoa(p.RAM),
			fmt.Sprintf("%d:%d", p.Threads, p.Stripe),
			humanTime(p.Phases[0]),
			humanTime(p.Phases[1]),
			humanTime(p.Phases[2]),
			humanTime(p.Phases[3]),
			humanTime(p.Phases[4]),
			humanTime(p.EndTime.Sub(p.StartTime).Seconds()),
			p.StartTime.Format("2006-01-02 15:04:05"),
			p.EndTime.Format("2006-01-02 15:04:05"),
		}
		err := w.Write(record)
		if err != nil {
			panic(err)
		}
	}

	w.Flush()
}
