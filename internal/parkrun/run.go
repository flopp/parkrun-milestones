package parkrun

import (
	"fmt"
	"time"

	"github.com/flopp/go-parkrunparser"
)

type Participant struct {
	Id          string
	Name        string
	AgeGroup    string
	Sex         parkrunparser.Sex
	Runs        int64
	Vols        int64
	Time        time.Duration
	Achievement parkrunparser.Achievement
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

	results, err := parkrunparser.ParseResults([]byte(buf))
	if err != nil {
		return fmt.Errorf("while parsing results for %s from %s: %w", event.Id, fileName, err)
	}

	run.IsComplete = true
	run.DataTime = dataTime
	for _, finisher := range results.Finishers {
		run.Runners = append(run.Runners, &Participant{finisher.Id, finisher.Name, finisher.AgeGroup.Name, finisher.AgeGroup.Sex, int64(finisher.NumberOfRuns), int64(finisher.NumberOfVolunteerings), finisher.Time, finisher.Achievement})
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

	for _, volunteer := range results.Volunteers {
		run.Volunteers = append(run.Volunteers, &Participant{volunteer.Id, volunteer.Name, "??", parkrunparser.SEX_UNKNOWN, -1, -1, 0, parkrunparser.AchievementNone})
	}

	return nil
}
