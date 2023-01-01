package parkrun

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/biter777/countries"
)

type Event struct {
	Id         string
	Name       string
	CountryUrl string
	Country    string
	IsComplete bool
	Runs       []*Run
}

func (event Event) NumberOfRuns() int {
	return len(event.Runs)
}

var byTLD map[string]string = nil
var patternCountryUrl = regexp.MustCompile(`^www\.parkrun.*(\.[^.]+)$`)

func lookupCountry(url string) (string, error) {
	match := patternCountryUrl.FindStringSubmatch(url)
	if match == nil {
		return "", fmt.Errorf("cannot extract TLD from %s", url)
	}
	tld := match[1]

	if len(byTLD) == 0 {
		byTLD = make(map[string]string)
		for _, countryCode := range countries.All() {
			byTLD[countryCode.Domain().String()] = countryCode.String()
		}
	}

	country, ok := byTLD[tld]
	if !ok {
		return "", fmt.Errorf("cannot determine country for %s (TLD=%s)", url, tld)
	}
	return country, nil
}

func AllEvents() ([]*Event, error) {
	buf, _, err := DownloadAndRead("https://images.parkrun.com/events.json", "events.json")
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(buf), &result); err != nil {
		return nil, err
	}

	countriesI, ok := result["countries"]
	if !ok {
		return nil, fmt.Errorf("cannot get 'countries' from 'events.json")
	}
	countries := countriesI.(map[string]interface{})

	countryLookup := make(map[string]string)
	for countryId, countryI := range countries {
		country := countryI.(map[string]interface{})

		urlI, ok := country["url"]
		if !ok {
			return nil, fmt.Errorf("cannot get 'countries/%s/url' from 'events.json", countryId)
		}

		countryLookup[countryId] = urlI.(string)
	}

	eventsI, ok := result["events"]
	if !ok {
		return nil, fmt.Errorf("cannot get 'events' from 'events.json")
	}
	events := eventsI.(map[string]interface{})

	featuresI, ok := events["features"]
	if !ok {
		return nil, fmt.Errorf("cannot get 'events/features' from 'events.json")
	}

	eventList := make([]*Event, 0)
	features := featuresI.([]interface{})
	for _, featureI := range features {
		feature := featureI.(map[string]interface{})
		propertiesI, ok := feature["properties"]
		if !ok {
			return nil, fmt.Errorf("cannot get 'events/features/properties' from 'events.json")
		}

		properties := propertiesI.(map[string]interface{})
		idI, ok := properties["eventname"]
		if !ok {
			return nil, fmt.Errorf("cannot get 'events/features/properties/eventname' from 'events.json")
		}
		nameI, ok := properties["EventLongName"]
		if !ok {
			return nil, fmt.Errorf("cannot get 'events/features/properties/EventLongName' from 'events.json")
		}
		countryCodeI, ok := properties["countrycode"]
		if !ok {
			return nil, fmt.Errorf("cannot get 'events/features/properties/countrycode' from 'events.json")
		}
		eventId := idI.(string)
		eventName := nameI.(string)
		countryCode := fmt.Sprintf("%.0f", countryCodeI.(float64))

		countryUrl, ok := countryLookup[countryCode]
		if !ok {
			return nil, fmt.Errorf("cannot get URL of contry '%s'", countryCode)
		}

		country, err := lookupCountry(countryUrl)
		if err != nil {
			country = "<UNKNOWN>"
		}

		eventList = append(eventList, &Event{eventId, eventName, countryUrl, country, false, nil})
	}

	sort.Slice(eventList, func(i, j int) bool {
		return eventList[i].Id < eventList[j].Id
	})
	return eventList, nil
}

func LookupEvent(eventId string) (*Event, error) {
	eventList, err := AllEvents()
	if err != nil {
		return nil, err
	}

	for _, event := range eventList {
		if event.Id == eventId {
			return event, nil
		}
	}

	return nil, fmt.Errorf("cannot find event '%s'", eventId)
}

func (event *Event) IsJuniorParkrun() bool {
	return strings.HasSuffix(event.Id, "-juniors")
}

var patternNumberOfRuns = regexp.MustCompile("<td class=\"Results-table-td Results-table-td--position\"><a href=\"\\.\\./(\\d+)\">(\\d+)</a></td>")
var patternRunRow = regexp.MustCompile(`<tr class="Results-table-row" data-parkrun="(\d+)" data-date="(\d+/\d+/\d+)" data-finishers="(\d+)" data-volunteers="(\d+)" data-male="([^"]*)" data-female="([^"]*)" data-maletime="(\d*)" data-femaletime="(\d*)">`)

func (event *Event) Complete() error {
	if event.IsComplete {
		return nil
	}

	url := fmt.Sprintf("https://%s/%s/results/eventhistory/", event.CountryUrl, event.Id)
	fileName := fmt.Sprintf("%s/%s/eventhistory", event.CountryUrl, event.Id)
	buf, _, err := DownloadAndRead(url, fileName)
	if err != nil {
		return err
	}

	match := patternNumberOfRuns.FindStringSubmatch(buf)
	if match == nil {
		event.IsComplete = true
		return nil
	}

	count, err := strconv.Atoi(match[1])
	if err != nil {
		return err
	}
	if count < 0 {
		return fmt.Errorf("%s: invalid number of runs: %s", event.Id, match[1])
	}

	event.Runs = make([]*Run, count)
	matches := patternRunRow.FindAllStringSubmatch(buf, -1)
	for _, match := range matches {
		index, err := strconv.Atoi(match[1])
		if err != nil {
			return err
		}
		if index <= 0 || index > count {
			return fmt.Errorf("%s: invalid run index: %s", event.Id, match[1])
		}

		date, err := time.Parse("02/01/2006", match[2])
		if err != nil {
			return fmt.Errorf("%s: invalid date: %s (%v)", event.Id, match[2], err)
		}

		if event.Runs[index-1] != nil {
			return fmt.Errorf("%s: duplicate run #%d", event.Id, index)
		}
		event.Runs[index-1] = CreateRun(event, uint64(index), date)
	}

	for index, run := range event.Runs {
		if run == nil {
			return fmt.Errorf("%s: missing run #%d", event.Id, index+1)
		}
	}

	event.IsComplete = true
	return nil
}

func (event *Event) getNumberOfRuns() (uint64, error) {
	err := event.Complete()
	if err != nil {
		return 0, err
	}

	return uint64(len(event.Runs)), nil
}

func (event *Event) getParkrunnersFromRun(runIndex uint64, parkrunners map[string]*Parkrunner) (map[string]*Parkrunner, error) {
	if runIndex < 1 || runIndex > uint64(len(event.Runs)) {
		return parkrunners, fmt.Errorf("%s: bad run #%d", event.Id, runIndex)
	}

	run := event.Runs[runIndex-1]
	err := run.Complete()
	if err != nil {
		return parkrunners, err
	}

	junior := event.IsJuniorParkrun()

	for _, participant := range run.Runners {
		if junior {
			parkrunners = updateParkrunner(parkrunners, participant.Id, participant.Name, run.DataTime, -1, participant.Runs, participant.Vols, runIndex)
		} else {
			parkrunners = updateParkrunner(parkrunners, participant.Id, participant.Name, run.DataTime, participant.Runs, -1, participant.Vols, runIndex)
		}
	}

	for _, participant := range run.Volunteers {
		parkrunners = updateParkrunner(parkrunners, participant.Id, participant.Name, run.DataTime, -1, -1, -1, runIndex)
	}

	return parkrunners, nil
}

func (event *Event) GetActiveParkrunners(minActiveRatio float64, examineNumberOfRuns uint64) ([]*Parkrunner, uint64, error) {
	numberOfRuns, err := event.getNumberOfRuns()
	if err != nil {
		return nil, 0, err
	}
	if numberOfRuns == 0 {
		return nil, 0, nil
	}

	toIndex := numberOfRuns
	fromIndex := uint64(1)
	if toIndex > examineNumberOfRuns {
		fromIndex = 1 + toIndex - examineNumberOfRuns
	}

	parkrunners := make(map[string]*Parkrunner)

	fmt.Printf("-- Fetching the latest %d result lists...\n", 1+toIndex-fromIndex)
	for index := fromIndex; index <= toIndex; index += 1 {
		if parkrunners, err = event.getParkrunnersFromRun(index, parkrunners); err != nil {
			return nil, 0, err
		}
	}

	activeLimit := int64(minActiveRatio * (float64(1 + toIndex - fromIndex)))
	lastRunDate := event.Runs[len(event.Runs)-1].Time
	updatesNeeded := 0
	for _, parkrunner := range parkrunners {
		if len(parkrunner.Active) >= int(activeLimit) {
			if parkrunner.NeedsUpdate() {
				updatesNeeded += 1
			}
		}
	}
	fmt.Printf("-- Updating %d incomplete parkrunners...\n", updatesNeeded)

	activeParkrunners := make([]*Parkrunner, 0)
	for _, parkrunner := range parkrunners {
		if len(parkrunner.Active) >= int(activeLimit) {
			if err = parkrunner.FetchMissingStats(lastRunDate); err != nil {
				return nil, 0, err
			}
			activeParkrunners = append(activeParkrunners, parkrunner)
		}
	}

	sort.Slice(activeParkrunners, func(i, j int) bool {
		return activeParkrunners[i].Name < activeParkrunners[j].Name
	})
	return activeParkrunners, (1 + toIndex - fromIndex), nil
}

type EventStats struct {
	FirstEvent []*Participant
	PB         []*Participant
	R1         []*Participant
	R25        []*Participant
	R50        []*Participant
	R100       []*Participant
	R250       []*Participant
	R500       []*Participant
	V1         []*Participant
	V25        []*Participant
	V50        []*Participant
	V100       []*Participant
	V250       []*Participant
	V500       []*Participant
}

func (event *Event) GetStats() *EventStats {
	if len(event.Runs) == 0 {
		fmt.Printf("No runs at %s\n", event.Name)
		return nil
	}

	run := event.Runs[len(event.Runs)-1]
	if err := run.Complete(); err != nil {
		panic(err)
	}

	stats := EventStats{}
	for _, participant := range run.Runners {
		if participant.Achievement == AchievementFirst {
			if participant.Runs == 1 {
				stats.R1 = append(stats.R1, participant)
			} else {
				stats.FirstEvent = append(stats.FirstEvent, participant)
			}
		} else if participant.Achievement == AchievementPB {
			stats.PB = append(stats.PB, participant)
		}
		if participant.Runs == 25 {
			stats.R25 = append(stats.R25, participant)
		} else if participant.Runs == 50 {
			stats.R50 = append(stats.R50, participant)
		} else if participant.Runs == 100 {
			stats.R100 = append(stats.R100, participant)
		} else if participant.Runs == 250 {
			stats.R250 = append(stats.R250, participant)
		} else if participant.Runs == 500 {
			stats.R500 = append(stats.R500, participant)
		}
	}
	for _, participant := range run.Volunteers {
		parkrunner := &Parkrunner{participant.Id, participant.Name, run.Time, -1, -1, -1, nil}
		if err := parkrunner.FetchMissingStats(run.Time); err != nil {
			panic(err)
		}

		if parkrunner.Vols == 1 {
			stats.V1 = append(stats.V1, participant)
		} else if parkrunner.Vols == 25 {
			stats.V25 = append(stats.V25, participant)
		} else if parkrunner.Vols == 50 {
			stats.V50 = append(stats.V50, participant)
		} else if parkrunner.Vols == 100 {
			stats.V100 = append(stats.V100, participant)
		} else if parkrunner.Vols == 250 {
			stats.V250 = append(stats.V250, participant)
		} else if parkrunner.Vols == 500 {
			stats.V500 = append(stats.V500, participant)
		}
	}

	return &stats
}
