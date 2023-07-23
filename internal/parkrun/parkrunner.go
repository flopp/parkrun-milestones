package parkrun

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Parkrunner struct {
	Id         string
	Name       string
	AgeGroup   int
	DataTime   time.Time
	Runs       int64
	JuniorRuns int64
	Vols       int64
	Active     map[uint64]bool
}

func Milestone(number int64) bool {
	return number == 25 || number == 50 || number == 100 || number == 250 || number == 500
}

func updateParkrunner(parkrunners map[string]*Parkrunner, id string, name string, ageGroup int, dataTime time.Time, runs int64, juniorRuns int64, vols int64, runIndex uint64) map[string]*Parkrunner {
	if parkrunner, ok := parkrunners[id]; ok {
		parkrunner.Active[runIndex] = true
		parkrunner.update(dataTime, ageGroup, runs, juniorRuns, vols)
	} else {
		parkrunners[id] = &Parkrunner{id, name, ageGroup, dataTime, runs, juniorRuns, vols, map[uint64]bool{runIndex: true}}
	}
	return parkrunners
}

var (
	patternR0  = regexp.MustCompile(`No results have been recorded yet for this parkrunner`)
	patternR1  = regexp.MustCompile(`<h3>(\d+) parkruns? total</h3>`)
	patternRJ1 = regexp.MustCompile(`<h3>(\d+) junior parkruns? total</h3>`)
	patternR2  = regexp.MustCompile(`<h3>(\d+) parkruns? & (\d+) junior parkruns? total</h3>`)
	patternV   = regexp.MustCompile(`<strong>Total Credits</strong></td><td><strong>(\d+)</strong>`)
)

func (parkrunner *Parkrunner) extractRunCount(buf string) (int, int, error) {
	match := patternR1.FindStringSubmatch(buf)
	if match != nil {
		r, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, 0, err
		}
		return r, 0, nil
	}

	match = patternRJ1.FindStringSubmatch(buf)
	if match != nil {
		j, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, 0, err
		}
		return 0, j, nil
	}

	match = patternR2.FindStringSubmatch(buf)
	if match != nil {
		r, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, 0, err
		}
		j, err := strconv.Atoi(match[2])
		if err != nil {
			return 0, 0, err
		}
		return r, j, nil
	}

	if patternR0.MatchString(buf) {
		return 0, 0, nil
	}

	// only volunteer
	if patternV.MatchString(buf) {
		return 0, 0, nil
	}

	return 0, 0, fmt.Errorf("cannot find running stats for %s", parkrunner.Id)
}

func (parkrunner *Parkrunner) NeedsUpdate() bool {
	// always update old data
	if parkrunner.Id == "" {
		return false
	}
	if parkrunner.DataTime.Add(MaxFileAge).Before(time.Now()) {
		return true
	}
	if parkrunner.Runs >= 0 || parkrunner.JuniorRuns >= 0 || parkrunner.Vols >= 0 {
		return false
	}
	return true
}

func (parkrunner *Parkrunner) update(dataTime time.Time, ageGroup int, runs int64, juniorRuns int64, vols int64) {
	if runs > parkrunner.Runs {
		parkrunner.Runs = runs
	}
	if juniorRuns > parkrunner.JuniorRuns {
		parkrunner.JuniorRuns = juniorRuns
	}
	if vols > parkrunner.Vols {
		parkrunner.Vols = vols
	}
	if ageGroup > parkrunner.AgeGroup {
		parkrunner.AgeGroup = ageGroup
	}
	if dataTime.After(parkrunner.DataTime) {
		parkrunner.DataTime = dataTime
	}
}

func (parkrunner *Parkrunner) FetchMissingStats(lastRunTime time.Time) error {
	if !parkrunner.NeedsUpdate() {
		return nil
	}

	url := fmt.Sprintf("https://www.parkrun.org.uk/parkrunner/%s/", parkrunner.Id)
	fileName := fmt.Sprintf("parkrunner/%s", parkrunner.Id)
	buf, dataTime, err := DownloadAndReadMaxMtime(url, fileName, lastRunTime.Add(24*time.Hour))
	if err != nil {
		return err
	}

	r, j, err := parkrunner.extractRunCount(buf)
	if err != nil {
		return err
	}

	v := 0
	matchV := patternV.FindStringSubmatch(buf)
	if matchV != nil {
		v, err = strconv.Atoi(matchV[1])
		if err != nil {
			return err
		}
	}

	parkrunner.update(dataTime, -2, int64(r), int64(j), int64(v))

	return nil
}

var reParkrunnerEvent = regexp.MustCompile(`<tr><td><a href=".*/(.*)/results">[^<]*</a></td><td>(\d+)</td><td>[^<]*</td><td>[^<]*</td><td><span class="pretty-time">[^<]*</span></td><td><a href="[^"]*"`)

func GetParkrunnerCountry(id string, eventCountries map[string]string) (string, error) {
	url := fmt.Sprintf("https://www.parkrun.org.uk/parkrunner/%s/", id)
	fileName := fmt.Sprintf("parkrunner/%s", id)
	buf, _, err := DownloadAndRead(url, fileName)
	if err != nil {
		return "", err
	}

	counts := make(map[string]int)
	for _, match := range reParkrunnerEvent.FindAllStringSubmatch(buf, -1) {
		eventId := match[1]
		count, err := strconv.Atoi(match[2])
		if err != nil {
			panic(err)
		}
		if country, found := eventCountries[eventId]; found {
			//fmt.Printf("%s => %s %d %s\n", id, eventId, count, country)
			counts[country] += count
		}
	}

	maxCount := -1
	maxCountry := "UNKNOWN"
	for country, count := range counts {
		if count > maxCount {
			maxCount = count
			maxCountry = country
		}
	}

	return maxCountry, nil
}
