package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flopp/go-parkrunparser"
	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] EVENTID [YEAR]
Show year statistics.

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload    bool
	eventId        string
	year           int
	guestCountries bool
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	guestCountries := flag.Bool("guestcountries", false, "determine guest countries (may take some time)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 2 {
		year, err := strconv.Atoi(flag.Args()[1])
		if err != nil {
			flag.Usage()
			os.Exit(1)
			return CommandLineOptions{}
		}
		return CommandLineOptions{
			*forceReload, flag.Args()[0], year, *guestCountries,
		}
	} else if len(flag.Args()) == 1 {
		return CommandLineOptions{
			*forceReload, flag.Args()[0], 0, *guestCountries,
		}
	} else {
		flag.Usage()
		os.Exit(1)
		return CommandLineOptions{}
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

func printLine(count int, names []string, exclusives, guests int) {
	sort.Strings(names)
	fmt.Printf("%3dx #=%d x=%d g=%d %s\n", count, len(names), exclusives, guests, strings.Join(names, ", "))
}

func printLineVol(count int, names []string) {
	sort.Strings(names)
	fmt.Printf("%3dx #=%d %s\n", count, len(names), strings.Join(names, ", "))
}

func printHistogram(items []CountId, names map[string]string, id_runs *map[string]int) {
	last_count := -1
	ns := make([]string, 0)
	exclusives := 0
	guests := 0
	for _, item := range items {
		if last_count != -1 && last_count != item.Count {
			if id_runs != nil {
				printLine(last_count, ns, exclusives, guests)
			} else {
				printLineVol(last_count, ns)
			}
			ns = make([]string, 0)
			exclusives = 0
			guests = 0
		}
		tag := ""
		if id_runs != nil {
			if r, ok := (*id_runs)[item.Id]; ok {
				if r == item.Count {
					tag = "/x"
					exclusives += 1
				} else if item.Count <= r/4 {
					tag = "/g"
					guests += 1
				}
			}
		}
		if n, ok := names[item.Id]; ok {
			ns = append(ns, n+tag)
		} else {
			ns = append(ns, item.Id+tag)
		}
		last_count = item.Count
	}
	if last_count != -1 {
		if id_runs != nil {
			printLine(last_count, ns, exclusives, guests)
		} else {
			printLineVol(last_count, ns)
		}
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
	ageGroups := make(map[string]int)
	runVol := make(map[string]int)

	names := make(map[string]string)
	id_runs := make(map[string]int)

	stat_participants := make([]int, 0)
	stat_volunteers := make([]int, 0)

	time_bins := make(map[int]int)
	var min_time time.Duration = 0
	var max_time time.Duration = 0
	var sum_time time.Duration = 0
	count_time := 0

	sex_female := 0
	sex_male := 0
	sex_unknown := 0

	people := make(map[string]*parkrun.Participant, 0)
	for _, run := range event.Runs {
		if options.year != 0 && run.Time.Year() != options.year {
			continue
		}
		runs += 1

		if err := run.Complete(); err != nil {
			panic(err)
		}

		r := 0
		v := 0
		stat_participants = append(stat_participants, len(run.Runners))
		stat_volunteers = append(stat_volunteers, len(run.Volunteers))

		rv := make(map[string]int)
		for _, p := range run.Runners {
			r += 1
			runners[p.Id] += 1
			names[p.Id] = p.Name
			id_runs[p.Id] = int(p.Runs)
			ageGroups[p.AgeGroup] += 1
			people[p.Id] = p
			rv[p.Id] += 1

			if p.Time != 0 {
				t := int(math.Floor(p.Time.Minutes()))
				time_bins[t] += 1
				if min_time == 0 || p.Time < min_time {
					min_time = p.Time
				}
				if max_time == 0 || p.Time > max_time {
					max_time = p.Time
				}
				sum_time += p.Time
				count_time += 1
			}

			if p.Sex == parkrunparser.SEX_FEMALE {
				sex_female += 1
			} else if p.Sex == parkrunparser.SEX_MALE {
				sex_male += 1
			} else {
				sex_unknown += 1
			}
		}
		for _, p := range run.Volunteers {
			v += 1
			volunteers[p.Id] += 1
			names[p.Id] = p.Name
			people[p.Id] = p
			rv[p.Id] += 1
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

		for id := range rv {
			runVol[id] += 1
		}
	}

	max_runs := -1
	max_runs_ids := make([]string, 0)
	count_runners := make([]CountId, 0, len(runners))
	for id, count := range runners {
		count_runners = append(count_runners, CountId{count, id})
		if id == "" {
			continue
		}
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

	count_runsvols := make([]CountId, 0, len(runners))
	for id, count := range runVol {
		count_runsvols = append(count_runsvols, CountId{count, id})
	}
	sort.Slice(count_runsvols, func(i, j int) bool {
		return count_runsvols[i].Count >= count_runsvols[j].Count
	})

	/*
		for id, _ := range people {
			r, rok := runners[id]
			v, vok := volunteers[id]
			if !rok {
				r = 0
			}
			if !vok {
				v = 0
			}
		}
	*/
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

	fmt.Println("\nParticipants Histogram (/g=guest, /x=exclusive)")
	printHistogram(count_runners, names, &id_runs)

	fmt.Println("\nEVENT;PARTICIPANTS;VOLUNTEERS")
	for i := 0; i < len(stat_participants); i += 1 {
		fmt.Printf("%d;%d;%d\n", i+1, stat_participants[i], stat_volunteers[i])
	}

	fmt.Println("\nSEXGROUP;COUNT")
	fmt.Printf("female;%d\n", sex_female)
	fmt.Printf("male;%d\n", sex_male)
	fmt.Printf("unknown;%d\n", sex_unknown)

	fmt.Println("\nAGEGROUP;COUNT")
	for ageGroup, count := range ageGroups {
		fmt.Printf("%s;%d\n", ageGroup, count)
	}

	events, err := parkrun.AllEvents()
	if err != nil {
		panic(err)
	}
	eventCountries := make(map[string]string)
	for _, event := range events {
		eventCountries[event.Id] = event.Country
	}
	countryCounts := make(map[string]int)
	guests := 0
	for _, count_id := range count_runners {
		total_count, ok := id_runs[count_id.Id]
		if !ok {
			panic(fmt.Errorf("no total runs for %s", count_id.Id))
		}

		if count_id.Count <= total_count/4 {
			if options.guestCountries {
				country, err := parkrun.GetParkrunnerCountry(count_id.Id, eventCountries)
				if err != nil {
					panic(err)
				}
				countryCounts[country] += count_id.Count
			}
			guests += count_id.Count
		}
	}
	fmt.Printf("\nGUESTS:\n%d\n", guests)
	if options.guestCountries {
		fmt.Println("\nGUEST COUNTRIES:")
		for country, count := range countryCounts {
			fmt.Printf("%s;%d\n", country, count)
		}
	}

	fmt.Println("\nRUN TIMES:")
	fmt.Printf("min=%v\n", min_time)
	fmt.Printf("max=%v\n", max_time)
	fmt.Printf("avg=%v\n", time.Duration(1000000000.0*sum_time.Seconds()/float64(count_time)))
	times := make([]int, 0, len(time_bins))
	for t, _ := range time_bins {
		times = append(times, t)
	}
	sort.Ints(times)

	fmt.Println("\nRUN TIME (MINUTES);COUNT")
	for _, t := range times {
		count := time_bins[t]
		fmt.Printf("%d;%d\n", t, count)
	}

	fmt.Println("\nVolunteers Histogram")
	printHistogram(count_vols, names, nil)

	fmt.Println("\nRun Or Vol")
	printHistogram(count_runsvols, names, nil)

	fmt.Println("\nNames Histogram")
	firstnames_count := make(map[string]int)
	for _, p := range people {
		name := p.Name
		a := strings.Split(name, " ")
		if len(a) < 2 {
			//fmt.Println(name)
		} else {
			firstnames_count[a[0]] += 1
		}
	}

	type str_int struct {
		s string
		i int
	}

	firstnames := make([]str_int, 0)
	for s, i := range firstnames_count {
		firstnames = append(firstnames, str_int{s, i})
	}
	sort.Slice(firstnames, func(i, j int) bool {
		return firstnames[i].i >= firstnames[j].i
	})
	fmt.Println("\nNAME;COUNT")
	for _, n := range firstnames {
		fmt.Printf("%s;%d\n", n.s, n.i)
	}
}
