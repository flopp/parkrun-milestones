package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] EVENTID YEAR
Show year statistics.

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload bool
	eventId     string
	year        int
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		panic("bad command line")
	}

	year, err := strconv.Atoi(flag.Args()[1])
	if err != nil {
		panic(err)
	}
	return CommandLineOptions{
		*forceReload, flag.Args()[0], year,
	}
}

func getEvent(eventId string) *parkrun.Event {
	event, err := parkrun.LookupEvent(eventId)
	if err != nil {
		panic(err)
	}
	return event
}

type CountId struct {
	Count int
	Id    string
}

func printLine(count int, names []string) {
	sort.Strings(names)
	fmt.Printf("%d => %s\n", count, strings.Join(names, ", "))
}

func printHistogram(items []CountId, names map[string]string) {
	last_count := -1
	ns := make([]string, 0)
	for _, item := range items {
		if last_count != -1 && last_count != item.Count {
			printLine(last_count, ns)
			ns = make([]string, 0)
		}
		if n, ok := names[item.Id]; ok {
			ns = append(ns, n)
		} else {
			ns = append(ns, item.Id)
		}
		last_count = item.Count
	}
	if last_count != -1 {
		printLine(last_count, ns)
	}
}

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	event := getEvent(options.eventId)
	if err := event.Complete(); err != nil {
		panic(err)
	}

	runs := 0
	min_volunteers := -1
	max_volunteers := -1
	sum_volunteers := 0
	volunteers := make(map[string]int)
	min_runners := -1
	max_runners := -1
	sum_runners := 0
	runners := make(map[string]int)

	names := make(map[string]string)

	for _, run := range event.Runs {
		if run.Time.Year() != options.year {
			continue
		}
		runs += 1

		if err := run.Complete(); err != nil {
			panic(err)
		}

		r := 0
		v := 0

		for _, p := range run.Runners {
			r += 1
			runners[p.Id] += 1
			names[p.Id] = p.Name
		}
		for _, p := range run.Volunteers {
			v += 1
			volunteers[p.Id] += 1
			names[p.Id] = p.Name
		}

		if min_runners < 0 || r < min_runners {
			min_runners = r
		}
		if max_runners < 0 || r > max_runners {
			max_runners = r
		}
		sum_runners += r

		if min_volunteers < 0 || v < min_volunteers {
			min_volunteers = v
		}
		if max_volunteers < 0 || v > max_volunteers {
			max_volunteers = v
		}
		sum_volunteers += v
	}

	max_runs := -1
	max_runs_ids := make([]string, 0)
	count_runners := make([]CountId, 0, len(runners))
	for id, count := range runners {
		if id == "" {
			continue
		}
		count_runners = append(count_runners, CountId{count, id})
		if max_runs < 0 || count > max_runs {
			max_runs = count
			max_runs_ids = make([]string, 0)
			max_runs_ids = append(max_runs_ids, id)
		} else if count >= max_runs {
			max_runs_ids = append(max_runs_ids, id)
		}
	}
	sort.Slice(count_runners, func(i, j int) bool {
		return count_runners[i].Count >= count_runners[j].Count
	})
	rn := make([]string, 0)
	for _, id := range max_runs_ids {
		n, ok := names[id]
		if ok {
			rn = append(rn, n)
		} else {
			rn = append(rn, id)
		}
	}
	max_runs_names := strings.Join(rn, ", ")

	max_vols := -1
	max_vols_ids := make([]string, 0)
	count_vols := make([]CountId, 0, len(runners))
	for id, count := range volunteers {
		count_vols = append(count_vols, CountId{count, id})
		if max_vols < 0 || count > max_vols {
			max_vols = count
			max_vols_ids = make([]string, 0)
			max_vols_ids = append(max_vols_ids, id)
		} else if count >= max_vols {
			max_vols_ids = append(max_vols_ids, id)
		}
	}
	sort.Slice(count_vols, func(i, j int) bool {
		return count_vols[i].Count >= count_vols[j].Count
	})
	vn := make([]string, 0)
	for _, id := range max_vols_ids {
		n, ok := names[id]
		if ok {
			vn = append(vn, n)
		} else {
			vn = append(vn, id)
		}
	}
	max_vols_names := strings.Join(vn, ", ")

	fmt.Printf("Statistics for %s in %d\n", event.Name, options.year)
	fmt.Printf("Runs: %d\n", runs)
	fmt.Printf("Participants:\n")
	fmt.Printf("    Total: %d\n", sum_runners)
	fmt.Printf("    Unique: %d\n", len(runners))
	fmt.Printf("    Max / ø / Min (per event): %d / %.1f / %d\n", max_runners, float64(sum_runners)/float64(runs), min_runners)
	fmt.Printf("    Max participartions: %d (%s)\n", max_runs, max_runs_names)
	fmt.Printf("Volunteers:\n")
	fmt.Printf("    Total: %d\n", sum_volunteers)
	fmt.Printf("    Unique: %d\n", len(volunteers))
	fmt.Printf("    Max / ø / Min (per event): %d / %.1f / %d\n", max_volunteers, float64(sum_volunteers)/float64(runs), min_volunteers)
	fmt.Printf("    Max volunteerings: %d (%s)\n", max_vols, max_vols_names)

	fmt.Println("\nParticipants Histogram")
	printHistogram(count_runners, names)
	fmt.Println("\nVolunteers Histogram")
	printHistogram(count_vols, names)
}
