package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	usage = `USAGE: %s [OPTIONS...] [PATTERN]
List all parkrun events (event id, event name, country) that contain the given pattern (if given).

OPTIONS:
`
)

func main() {
	forceReload := flag.Bool("force", false, "force reload of all data")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) > 1 {
		panic("Too many arguments")
	}

	pattern := ""
	if len(flag.Args()) == 1 {
		pattern = strings.ToLower(flag.Arg(0))
	}

	if *forceReload {
		parkrun.MaxFileAge = 0
	}

	eventList, err := parkrun.AllEvents()
	if err != nil {
		panic(err)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Event Id", "Event Name", "Country"})
	for _, event := range eventList {
		if strings.Contains(strings.ToLower(event.Id), pattern) || strings.Contains(strings.ToLower(event.Name), pattern) {
			t.AppendRow([]interface{}{event.Id, event.Name, event.Country})
		}
	}
	t.Render()
}
