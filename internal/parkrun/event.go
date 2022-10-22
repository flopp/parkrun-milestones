package parkrun

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"

	download "github.com/flopp/parkrun-milestones/internal/download"
	file "github.com/flopp/parkrun-milestones/internal/file"
)

type Event struct {
	Id         string
	Name       string
	CountryUrl string
}

func AllEvents() ([]*Event, error) {
	eventList := make([]*Event, 0)

	if err := download.DownloadFile("https://images.parkrun.com/events.json", ".data/events.json", MaxFileAge); err != nil {
		return nil, err
	}

	buf, err := file.ReadFile(".data/events.json")
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

		eventList = append(eventList, &Event{eventId, eventName, countryUrl})
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

func (event *Event) getNumberOfRuns() (uint64, error) {
	url := fmt.Sprintf("https://%s/%s/results/eventhistory/", event.CountryUrl, event.Id)
	filePath := fmt.Sprintf(".data/%s/%s/eventhistory", event.CountryUrl, event.Id)
	if err := download.DownloadFile(url, filePath, MaxFileAge); err != nil {
		return 0, err
	}

	buf, err := file.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	pattern := regexp.MustCompile("<td class=\"Results-table-td Results-table-td--position\"><a href=\"\\.\\./(\\d+)\">(\\d+)</a></td>")
	match := pattern.FindStringSubmatch(buf)
	if match == nil {
		return 0, fmt.Errorf("%s: cannot find number of runs", event.Id)
	}

	count, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("%s: invalid number of runs: %s", match[1])
	}

	return uint64(count), nil
}

func (event *Event) getParkrunnersFromRun(runIndex uint64, parkrunners map[string]*Parkrunner) (map[string]*Parkrunner, error) {
	url := fmt.Sprintf("https://%s/%s/results/%d/", event.CountryUrl, event.Id, runIndex)
	filePath := fmt.Sprintf(".data/%s/%s/%d", event.CountryUrl, event.Id, runIndex)
	if err := download.DownloadFile(url, filePath, MaxFileAge); err != nil {
		return parkrunners, err
	}

	buf, err := file.ReadFile(filePath)
	if err != nil {
		return parkrunners, err
	}

	pattern := regexp.MustCompile(`<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="[^"]*" data-club="[^"]*" data-gender="[^"]*" data-position="[^"]*" data-runs="([^"]*)" data-vols="([^"]*)" data-agegrade="[^"]*" data-achievement="[^"]*"><td class="Results-table-td Results-table-td--position">[^<]*</td><td class="Results-table-td Results-table-td--name"><div class="compact"><a href="[^"]*/(\d+)"`)
	matches := pattern.FindAllStringSubmatch(buf, -1)
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

		if parkrunners, err = updateParkrunner(parkrunners, id, name, int64(runs), int64(vols), runIndex); err != nil {
			return parkrunners, err
		}
	}

	patternV := regexp.MustCompile(`<a href='\./athletehistory/\?athleteNumber=(\d+)'>([^<]+)</a>`)
	matchesV := patternV.FindAllStringSubmatch(buf, -1)
	for _, match := range matchesV {
		id := match[1]
		name := html.UnescapeString(match[2])

		if parkrunners, err = updateParkrunner(parkrunners, id, name, -1, -1, runIndex); err != nil {
			return parkrunners, err
		}
	}
	return parkrunners, nil
}

func (event *Event) GetActiveParkrunners(minActiveRatio float64, examineNumberOfRuns uint64) ([]*Parkrunner, uint64, error) {
	numberOfRuns, err := event.getNumberOfRuns()
	if err != nil {
		return nil, 0, err
	}
	fmt.Printf("Latest run at %s: #%d\n", event.Name, numberOfRuns)
	if numberOfRuns == 0 {
		return nil, 0, fmt.Errorf("%s: no runs", event.Id)
	}

	toIndex := numberOfRuns
	fromIndex := uint64(1)
	if toIndex > examineNumberOfRuns {
		fromIndex = 1 + toIndex - examineNumberOfRuns
	}

	parkrunners := make(map[string]*Parkrunner)

	for index := fromIndex; index <= toIndex; index += 1 {
		fmt.Printf("-- Examining run #%d\n", index)
		if parkrunners, err = event.getParkrunnersFromRun(index, parkrunners); err != nil {
			return nil, 0, err
		}
	}

	activeLimit := int64(minActiveRatio * (float64(1 + toIndex - fromIndex)))

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
