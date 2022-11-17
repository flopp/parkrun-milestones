package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] [EVENTID...]
Determine the milestone candidates of the specified event(s) or
of all events of a county (if -country NAME is given). 

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload bool
	fancy       bool
	country     string
	eventIds    []string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	fancy := flag.Bool("fancy", false, "fancy formatting using emoji")
	country := flag.String("country", "", "select all events of the specified country")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *country == "" && len(flag.Args()) == 0 {
		panic("You have to specify either one or more EVENTID... or -country NAME")
	}
	if *country != "" && len(flag.Args()) != 0 {
		panic("You must not specify both one or more EVENTID... and -country NAME")
	}

	return CommandLineOptions{
		*forceReload, *fancy, *country, flag.Args(),
	}
}

func getEvents(eventIds []string, country string) []*parkrun.Event {
	events := make([]*parkrun.Event, 0)
	for _, eventId := range eventIds {
		event, err := parkrun.LookupEvent(eventId)
		if err != nil {
			panic(err)
		}
		events = append(events, event)
	}
	if country != "" {
		eventList, err := parkrun.AllEvents()
		if err != nil {
			panic(err)
		}
		lowerCountry := strings.TrimSpace(strings.ToLower(country))
		for _, event := range eventList {
			if strings.ToLower(event.Country) == lowerCountry {
				events = append(events, event)
			}
		}
	}
	return events
}

func pi(n int, icon string, text string) {
	sep := "\u2003"
	if text != "" {
		fmt.Printf("%s%s%s: %d\n", icon, sep, text, n)
	} else {
		fmt.Printf("%s%s%d\n", icon, sep, n)
	}
}

func ps(s string, icon string, text string) {
	sep := "\u2003"
	if text != "" {
		fmt.Printf("%s%s%s: %s\n", icon, sep, text, s)
	} else {
		fmt.Printf("%s%s%s\n", icon, sep, s)
	}
}

func pi2(n int, icon string, text string) {
	if n == 0 {
		return
	}
	indent := "\u2003\u2003"
	sep := "\u2003"
	fmt.Printf("%s%s%s%s: %d\n", indent, icon, sep, text, n)
}

func printFancy(event *parkrun.Event, run *parkrun.Run, r500, r250, r100, r50, r25, r1, pb, firstEvent, v500, v250, v100, v50, v25, v1 int) {
	fmt.Printf("%s\n", event.Name)
	pi(int(run.Index), "#️⃣", "")
	ps(run.Time.Format("2006-01-02"), "📅", "")
	pi(len(run.Runners), "🏃", "Runners")
	pi2(pb, "⏱", "new PB")
	pi2(firstEvent, "🧳", "first visitors")
	pi2(r1, "⭐️", "new parkrunners")
	pi2(r25, "🏆", "25. run anniversary")
	pi2(r50, "🏆", "50. run anniversary")
	pi2(r100, "🏆", "100. run anniversary")
	pi2(r250, "🏆", "250. run anniversary")
	pi2(r500, "🏆", "500. run anniversary")
	pi(len(run.Volunteers), "🦺", "Volunteers")
	pi2(v1, "⭐️", "new volunteers")
	pi2(v25, "🏆", "25. vol. anniversary")
	pi2(v50, "🏆", "50. vol. anniversary")
	pi2(v100, "🏆", "100. vol. anniversary")
	pi2(v250, "🏆", "250. vol. anniversary")
	pi2(v500, "🏆", "500. vol. anniversary")
	ps(fmt.Sprintf("https://%s/%s/results/%d/", event.CountryUrl, event.Id, run.Index), "👀", "")
}

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	events := getEvents(options.eventIds, options.country)
	for _, event := range events {
		if err := event.Complete(); err != nil {
			panic(err)
		}
		if len(event.Runs) == 0 {
			fmt.Printf("No runs at %s\n", event.Name)
			continue
		}

		run := event.Runs[len(event.Runs)-1]
		if err := run.Complete(); err != nil {
			panic(err)
		}

		firstEvent := 0
		pb := 0
		r1 := 0
		r25 := 0
		r50 := 0
		r100 := 0
		r250 := 0
		r500 := 0
		for _, participant := range run.Runners {
			if participant.Achievement == parkrun.AchievementFirst {
				if participant.Runs == 1 {
					r1 += 1
				} else {
					firstEvent += 1
				}
			} else if participant.Achievement == parkrun.AchievementPB {
				pb += 1
			}
			if participant.Runs == 25 {
				r25 += 1
			} else if participant.Runs == 50 {
				r50 += 1
			} else if participant.Runs == 100 {
				r100 += 1
			} else if participant.Runs == 250 {
				r250 += 1
			} else if participant.Runs == 500 {
				r500 += 1
			}
		}
		v1 := 0
		v25 := 0
		v50 := 0
		v100 := 0
		v250 := 0
		v500 := 0
		for _, participant := range run.Volunteers {
			parkrunner := &parkrun.Parkrunner{participant.Id, participant.Name, run.Time, -1, -1, -1, nil}
			if err := parkrunner.FetchMissingStats(run.Time); err != nil {
				panic(err)
			}

			if parkrunner.Vols == 1 {
				v1 += 1
			} else if parkrunner.Vols == 25 {
				v25 += 1
			} else if parkrunner.Vols == 50 {
				v50 += 1
			} else if parkrunner.Vols == 100 {
				v100 += 1
			} else if parkrunner.Vols == 250 {
				v250 += 1
			} else if parkrunner.Vols == 500 {
				v500 += 1
			}
		}

		if options.fancy {
			printFancy(event, run, r500, r250, r100, r50, r25, r1, pb, firstEvent, v500, v250, v100, v50, v25, v1)
			continue
		}

		fmt.Printf("%s #%d %s\n", event.Name, run.Index, run.Time.Format("2006-01-02"))
		fmt.Printf("Runners: %d\n", len(run.Runners))
		if r500 > 0 {
			fmt.Printf("- r500: %d\n", r500)
		}
		if r250 > 0 {
			fmt.Printf("- r250: %d\n", r250)
		}
		if r100 > 0 {
			fmt.Printf("- r100: %d\n", r100)
		}
		if r50 > 0 {
			fmt.Printf("- r50: %d\n", r50)
		}
		if r25 > 0 {
			fmt.Printf("- r25: %d\n", r25)
		}
		if r1 > 0 {
			fmt.Printf("- r1: %d\n", r1)
		}
		if firstEvent > 0 {
			fmt.Printf("- first @ event: %d\n", firstEvent)
		}
		if pb > 0 {
			fmt.Printf("- pb: %d\n", pb)
		}
		fmt.Printf("Volunteers: %d\n", len(run.Volunteers))
		if v500 > 0 {
			fmt.Printf("- v500: %d\n", v500)
		}
		if v250 > 0 {
			fmt.Printf("- v250: %d\n", v250)
		}
		if v100 > 0 {
			fmt.Printf("- v100: %d\n", v100)
		}
		if v50 > 0 {
			fmt.Printf("- v50: %d\n", v50)
		}
		if v25 > 0 {
			fmt.Printf("- v25: %d\n", v25)
		}
		if v1 > 0 {
			fmt.Printf("- v1: %d\n", v1)
		}
		fmt.Printf("Results: https://%s/%s/results/%d/\n", event.CountryUrl, event.Id, run.Index)
	}
}
