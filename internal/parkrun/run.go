package parkrun

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AchievementEnum int64

const (
	AchievementNone AchievementEnum = iota
	AchievementFirst
	AchievementPB
)

const (
	SEX_UNKNOWN = iota
	SEX_FEMALE
	SEX_MALE
)

type Participant struct {
	Id          string
	Name        string
	AgeGroup    string
	Sex         int
	Runs        int64
	Vols        int64
	Time        time.Duration
	Achievement AchievementEnum
}

var reAgeGroup1 = regexp.MustCompile(`^[A-Z]([fFmMwW])(\d+-\d+)$`)
var reAgeGroup2 = regexp.MustCompile(`^[A-Z]([fFmMwW])(\d+)$`)
var reAgeGroup3 = regexp.MustCompile(`^[A-Z]([fFmMwW])(---)$`)
var reAgeGroup4 = regexp.MustCompile(`^([fFmMwW])(WC)$`)

func ParseAgeGroup(s string) (string, int, error) {
	if s == "" {
		return "??", SEX_UNKNOWN, nil
	}
	if match := reAgeGroup1.FindStringSubmatch(s); match != nil {
		if match[1] == "f" || match[1] == "F" || match[1] == "w" || match[1] == "W" {
			return match[2], SEX_FEMALE, nil
		}
		return match[2], SEX_MALE, nil
	}
	if match := reAgeGroup2.FindStringSubmatch(s); match != nil {
		if match[1] == "f" || match[1] == "F" || match[1] == "w" || match[1] == "W" {
			return match[2], SEX_FEMALE, nil
		}
		return match[2], SEX_MALE, nil
	}
	if match := reAgeGroup3.FindStringSubmatch(s); match != nil {
		if match[1] == "f" || match[1] == "F" || match[1] == "w" || match[1] == "W" {
			return match[2], SEX_FEMALE, nil
		}
		return match[2], SEX_MALE, nil
	}
	if match := reAgeGroup4.FindStringSubmatch(s); match != nil {
		if match[1] == "f" || match[1] == "F" || match[1] == "w" || match[1] == "W" {
			return match[2], SEX_FEMALE, nil
		}
		return match[2], SEX_MALE, nil
	}

	return s, SEX_UNKNOWN, fmt.Errorf("unknown age group: %s", s)
}

func ParseAchievement(s string, country string) (AchievementEnum, error) {
	if s == "" {
		return AchievementNone, nil
	}

	var first = [...]string{
		"First Timer!", // UK, SA, CA, US, NZ, IE, MY, AUS
		"Erstläufer!",  // Germany
		"Erstteilnahme!",
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
		if fmt.Sprintf("[parkrun_translate phrase='%s']", pattern) == s {
			return AchievementFirst, nil
		}
	}
	for _, pattern := range pb {
		if pattern == s {
			return AchievementPB, nil
		}
		if fmt.Sprintf("[parkrun_translate phrase='%s']", pattern) == s {
			return AchievementPB, nil
		}
	}

	return AchievementNone, fmt.Errorf("cannot parse achievement: %s", s)
}

type Run struct {
	Parent      *Event
	Index       uint64
	Time        time.Time
	IsComplete  bool
	DataTime    time.Time
	NRunners    uint64
	NVolunteers uint64
	Runners     []*Participant
	Volunteers  []*Participant
}

func CreateRun(parent *Event, index uint64, t time.Time, nFinishers, nVolunteers uint64) *Run {
	return &Run{parent, index, t, false, time.Time{}, nFinishers, nVolunteers, nil, nil}
}

var patternDateIndex = regexp.MustCompile(`<h3><span class="format-date">([^<]+)</span><span class="spacer"> | </span><span>#([0-9]+)</span></h3>`)
var patternRunnerRow0 = regexp.MustCompile(`<tr class="Results-table-row" [^<]*><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact">(<a href="[^"]*/\d+")?.*?</tr>`)
var patternRunnerRow = regexp.MustCompile(`^<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="([^"]*)" data-club="[^"]*" data-gender="[^"]*" data-position="\d+" data-runs="(\d+)" data-vols="(\d+)" data-agegrade="[^"]*" data-achievement="([^"]*)"><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact"><a href="[^"]*/(\d+)"`)
var patternRunnerRowUnknown = regexp.MustCompile(`^<tr class="Results-table-row" data-name="([^"]*)" data-agegroup="" data-club="" data-position="\d+" data-runs="0" data-agegrade="0" data-achievement=""><td class="Results-table-td Results-table-td--position">\d+</td><td class="Results-table-td Results-table-td--name"><div class="compact">.*`)
var patternTime = regexp.MustCompile(`Results-table-td--time[^"]*&#10;                      "><div class="compact">(\d?:?\d\d:\d\d)</div>`)
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

			ageGroup, sex, err := ParseAgeGroup(match[2])
			if err != nil {
				return err
			}

			runs, err := strconv.Atoi(match[3])
			if err != nil {
				return err
			}

			vols, err := strconv.Atoi(match[4])
			if err != nil {
				return err
			}

			achievement, err := ParseAchievement(match[5], run.Parent.Country)
			if err != nil {
				return err
			}

			id := match[6]

			var runTime time.Duration = 0
			if matchTime := patternTime.FindStringSubmatch(match0[0]); matchTime != nil {
				split := strings.Split(matchTime[1], ":")
				if len(split) == 3 {
					t, err := time.ParseDuration(fmt.Sprintf("%sh%sm%ss", split[0], split[1], split[2]))
					if err != nil {
						panic(err)
					}
					runTime = t
				} else if len(split) == 2 {
					t, err := time.ParseDuration(fmt.Sprintf("%sm%ss", split[0], split[1]))
					if err != nil {
						panic(err)
					}
					runTime = t
				} else {
					panic(fmt.Errorf("cannot parse duration: %s", matchTime[1]))
				}
			}

			run.Runners = append(run.Runners, &Participant{id, name, ageGroup, sex, int64(runs), int64(vols), runTime, achievement})
			continue
		}

		if match := patternRunnerRowUnknown.FindStringSubmatch(match0[0]); match != nil {
			name := html.UnescapeString(match[1])
			run.Runners = append(run.Runners, &Participant{"", name, "??", SEX_UNKNOWN, 0, 0, 0, AchievementNone})
			continue
		}

		return fmt.Errorf("cannot parse table row: %s", match0[0])
	}

	var runnerWithTime *Participant = nil
	for _, p := range run.Runners {
		if p.Time != 0 {
			runnerWithTime = p
			break
		}
	}
	if runnerWithTime != nil {
		for _, p := range run.Runners {
			if p.Time != 0 {
				runnerWithTime = p
			} else {
				p.Time = runnerWithTime.Time
			}
		}
	}

	matchesV := patternVolunteerRow.FindAllStringSubmatch(buf, -1)
	for _, match := range matchesV {
		id := match[1]
		name := html.UnescapeString(match[2])

		run.Volunteers = append(run.Volunteers, &Participant{id, name, "??", SEX_UNKNOWN, -1, -1, 0, AchievementNone})
	}

	return nil
}
