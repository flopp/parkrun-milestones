package parkrun

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"time"
)

type AchievementEnum int64

const (
	AchievementNone AchievementEnum = iota
	AchievementFirst
	AchievementPB
)

type Participant struct {
	Id          string
	Name        string
	Runs        int64
	Vols        int64
	Achievement AchievementEnum
}

func ParseAchievement(s string, country string) (AchievementEnum, error) {
	if s == "" {
		return AchievementNone, nil
	}

	var first = [...]string{
		"First Timer!",     // UK, SA, CA, US, NZ, IE, MY, AUS
		"Erstläufer!",      // Germany
		"Première perf' !", // France
		"Prima volta!",     // Italy
		"Debut!",           // Sweden
		"Debiutant",        // Poland
		"Nieuwe loper!",    // Netherlands
		"Første gang!",     // Denmark
		"初参加!",             // Japan
	}
	var pb = [...]string{
		"New PB!",           // UK, SA, CA, US, NZ, IE, MY, AUS
		"Neue PB!",          // Germany
		"Meilleure perf' !", // France
		"Nuovo PB!",         // Italy
		"Nytt PB!",          // Sweden
		"Nowy PB!",          // Poland
		"Nieuw PR!",         // Netherlands
		"Ny PB!",            // Denmark
		"自己ベスト!",            // Japan
	}

	for _, pattern := range first {
		if pattern == s {
			return AchievementFirst, nil
		}
	}
	for _, pattern := range pb {
		if pattern == s {
			return AchievementPB, nil
		}
	}

	return AchievementNone, fmt.Errorf("cannot parse achievement: %s", s)
}

type Run struct {
	Parent     *Event
	Index      uint64
	Time       time.Time
	IsComplete bool
	DataTime   time.Time
	Runners    []*Participant
	Volunteers []*Participant
}

func CreateRun(parent *Event, index uint64, t time.Time) *Run {
	return &Run{parent, index, t, false, time.Time{}, nil, nil}
}

var patternRunnerRow0 = regexp.MustCompile(`<tr class="Results-table-row" [^<]*><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact">(<a href="[^"]*/\d+")?`)
var patternRunnerRow = regexp.MustCompile(`^<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="[^"]*" data-club="[^"]*" data-gender="[^"]*" data-position="\d+" data-runs="(\d+)" data-vols="(\d+)" data-agegrade="[^"]*" data-achievement="([^"]*)"><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact"><a href="[^"]*/(\d+)"`)
var patternRunnerRowUnknown = regexp.MustCompile(`^<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="" data-club="" data-position="\d+" data-runs="0" data-agegrade="0" data-achievement=""><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact">.*`)
var patternVolunteerRow = regexp.MustCompile(`<a href='\./athletehistory/\?athleteNumber=(\d+)'>([^<]+)</a>`)

func (run *Run) Complete() error {
	if run.IsComplete {
		return nil
	}

	event := run.Parent
	url := fmt.Sprintf("https://%s/%s/results/%d/", event.CountryUrl, event.Id, run.Index)
	fileName := fmt.Sprintf("%s/%s/%d", event.CountryUrl, event.Id, run.Index)
	buf, dataTime, err := DownloadAndRead(url, fileName)
	if err != nil {
		return err
	}
	reNewline := regexp.MustCompile(`\r?\n`)
	buf = reNewline.ReplaceAllString(buf, " ")

	run.IsComplete = true
	run.DataTime = dataTime

	matchesR0 := patternRunnerRow0.FindAllStringSubmatch(buf, -1)
	for _, match0 := range matchesR0 {
		if match := patternRunnerRow.FindStringSubmatch(match0[0]); match != nil {
			name := html.UnescapeString(match[1])
			runs, err := strconv.Atoi(match[2])
			if err != nil {
				return err
			}
			vols, err := strconv.Atoi(match[3])
			if err != nil {
				return err
			}

			achievement, err := ParseAchievement(match[4], run.Parent.Country)
			if err != nil {
				return err
			}

			id := match[5]
			run.Runners = append(run.Runners, &Participant{id, name, int64(runs), int64(vols), achievement})
			continue
		}

		if match := patternRunnerRowUnknown.FindStringSubmatch(match0[0]); match != nil {
			name := html.UnescapeString(match[1])
			run.Runners = append(run.Runners, &Participant{"", name, 0, 0, AchievementNone})
			continue
		}

		return fmt.Errorf("cannot parse table row: %s", match0[0])
	}

	matchesV := patternVolunteerRow.FindAllStringSubmatch(buf, -1)
	for _, match := range matchesV {
		id := match[1]
		name := html.UnescapeString(match[2])

		run.Volunteers = append(run.Volunteers, &Participant{id, name, -1, -1, AchievementNone})
	}

	return nil
}
