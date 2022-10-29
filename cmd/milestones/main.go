package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func formatMilestone(number int64) string {
	s := fmt.Sprintf("%d", number)
	if parkrun.Milestone(number + 1) {
		return "*" + s
	}
	return s
}

const (
	usage = `USAGE: %s [OPTIONS...] [EVENTID...]
Determine the milestone candidates of the specified event(s) or
of all events of a county (if -country NAME is given). 

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload    bool
	minActiveRatio float64
	runs           uint64
	country        string
	eventIds       []string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	minActiveRatio := flag.Float64("active", 0.3, "minimum active ratio")
	runs := flag.Uint64("runs", 10, "consider at most the X latest runs of the event")
	country := flag.String("country", "", "select all events of the specified country")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *minActiveRatio < 0.0 || *minActiveRatio > 1.0 {
		panic(fmt.Errorf("invalid -active value: %f; must be between 0 and 1", *minActiveRatio))
	}
	if *country == "" && len(flag.Args()) == 0 {
		panic("You have to specify either one or more EVENTID... or -country NAME")
	}
	if *country != "" && len(flag.Args()) != 0 {
		panic("You must not specify both one or more EVENTID... and -country NAME")
	}

	return CommandLineOptions{
		*forceReload, *minActiveRatio, *runs, *country, flag.Args(),
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

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	events := getEvents(options.eventIds, options.country)
	for _, event := range events {
		fmt.Printf("-- Fetching data for %s...\n", event.Name)
		parkrunners, examinedRuns, err := event.GetActiveParkrunners(options.minActiveRatio, options.runs)
		if err != nil {
			panic(err)
		}

		junior := event.IsJuniorParkrun()

		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(fmt.Sprintf("Expected Milestones at\n%s\nRun #%d", event.Name, event.LastRun+1))
		t.AppendHeader(table.Row{"Name", "Runs", "Vols", "Active"})
		if junior {
			for _, parkrunner := range parkrunners {
				if parkrun.Milestone(parkrunner.JuniorRuns+1) || parkrun.Milestone(parkrunner.Vols+1) {
					t.AppendRow([]interface{}{parkrunner.Name, formatMilestone(parkrunner.JuniorRuns), formatMilestone(parkrunner.Vols), fmt.Sprintf("%d/%d", len(parkrunner.Active), examinedRuns)})
				}
			}
		} else {
			for _, parkrunner := range parkrunners {
				if parkrun.Milestone(parkrunner.Runs+1) || parkrun.Milestone(parkrunner.Vols+1) {
					t.AppendRow([]interface{}{parkrunner.Name, formatMilestone(parkrunner.Runs), formatMilestone(parkrunner.Vols), fmt.Sprintf("%d/%d", len(parkrunner.Active), examinedRuns)})
				}
			}
		}
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, WidthMin: 30, AlignHeader: text.AlignLeft},
			{Number: 2, WidthMin: 4, Align: text.AlignRight, AlignHeader: text.AlignLeft},
			{Number: 3, WidthMin: 4, Align: text.AlignRight, AlignHeader: text.AlignLeft},
			{Number: 4, WidthMin: 5, Align: text.AlignRight, AlignHeader: text.AlignLeft},
		})
		t.Render()
		fmt.Println()
	}
}
