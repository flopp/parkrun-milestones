package parkrun

import (
	"fmt"
	"regexp"
	"strconv"

	download "github.com/flopp/parkrun-milestones/internal/download"
	file "github.com/flopp/parkrun-milestones/internal/file"
)

type Parkrunner struct {
	Id     string
	Name   string
	Runs   int64
	Vols   int64
	Active map[uint64]bool
}

func Milestone(number int64) bool {
	return number == 25 || number == 50 || number == 100 || number == 250 || number == 500
}

func (parkrunner Parkrunner) IsMilstoneCandidate() bool {
	return Milestone(parkrunner.Runs+1) || Milestone(parkrunner.Vols+1)
}

func updateParkrunner(parkrunners map[string]*Parkrunner, id string, name string, runs int64, vols int64, runIndex uint64) (map[string]*Parkrunner, error) {
	if parkrunner, ok := parkrunners[id]; ok {
		parkrunner.Active[runIndex] = true

		if runs > parkrunner.Runs {
			parkrunner.Runs = runs
		}
		if vols > parkrunner.Vols {
			parkrunner.Vols = vols
		}
	} else {
		parkrunners[id] = &Parkrunner{id, name, runs, vols, map[uint64]bool{runIndex: true}}
	}
	return parkrunners, nil
}

func (parkrunner *Parkrunner) fetchMissingStats() error {
	if parkrunner.Runs >= 0 && parkrunner.Vols >= 0 {
		return nil
	}

	url := fmt.Sprintf("https://www.parkrun.org.uk/parkrunner/%s/", parkrunner.Id)
	filePath, err := CachePath("parkrunner/%s", parkrunner.Id)
	if err != nil {
		return err
	}
	if err := download.DownloadFile(url, filePath, MaxFileAge); err != nil {
		return err
	}

	buf, err := file.ReadFile(filePath)
	if err != nil {
		return err
	}

	patternR0 := regexp.MustCompile(`No results have been recorded yet for this parkrunner`)
	patternR := regexp.MustCompile(`<h3>(\d+) parkruns? total</h3>`)
	patternV := regexp.MustCompile(`<strong>Total Credits</strong></td><td><strong>(\d+)</strong>`)

	r := 0
	matchR := patternR.FindStringSubmatch(buf)
	if matchR == nil {
		if !patternR0.MatchString(buf) {
			return fmt.Errorf("cannot find running stats for %s", parkrunner.Id)
		}
	} else {
		r, err = strconv.Atoi(matchR[1])
		if err != nil {
			return err
		}
	}

	v := 0
	matchV := patternV.FindStringSubmatch(buf)
	if matchV == nil {
		return fmt.Errorf("cannot find volunteering stats for %s", parkrunner.Id)
	} else {
		v, err = strconv.Atoi(matchV[1])
		if err != nil {
			return err
		}
	}

	parkrunner.Runs = int64(r)
	parkrunner.Vols = int64(v)

	return nil
}
