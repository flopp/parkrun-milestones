package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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
	table       bool
	country     string
	eventIds    []string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	fancy := flag.Bool("fancy", false, "fancy formatting using emoji")
	table := flag.Bool("table", false, "csv style output")
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
		*forceReload, *fancy, *table, *country, flag.Args(),
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
	if n == 0 {
		return
	}
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
	fmt.Printf("%s #️⃣ %d\n", event.Name, int(run.Index))
	ps(run.Time.Format("02.01.2006"), "📅", "")
	ps("", "⛅", "Wetter / weather")
	ps("", "🎁", "Special")
	pi(len(run.Runners), "🏃", "Teilnehmer / runners")
	pi(pb, "⏱️", "Neue Bestzeiten / new PB")
	pi(firstEvent, "🌍", "Besucher / visitors")
	pi(r1, "⭐️", "Neue Teilnehmer / first-time runners")
	pi(len(run.Volunteers), "🦺", "Helfende / volunteers")
	pi(v1, "⭐️", "Neue Helfende / first-time volunteers")
	if (r25 + r50 + r100 + r250 + r500 + v25 + v50 + v100 + v250 + v500) > 0 {
		m := make([]string, 0)
		if r25 > 0 {
			m = append(m, fmt.Sprintf("%dxR25", r25))
		}
		if r50 > 0 {
			m = append(m, fmt.Sprintf("%dxR50", r50))
		}
		if r100 > 0 {
			m = append(m, fmt.Sprintf("%dxR100", r100))
		}
		if r250 > 0 {
			m = append(m, fmt.Sprintf("%dxR250", r250))
		}
		if r500 > 0 {
			m = append(m, fmt.Sprintf("%dxR500", r500))
		}
		if v25 > 0 {
			m = append(m, fmt.Sprintf("%dxV25", v25))
		}
		if v50 > 0 {
			m = append(m, fmt.Sprintf("%dxV50", v50))
		}
		if v100 > 0 {
			m = append(m, fmt.Sprintf("%dxV100", v100))
		}
		if v250 > 0 {
			m = append(m, fmt.Sprintf("%dxV250", v250))
		}
		if v500 > 0 {
			m = append(m, fmt.Sprintf("%dxV500", v500))
		}
		ps(strings.Join(m, ", "), "🏆", "Milestones")
	}
	fmt.Printf("\nhttps://%s/%s/results/%d/\n", event.CountryUrl, event.Id, run.Index)
	fmt.Println("#parkrun #running #laufen #mastodonlauftreff")
}

func fmtAgeGroup(ageGroup string) string {
	return ageGroup
}

func fmtTime(t time.Duration) string {
	if t == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d", int64(t.Seconds()))
}

func printTable(event *parkrun.Event, run *parkrun.Run) {
	fmt.Printf("%s #%d %s\n", event.Name, run.Index, run.Time.Format("2006-01-02"))

	fmt.Println("\nRunners")
	fmt.Println("Name;Age Group;Total Runs;Finishing Time;Special")
	for _, participant := range run.Runners {
		fmt.Printf("%s;", participant.Name)
		if participant.Id != "" {
			fmt.Printf("%s;", fmtAgeGroup(participant.AgeGroup))
			fmt.Printf("%d;", participant.Runs)
			fmt.Printf("%s;", fmtTime(participant.Time))
		} else {
			fmt.Printf("n/a;n/a;n/a;")
		}

		if participant.Achievement == parkrun.AchievementFirst {
			if participant.Runs == 1 {
				fmt.Printf("first parkrun")
			} else {
				fmt.Printf("first time at %s", event.Name)
			}
		} else if participant.Achievement == parkrun.AchievementPB {
			fmt.Printf("new personal best at %s", event.Name)
		}

		fmt.Println("")
	}

	fmt.Println("\nVolunteers")
	fmt.Println("Name;Total Volunteerings")
	for _, participant := range run.Volunteers {
		parkrunner := &parkrun.Parkrunner{Id: participant.Id, Name: participant.Name, AgeGroup: "??", DataTime: run.Time, Runs: -1, JuniorRuns: -1, Vols: -1, Active: nil}
		if err := parkrunner.FetchMissingStats(run.Time); err != nil {
			panic(err)
		}
		fmt.Printf("%s;%d\n", participant.Name, parkrunner.Vols)
	}
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

		stats := event.GetStats()
		if stats == nil {
			continue
		}

		run := event.Runs[len(event.Runs)-1]
		firstEvent := len(stats.FirstEvent)
		pb := len(stats.PB)
		r1 := len(stats.R1)
		r25 := len(stats.R25)
		r50 := len(stats.R50)
		r100 := len(stats.R100)
		r250 := len(stats.R250)
		r500 := len(stats.R500)
		v1 := len(stats.V1)
		v25 := len(stats.V25)
		v50 := len(stats.V50)
		v100 := len(stats.V100)
		v250 := len(stats.V250)
		v500 := len(stats.V500)

		if options.table {
			printTable(event, run)
			continue
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
