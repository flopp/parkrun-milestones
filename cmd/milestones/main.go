package main

import (
	"flag"
	"fmt"
	"os"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
	"github.com/jedib0t/go-pretty/v6/table"
)

func formatMilestone(number int64) string {
	s := fmt.Sprintf("%d", number)
	if parkrun.Milestone(number + 1) {
		return "*" + s + "*"
	}
	return s
}

func listEvents() {
	fmt.Printf("EVENTS\n")
	eventList, err := parkrun.AllEvents()
	if err != nil {
		panic(err)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Event Id", "Event Name", "Country"})
	for _, event := range eventList {
		country, err := parkrun.LookupCountry(event.CountryUrl)
		if err != nil {
			panic(err)
		}
		t.AppendRow([]interface{}{event.Id, event.Name, country})
	}
	t.Render()
}

func findMilestones(eventId string, minActiveRatio float64, examineMaxRuns uint64) {
	fmt.Printf("EVENT: %s\n", eventId)
	event, err := parkrun.LookupEvent(eventId)
	if err != nil {
		panic(err)
	}

	parkrunners, examinedRuns, err := event.GetActiveParkrunners(minActiveRatio, examineMaxRuns)
	if err != nil {
		panic(err)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Runs", "Vols", "Active"})
	for _, parkrunner := range parkrunners {
		if parkrunner.IsMilstoneCandidate() {
			t.AppendRow([]interface{}{parkrunner.Name, formatMilestone(parkrunner.Runs), formatMilestone(parkrunner.Vols), fmt.Sprintf("%d/%d", len(parkrunner.Active), examinedRuns)})
		}
	}
	t.Render()
}

const (
	usage = `USAGE: %s [OPTIONS...] [EVENTID]
If no EVENTID is specified, list all parkrun events (event id, event name).
Otherwise determine the milestone candidates of the specified event. 

OPTIONS:
`
)

func main() {
	forceReload := flag.Bool("force", false, "force reload of all data")
	minActiveRatio := flag.Float64("active", 0.3, "minimum active ratio")
	runs := flag.Uint64("runs", 10, "examine that many latest runs of the event")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *minActiveRatio < 0.0 || *minActiveRatio > 1.0 {
		panic(fmt.Errorf("invalid -active value: %f; must be between 0 and 1", *minActiveRatio))
	}
	if len(flag.Args()) > 1 {
		panic("Too many arguments")
	}

	if *forceReload {
		parkrun.MaxFileAge = 0
	}

	if len(flag.Args()) == 0 {
		listEvents()
	} else {
		eventId := flag.Arg(0)
		findMilestones(eventId, *minActiveRatio, *runs)
	}
}
