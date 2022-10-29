package parkrun

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/biter777/countries"
)

type Event struct {
	Id         string
	Name       string
	CountryUrl string
	Country    string
	LastRun    int64
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
	buf, err := DownloadAndRead("https://images.parkrun.com/events.json", "events.json")
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

		eventList = append(eventList, &Event{eventId, eventName, countryUrl, country, -1})
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

func (event *Event) getNumberOfRuns() (uint64, error) {
	url := fmt.Sprintf("https://%s/%s/results/eventhistory/", event.CountryUrl, event.Id)
	fileName := fmt.Sprintf("%s/%s/eventhistory", event.CountryUrl, event.Id)
	buf, err := DownloadAndRead(url, fileName)
	if err != nil {
		return 0, err
	}

	match := patternNumberOfRuns.FindStringSubmatch(buf)
	if match == nil {
		// return 0, fmt.Errorf("%s: cannot find number of runs", event.Id)
		return 0, nil
	}

	count, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("%s: invalid number of runs: %s", event.Id, match[1])
	}

	return uint64(count), nil
}

var patternParkrunnerRow = regexp.MustCompile(`<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="[^"]*" data-club="[^"]*" data-gender="[^"]*" data-position="[^"]*" data-runs="([^"]*)" data-vols="([^"]*)" data-agegrade="[^"]*" data-achievement="[^"]*"><td class="Results-table-td Results-table-td--position">[^<]*</td><td class="Results-table-td Results-table-td--name"><div class="compact"><a href="[^"]*/(\d+)"`)
var patternVolunteer = regexp.MustCompile(`<a href='\./athletehistory/\?athleteNumber=(\d+)'>([^<]+)</a>`)

func (event *Event) getParkrunnersFromRun(runIndex uint64, parkrunners map[string]*Parkrunner) (map[string]*Parkrunner, error) {
	url := fmt.Sprintf("https://%s/%s/results/%d/", event.CountryUrl, event.Id, runIndex)
	fileName := fmt.Sprintf("%s/%s/%d", event.CountryUrl, event.Id, runIndex)
	buf, err := DownloadAndRead(url, fileName)
	if err != nil {
		return parkrunners, err
	}

	junior := event.IsJuniorParkrun()

	matches := patternParkrunnerRow.FindAllStringSubmatch(buf, -1)
	for _, match := range matches {
		name := html.UnescapeString(match[1])
		runs, err := strconv.Atoi(match[2])
		if err != nil {
			return parkrunners, err
		}
		vols, err := strconv.Atoi(match[3])
		if err != nil {
			return parkrunners, err
		}
		id := match[4]

		if junior {
			parkrunners = updateParkrunner(parkrunners, id, name, -1, int64(runs), int64(vols), runIndex)
		} else {
			parkrunners = updateParkrunner(parkrunners, id, name, int64(runs), -1, int64(vols), runIndex)
		}
	}

	matchesV := patternVolunteer.FindAllStringSubmatch(buf, -1)
	for _, match := range matchesV {
		id := match[1]
		name := html.UnescapeString(match[2])

		parkrunners = updateParkrunner(parkrunners, id, name, -1, -1, -1, runIndex)
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
	event.LastRun = int64(numberOfRuns)

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

	updatesNeeded := 0
	for _, parkrunner := range parkrunners {
		if len(parkrunner.Active) >= int(activeLimit) {
			if parkrunner.needsUpdate() {
				updatesNeeded += 1
			}
		}
	}
	fmt.Printf("-- Updating %d incomplete parkrunners...\n", updatesNeeded)

	activeParkrunners := make([]*Parkrunner, 0)
	for _, parkrunner := range parkrunners {
		if len(parkrunner.Active) >= int(activeLimit) {
			if err = parkrunner.fetchMissingStats(); err != nil {
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
