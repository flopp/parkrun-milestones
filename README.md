[![PkgGoDev](https://pkg.go.dev/badge/github.com/flopp/go-staticmaps)](https://pkg.go.dev/github.com/flopp/parkrun-milestones)
[![Go Report Card](https://goreportcard.com/badge/github.com/flopp/parkrun-milestones)](https://goreportcard.com/report/flopp/parkrun-milestones)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/flopp/parkrun-milestones/)

# parkrun-milestones

Try to determine milestone candidates to the next run at a parkrun event.

## Commands

### parkrun-events

You can use this command to search for events (e.g. in order to find out the id of a specific event).

#### Parameters

| Parameter | Description |
| --- | --- |
| `PATTERN` | Optional positional search text matched against event IDs and event names, case-insensitively. If omitted, all events are listed. |
| `-force` | Reload all event data instead of using cached data (default: `false`). |

Example:

```
$ ./parkrun-events east 
┌────────────────────┬────────────────────────────────┬────────────────┐
│ EVENT ID           │ EVENT NAME                     │ COUNTRY        │
├────────────────────┼────────────────────────────────┼────────────────┤
│ eastbourne         │ Eastbourne parkrun             │ United Kingdom │
│ eastbourne-juniors │ Eastbourne junior parkrun      │ United Kingdom │
│ eastbrighton       │ East Brighton parkrun          │ United Kingdom │
│ eastcoastbrewery   │ East Coast Brewery parkrun     │ South Africa   │
│ eastcoastpark      │ East Coast Park parkrun        │ Singapore      │
│ eastend            │ East End parkrun, New Plymouth │ New Zealand    │
│ easterngardens     │ Eastern Gardens parkrun        │ Australia      │
│ eastgrinstead      │ East Grinstead parkrun         │ United Kingdom │
│ eastleigh          │ Eastleigh parkrun              │ United Kingdom │
│ eastney-juniors    │ Eastney junior parkrun         │ United Kingdom │
│ eastpark           │ East Park parkrun              │ United Kingdom │
│ eastrichmond       │ East Richmond parkrun          │ Australia      │
│ eastville          │ Eastville parkrun              │ United Kingdom │
│ eastville-juniors  │ Eastville junior parkrun       │ United Kingdom │
│ reynellaeast       │ Reynella East parkrun          │ Australia      │
│ somerseteast       │ Somerset East parkrun          │ South Africa   │
└────────────────────┴────────────────────────────────┴────────────────┘
```

### parkrun-milestones

Determine possible milestone candidates for the next run of a given event.
A milestone candidate is a runner or volunteer, who will probably have a milestone number of runs or volunteerings (25, 50, 100, 250, 500) at the upcoming run, and who was active (running or volunteering) in at least 30% (parameter `-active`) the last 10 runs of the event (parameter `-runs`).

#### Parameters

| Parameter | Description |
| --- | --- |
| `EVENTID...` | One or more parkrun event IDs to examine. Cannot be combined with `-country`. |
| `-force` | Reload all data instead of using cached data (default: `false`). |
| `-active` | Minimum fraction of the selected recent runs in which a person must participate or volunteer. Accepts a value from `0` to `1` (default: `0.3`, or 30%). |
| `-runs` | Maximum number of the latest event runs to examine (default: `10`). |
| `-country` | Examine every event in the specified country instead of accepting `EVENTID...`. The country name is matched without case or surrounding whitespace. |

Example:

```
$ ./parkrun-milestones eastville
┌───────────────────────────────────────────────────────┐
│ Expected Milestones at                                │
│ Eastville parkrun                                     │
│ Run #178                                              │
├────────────────────────────────┬──────┬──────┬────────┤
│ NAME                           │ RUNS │ VOLS │ ACTIVE │
├────────────────────────────────┼──────┼──────┼────────┤
│ Darren CLINTON                 │  *49 │    6 │ 4/10   │
│ Elena THODE MINGUET            │  *99 │    9 │ 5/10   │
│ Helen SAWYER                   │  193 │  *49 │ 4/10   │
│ James HARRISON                 │  *99 │  134 │ 6/10   │
│ James RODLIFF                  │  *99 │   13 │ 3/10   │
│ Joseph BRAZIER                 │  *49 │    0 │ 7/10   │
│ Philip SIM                     │  *49 │    0 │ 4/10   │
│ Rosie BURROWS                  │  *24 │  114 │ 5/10   │
└────────────────────────────────┴──────┴──────┴────────┘
```

### parkrun-runstats
Prints the stats of the latest run in list format; suitable for sharing in text-based social media (mastodon, twitter, etc.).

#### Parameters

| Parameter | Description |
| --- | --- |
| `EVENTID...` | One or more parkrun event IDs to report. Cannot be combined with `-country`. |
| `-force` | Reload all data instead of using cached data (default: `false`). |
| `-fancy` | Use fancy social-media formatting with emoji (default: `false`). |
| `-table` | Print runner and volunteer details as semicolon-separated CSV-style text (default: `false`). This format takes precedence over `-fancy`. |
| `-country` | Report every event in the specified country instead of accepting `EVENTID...`. The country name is matched without case or surrounding whitespace. |

Example:

```
$ ./parkrun-runstats -fancy bushy
Bushy parkrun
#️⃣ 902
📅 2022-11-12
🏃 Runners: 1154
  ⏱ new PB: 108
  🧳 first visitors: 111
  ⭐️ new parkrunners: 37
  🏆 25. run anniversary: 7
  🏆 50. run anniversary: 6
  🏆 100. run anniversary: 3
🦺 Volunteers: 69
  ⭐️ new volunteers: 6
  🏆 25. vol. anniversary: 1
  🏆 100. vol. anniversary: 1
👀 https://www.parkrun.org.uk/bushy/results/902/
```

### parkrun-webgen

Generate an HTML page for one or more events, or for all events in a country.

#### Parameters

| Parameter | Description |
| --- | --- |
| `EVENTID...` | One or more parkrun event IDs to generate pages for. Cannot be combined with `-country`. |
| `-force` | Reload all data instead of using cached data (default: `false`). |
| `-outdir` | Directory in which generated HTML files are written (default: `html`). The directory is created if necessary. |
| `-country` | Generate pages for every event in the specified country instead of accepting `EVENTID...`. The country name is matched without case or surrounding whitespace. |

### parkrun-year

Show runner, volunteer, guest, sex, age-group, run-time, and name statistics for an event, optionally restricted to a year.

#### Parameters

| Parameter | Description |
| --- | --- |
| `EVENTID` | Required parkrun event ID to analyse. |
| `YEAR` | Optional four-digit year to restrict the statistics. If omitted, all available runs are included. |
| `-force` | Reload all data instead of using cached data (default: `false`). |
| `-guestcountries` | Determine the country of guest parkrunners. This may take some time (default: `false`). |

### parkrun-people

Generate an HTML list containing all runners and volunteers recorded for an event.

#### Parameters

| Parameter | Description |
| --- | --- |
| `EVENTID` | Required parkrun event ID to analyse. |
| `TARGETFILE` | Required output HTML file path. |
| `-force` | Reload all data instead of using cached data (default: `false`). |

### parkrun-person

Fetch and print the name and total run, junior-run, and volunteer counts for a parkrunner.

#### Parameters

| Parameter | Description |
| --- | --- |
| `PARKRUNNER_ID` | Required numeric parkrunner ID. |
| `-force` | Reload all parkrunner data instead of using cached data (default: `false`). |