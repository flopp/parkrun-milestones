package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] EVENTID TARGETFILE
Print all time people (runners, volunteers) list

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload bool
	eventId     string
	targetFile  string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		panic("bad command line")
	}

	return CommandLineOptions{
		*forceReload, flag.Args()[0], flag.Args()[1],
	}
}

func getEvent(eventId string) *parkrun.Event {
	event, err := parkrun.LookupEvent(eventId)
	if err != nil {
		panic(err)
	}
	return event
}

type Person struct {
	Id      string
	Name    string
	Last    *parkrun.Run
	Runs    uint64
	Vols    uint64
	Active  uint64
	PB      time.Duration
	RunsAll int64
	VolsAll int64
}

func (person *Person) PBStr() string {
	if person.PB == 0 {
		return "n/a"
	}
	return person.PB.String()
}

func (person *Person) update(run *parkrun.Run, r uint64, v uint64, pb time.Duration, rAll int64, vAll int64) {
	if run != nil {
		person.Last = run
	}
	person.Runs += r
	person.Vols += v
	if r+v > 0 {
		person.Active += 1
	}
	if person.PB == 0 || (pb > 0 && pb < person.PB) {
		person.PB = pb
	}
	if rAll > person.RunsAll {
		person.RunsAll = rAll
	}
	if vAll > person.VolsAll {
		person.VolsAll = vAll
	}
}

func createPerson(i string, n string, run *parkrun.Run, r uint64, v uint64, pb time.Duration, rAll int64, vAll int64) *Person {
	return &Person{i, n, run, r, v, 1, pb, rAll, vAll}
}

func (p *Person) fetchMissingStats() error {
	if p.RunsAll >= 0 || p.VolsAll >= 0 {
		return nil
	}
	url := fmt.Sprintf("https://www.parkrun.org.uk/parkrunner/%s/", p.Id)
	fmt.Printf("Updating %s %s\n", p.Name, url)
	fileName := fmt.Sprintf("parkrunner/%s", p.Id)
	buf, _, err := parkrun.DownloadAndRead(url, fileName)
	if err != nil {
		return err
	}

	r, j, v, err := parkrun.ExtractData(buf)
	if err != nil {
		return err
	}

	p.update(nil, 0, 0, 0, int64(r+j), int64(v))
	return nil
}

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	event := getEvent(options.eventId)
	if err := event.Complete(); err != nil {
		panic(err)
	}

	personsMap := make(map[string]*Person)
	for _, run := range event.Runs {
		if err := run.Complete(); err != nil {
			panic(err)
		}

		type rv struct{ r, v *parkrun.Participant }
		pp := make(map[string]*rv)

		for _, p := range run.Runners {
			if p.Id == "" {
				continue
			}
			pp[p.Id] = &rv{p, nil}
		}
		for _, p := range run.Volunteers {
			if p.Id == "" {
				continue
			}
			ppid, found := pp[p.Id]
			if found {
				ppid.v = p
			} else {
				pp[p.Id] = &rv{nil, p}
			}
		}

		for id, ppid := range pp {
			person, found := personsMap[id]
			if !found {
				if ppid.r != nil {
					name := ppid.r.Name
					time := ppid.r.Time
					r := ppid.r.Runs
					v := ppid.r.Vols
					if ppid.v != nil {
						personsMap[id] = createPerson(id, name, run, 1, 1, time, r, v)
					} else {
						personsMap[id] = createPerson(id, name, run, 1, 0, time, r, v)
					}
				} else {
					name := ppid.v.Name
					time := ppid.v.Time
					r := ppid.v.Runs
					v := ppid.v.Vols
					personsMap[id] = createPerson(id, name, run, 0, 1, time, r, v)
				}
			} else {
				if ppid.r != nil {
					time := ppid.r.Time
					r := ppid.r.Runs
					v := ppid.r.Vols
					if ppid.v != nil {
						person.update(run, 1, 1, time, r, v)
					} else {
						person.update(run, 1, 0, time, r, v)
					}
				} else {
					time := ppid.v.Time
					r := ppid.v.Runs
					v := ppid.v.Vols
					person.update(run, 0, 1, time, r, v)
				}
			}
		}
	}

	persons := make([]*Person, 0, len(personsMap))
	for _, p := range personsMap {
		err := p.fetchMissingStats()
		if err != nil {
			panic(err)
		}
		persons = append(persons, p)
	}
	sort.Slice(persons, func(i, j int) bool {
		pi := persons[i]
		pj := persons[j]
		ni := pi.Active
		nj := pj.Active
		if ni != nj {
			return ni >= nj
		}
		return pi.Name <= pj.Name
	})

	t, err := template.ParseFiles("cmd/people/template.html")
	if err != nil {
		panic(err)
	}

	out, err := os.Create(options.targetFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = t.Execute(out, persons)
	if err != nil {
		panic(err)
	}
}
