package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] [EVENTID...]

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload bool
	outdir      string
	country     string
	eventIds    []string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	outdir := flag.String("outdir", "html", "select output directory")
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
		*forceReload, *outdir, *country, flag.Args(),
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

type Milestone struct {
	Parkrunner *parkrun.Parkrunner
	NextRun    bool
	NextVol    bool
	Active     string
}

func printEvent(event *parkrun.Event, events []*parkrun.Event, outdir string, t *template.Template) {
	if err := os.MkdirAll(outdir, 0770); err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("%s/%s.html", outdir, event.Id)
	out, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	stats := event.GetStats()
	var run *parkrun.Run
	if len(event.Runs) > 0 {
		run = event.Runs[len(event.Runs)-1]
	}

	parkrunners, examinedRuns, err := event.GetActiveParkrunners(0.3, 10)
	if err != nil {
		panic(err)
	}
	var milestones []Milestone
	for _, p := range parkrunners {
		mr := parkrun.Milestone(p.Runs + 1)
		mv := parkrun.Milestone(p.Vols + 1)
		if mr || mv {
			active := fmt.Sprintf("%.0f%%", float64(100*len(p.Active))/float64(examinedRuns))
			milestones = append(milestones, Milestone{p, mr, mv, active})
		}
	}

	data := struct {
		Event          *parkrun.Event
		Stats          *parkrun.EventStats
		Run            *parkrun.Run
		NextMilestones []Milestone
		Events         []*parkrun.Event
	}{
		Event:          event,
		Stats:          stats,
		Run:            run,
		NextMilestones: milestones,
		Events:         events,
	}

	err = t.Execute(out, data)
	if err != nil {
		panic(err)
	}
}

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	t, err := template.ParseFiles("cmd/webgen/event.html")
	if err != nil {
		panic(err)
	}

	events := getEvents(options.eventIds, options.country)
	for _, event := range events {
		if err := event.Complete(); err != nil {
			panic(err)
		}

		printEvent(event, events, options.outdir, t)
	}
}
