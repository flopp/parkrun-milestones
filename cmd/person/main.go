package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
)

const (
	usage = `USAGE: %s [OPTIONS...] PARKRUNNER_ID
Fetch and print data for the specified parkrunner

OPTIONS:
`
)

type CommandLineOptions struct {
	forceReload  bool
	parkrunnerId string
}

func parseCommandLine() CommandLineOptions {
	forceReload := flag.Bool("force", false, "force reload of all data")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) != 1 {
		panic("bad command line")
	}

	return CommandLineOptions{
		*forceReload, flag.Args()[0],
	}
}

func main() {
	options := parseCommandLine()

	if options.forceReload {
		parkrun.MaxFileAge = 0
	}

	now := time.Now()

	parkrunner := parkrun.Parkrunner{Id: options.parkrunnerId, Runs: -1, Vols: -1, JuniorRuns: -1}
	if err := parkrunner.FetchMissingStats(now); err != nil {
		panic(err)
	}

	fmt.Printf("NAME        = %s\nID          = %s\nRUNS        = %d\nJUNIOR RUNS = %d\nVOLS        = %d\n", parkrunner.Name, parkrunner.Id, parkrunner.Runs, parkrunner.JuniorRuns, parkrunner.Vols)
}
