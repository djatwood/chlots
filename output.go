package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var cols = []string{"K", "RAM", "Threads", "Phase 1", "Phase 2", "Phase 3", "Phase 4", "Copy", "Total", "Start", "End"}

func export(plots []*plot, failed map[string]error, out options) {
	switch out.format {
	case "default":
		defaultFormat(plots, failed, out)
	case "csv":
		csvFormat(plots)
	default:
		fmt.Fprintln(os.Stderr, "Unknown output extension")
	}
}

func printTable(name [2]string, cols []string, values [][]string, padding int) error {
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
	tableWidth := -padding
	for _, n := range widths {
		tableWidth += n + padding
	}

	var format string
	for i := range cols {
		format += fmt.Sprintf("%%-%dv", widths[i]+padding)
	}

	fmt.Printf("%s%s%s\n", name[0], strings.Repeat(" ", tableWidth-len(name[0])-len(name[1])), name[1])
	fmt.Println(strings.Repeat("─", tableWidth))
	for i, c := range cols {
		fmt.Printf("%s%s", c, strings.Repeat(" ", padding+widths[i]-len(c)))
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

func defaultFormat(plots []*plot, failed map[string]error, out options) {
	configAverages := make(map[string][]plot)
	parallelAverages := make(map[int][]plot)
	pStart, pEnd := 0, 0
	prevDate := [3]int{}
	table := [][]string{}
	for i, p := range plots {
		configKey := fmt.Sprintf("%d:%d:%d:%d", p.KSize, p.Buffer, p.Threads, p.Stripe)
		configAverages[configKey] = append(configAverages[configKey], *p)

		for pStart > 0 && !plots[pStart].EndTime.Before(p.StartTime) {
			pStart--
		}
		for pStart < len(plots) && plots[pStart].EndTime.Before(p.StartTime) {
			pStart++
		}
		for pEnd < len(plots) && !plots[pEnd].StartTime.After(p.EndTime) {
			pEnd++
		}
		for pEnd < len(plots) && plots[pEnd].StartTime.After(p.EndTime) {
			pEnd--
		}
		parallel := pEnd - pStart + 1
		parallelAverages[parallel] = append(parallelAverages[parallel], *p)

		year, month, day := p.EndTime.Date()
		if len(table) > 0 && (prevDate[0] != year || prevDate[1] != int(month) || prevDate[2] != day) {
			durs := make([]float64, len(table))
			for j, x := range plots[i-len(table) : i] {
				durs[j] = x.TotalTime
			}
			name := defaultTableName(prevDate, len(table), mean(durs))
			printTable(name, cols, table, out.padding)
			fmt.Println()
			table = [][]string{}
		}

		table = append(table, []string{
			strconv.Itoa(p.KSize),
			strconv.Itoa(p.Buffer),
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

	if len(table) > 0 {
		durs := make([]float64, len(table))
		for j, x := range plots[len(plots)-len(table):] {
			durs[j] = x.TotalTime
		}
		name := defaultTableName(prevDate, len(table), mean(durs))
		printTable(name, cols, table, out.padding)
		fmt.Println()
		table = [][]string{}
	}

	if out.averages {
		printConfigAverages(configAverages, out.padding)
		fmt.Println()
		printParallelAverages(parallelAverages, out.padding)
	}

	if len(failed) > 0 {
		fmt.Fprintln(os.Stderr, "\nFailed to parse the following plots")
		for loc, err := range failed {
			fmt.Fprintln(os.Stderr, loc, err)
		}
	}
}

func defaultTableName(date [3]int, amount int, duration float64) [2]string {
	return [2]string{
		fmt.Sprintf("%s %2d, %d; %d plots", time.Month(date[1]), date[2], date[0], amount),
		"Average duration: " + humanTime(duration),
	}
}

func printConfigAverages(groups map[string][]plot, padding int) {
	table := [][]string{}
	for _, plots := range groups {
		values := make([][]float64, 6)
		avg := plot{
			KSize:   plots[0].KSize,
			Buffer:  plots[0].Buffer,
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
			strconv.Itoa(avg.Buffer),
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
	printTable([2]string{"Config Averages"}, cols, table, padding)
}

func printParallelAverages(groups map[int][]plot, padding int) {
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
	err := printTable([2]string{"Parallel Averages"}, cols, table, padding)
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
			strconv.Itoa(p.Buffer),
			strconv.Itoa(p.Threads),
			strconv.Itoa(p.Stripe),
			strconv.Itoa(int(math.Round(p.Phases[0]))),
			strconv.Itoa(int(math.Round(p.Phases[1]))),
			strconv.Itoa(int(math.Round(p.Phases[2]))),
			strconv.Itoa(int(math.Round(p.Phases[3]))),
			strconv.Itoa(int(math.Round(p.Phases[4]))),
			strconv.Itoa(int(math.Round(p.TotalTime))),
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
