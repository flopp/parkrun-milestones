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
	Id         string
	Name       string
	LastActive time.Time
	Runs       uint64
	Vols       uint64
	PB         time.Duration
}

func (person *Person) PBStr() string {
	if person.PB == 0 {
		return "n/a"
	}
	return person.PB.String()
}

func (person *Person) update(t time.Time, r uint64, v uint64, pb time.Duration) {
	if t.After(person.LastActive) {
		person.LastActive = t
	}
	person.Runs += r
	person.Vols += v
	if person.PB == 0 || (pb > 0 && pb < person.PB) {
		person.PB = pb
	}
}

func createPerson(i string, n string, t time.Time, r uint64, v uint64, pb time.Duration) *Person {
	return &Person{i, n, t, r, v, pb}
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

		for _, p := range run.Runners {
			if p.Id == "" {
				continue
			}
			person, found := personsMap[p.Id]
			if !found {
				personsMap[p.Id] = createPerson(p.Id, p.Name, run.Time, 1, 0, p.Time)
			} else {
				person.update(run.Time, 1, 0, p.Time)
			}
		}

		for _, p := range run.Volunteers {
			if p.Id == "" {
				continue
			}
			person, found := personsMap[p.Id]
			if !found {
				personsMap[p.Id] = createPerson(p.Id, p.Name, run.Time, 0, 1, 0)
			} else {
				person.update(run.Time, 0, 1, 0)
			}
		}
	}

	persons := make([]*Person, 0, len(personsMap))
	for _, p := range personsMap {
		persons = append(persons, p)
	}
	sort.Slice(persons, func(i, j int) bool {
		pi := persons[i]
		pj := persons[j]
		ni := pi.Runs + pi.Vols
		nj := pj.Runs + pj.Vols
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
