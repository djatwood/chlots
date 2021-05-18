package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const tablePadding = 4

var outputFormat = flag.String("f", "default", "output format")
var outputAverages = flag.Bool("a", false, "display averages")

var cols = []string{"K", "RAM", "Threads", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Copy", "Total", "Start", "End"}

func export(plots []*plot, failed map[string]error) {
	flag.Parse()

	switch *outputFormat {
	case "default":
		defaultFormat(plots, failed)
	case "csv":
		csvFormat(plots)
	default:
		fmt.Fprintln(os.Stderr, "Unknown output extension")
	}
}

func printTable(name string, cols []string, values [][]string) error {
	widths := make([]int, len(cols))
	for _, row := range append(values, cols) {
		if len(row) > len(cols) {
			return fmt.Errorf("mismatched table and column length: expected length %d got %d", len(cols), len(row))
		}
		for i, value := range row {
			if len(value) > widths[i] {
				widths[i] = len(value)
			}
		}
	}
	tableWidth := -tablePadding
	for _, n := range widths {
		tableWidth += n + tablePadding
	}

	var format string
	for i := range cols {
		format += fmt.Sprintf("%%-%dv", widths[i]+tablePadding)
	}

	fmt.Println(name)
	fmt.Println(strings.Repeat("-", tableWidth))
	for i, c := range cols {
		fmt.Printf("%s%s", c, strings.Repeat(" ", tablePadding+widths[i]-len(c)))
	}
	fmt.Println()

	for _, row := range values {
		r := make([]interface{}, len(row))
		for i := range row {
			r[i] = row[i]
		}
		fmt.Printf(format+"\n", r...)
	}

	return nil
}

func defaultFormat(plots []*plot, failed map[string]error) {
	configAverages := make(map[string][]plot)
	parallelAverages := make(map[int][]plot)
	activePlots := []plot{}
	prevDate := [3]int{}
	table := [][]string{}
	for _, p := range plots {
		configKey := fmt.Sprintf("%d:%d:%d:%d", p.KSize, p.RAM, p.Threads, p.Stripe)
		configAverages[configKey] = append(configAverages[configKey], *p)

		parallel := len(activePlots) == 0
		for _, x := range activePlots {
			if (!p.StartTime.Before(x.StartTime) && !p.StartTime.After(x.EndTime)) ||
				(!p.EndTime.Before(x.StartTime) && !p.EndTime.After(x.EndTime)) ||
				(!x.StartTime.Before(p.StartTime) && !x.StartTime.After(p.EndTime)) ||
				(!x.EndTime.Before(p.StartTime) && !x.EndTime.After(p.EndTime)) {
				parallel = true
				break
			}
		}

		if parallel {
			activePlots = append(activePlots, *p)
		} else {
			parallelAverages[len(activePlots)] = append(parallelAverages[len(activePlots)], activePlots...)
			activePlots = []plot{*p}
		}

		year, month, day := p.EndTime.Date()
		if len(table) > 0 && (prevDate[0] != year || prevDate[1] != int(month) || prevDate[2] != day) {
			printTable(fmt.Sprintf("%s %d, %d", time.Month(prevDate[1]), prevDate[2], prevDate[0]), cols, table)
			fmt.Println()
			table = [][]string{}
		}

		table = append(table, []string{
			strconv.Itoa(p.KSize),
			strconv.Itoa(p.RAM),
			fmt.Sprintf("%d:%d", p.Threads, p.Stripe),
			humanTime(p.Phases[0]),
			humanTime(p.Phases[1]),
			humanTime(p.Phases[2]),
			humanTime(p.Phases[3]),
			humanTime(p.Phases[4]),
			humanTime(p.TotalTime),
			p.StartTime.Format("15:04"),
			p.EndTime.Format("15:04"),
		})

		prevDate = [3]int{year, int(month), day}
	}

	if len(activePlots) > 0 {
		parallelAverages[len(activePlots)] = append(parallelAverages[len(activePlots)], activePlots...)
	}

	if len(table) > 0 {
		printTable(fmt.Sprintf("%s %d, %d", time.Month(prevDate[1]), prevDate[2], prevDate[0]), cols, table)
		fmt.Println()
		table = [][]string{}
	}

	if *outputAverages {
		printConfigAverages(configAverages)
		fmt.Println()
		printParallelAverages(parallelAverages)
	}

	if len(failed) > 0 {
		fmt.Fprintln(os.Stderr, "\nFailed to parse the following plots")
		for loc, err := range failed {
			fmt.Fprintln(os.Stderr, loc, err)
		}
	}
}

func printConfigAverages(groups map[string][]plot) {
	table := [][]string{}
	for _, plots := range groups {
		values := make([][]float64, 6)
		avg := plot{
			KSize:   plots[0].KSize,
			RAM:     plots[0].RAM,
			Threads: plots[0].Threads,
			Stripe:  plots[0].Stripe,
		}
		for _, p := range plots {
			values[0] = append(values[0], p.Phases[0])
			values[1] = append(values[1], p.Phases[1])
			values[2] = append(values[2], p.Phases[2])
			values[3] = append(values[3], p.Phases[3])
			values[4] = append(values[4], p.Phases[4])
			values[5] = append(values[5], p.TotalTime)
		}

		table = append(table, []string{
			strconv.Itoa(avg.KSize),
			strconv.Itoa(avg.RAM),
			fmt.Sprintf("%d:%d", avg.Threads, avg.Stripe),
			humanTime(mean(values[0])),
			humanTime(mean(values[1])),
			humanTime(mean(values[2])),
			humanTime(mean(values[3])),
			humanTime(mean(values[4])),
			humanTime(mean(values[5])),
			strconv.Itoa(len(plots)),
		})
	}

	cols := append(cols[:9], "Plots")
	printTable("Config Averages", cols, table)
}

func printParallelAverages(groups map[int][]plot) {
	table := [][]string{}
	for c, plots := range groups {
		values := make([][]float64, 6)
		for _, p := range plots {
			values[0] = append(values[0], p.Phases[0])
			values[1] = append(values[1], p.Phases[1])
			values[2] = append(values[2], p.Phases[2])
			values[3] = append(values[3], p.Phases[3])
			values[4] = append(values[4], p.Phases[4])
			values[5] = append(values[5], p.TotalTime)
		}

		table = append(table, []string{
			humanTime(mean(values[0])),
			humanTime(mean(values[1])),
			humanTime(mean(values[2])),
			humanTime(mean(values[3])),
			humanTime(mean(values[4])),
			humanTime(mean(values[5])),
			strconv.Itoa(c),
			strconv.Itoa(len(plots)),
		})
	}

	cols := []string{"Phase 1", "Phase 2", "Phase 3", "Phase 4", "Copy", "Total", "Parallel", "Plots"}
	err := printTable("Parallel Averages", cols, table)
	if err != nil {
		panic(err)
	}
}

func csvFormat(plots []*plot) {
	cols := []string{"K", "RAM", "Threads", "Stripe", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Copy", "Total", "Start", "End", "Temp 1", "Temp 2", "Dest"}
	w := csv.NewWriter(os.Stdout)
	w.Write(cols)

	for _, p := range plots {
		record := []string{
			strconv.Itoa(p.KSize),
			strconv.Itoa(p.RAM),
			strconv.Itoa(p.Threads),
			strconv.Itoa(p.Stripe),
			strings.TrimSpace(humanTime(p.Phases[0])),
			strings.TrimSpace(humanTime(p.Phases[1])),
			strings.TrimSpace(humanTime(p.Phases[2])),
			strings.TrimSpace(humanTime(p.Phases[3])),
			strings.TrimSpace(humanTime(p.Phases[4])),
			strings.TrimSpace(humanTime(p.TotalTime)),
			p.StartTime.Format("2006-01-02 15:04:05"),
			p.EndTime.Format("2006-01-02 15:04:05"),
			p.TempDirs[0],
			p.TempDirs[1],
			p.DestDir,
		}
		err := w.Write(record)
		if err != nil {
			panic(err)
		}
	}

	w.Flush()
}
