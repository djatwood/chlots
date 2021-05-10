package main

import "fmt"

func print(plots []*plot, failed map[string]error) {
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
		fmt.Println("\nFailed to parse the following plots")
		for loc, err := range failed {
			fmt.Println(loc, err)
		}
	}
}
