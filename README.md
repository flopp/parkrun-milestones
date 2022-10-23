# parkrun-milestones

Try to determine milestone candidates to the next run at a parkrun event.

## Commands

### parkrun-events

You can use this command to search for events (e.g. in order to find out the id of a specific event).

Example:

```
$ ./parkrun-events east 
+--------------------+--------------------------------+----------------+
| EVENT ID           | EVENT NAME                     | COUNTRY        |
+--------------------+--------------------------------+----------------+
| eastbourne         | Eastbourne parkrun             | United Kingdom |
| eastbourne-juniors | Eastbourne junior parkrun      | United Kingdom |
| eastbrighton       | East Brighton parkrun          | United Kingdom |
| eastcoastbrewery   | East Coast Brewery parkrun     | South Africa   |
| eastcoastpark      | East Coast Park parkrun        | Singapore      |
| eastend            | East End parkrun, New Plymouth | New Zealand    |
| easterngardens     | Eastern Gardens parkrun        | Australia      |
| eastgrinstead      | East Grinstead parkrun         | United Kingdom |
| eastleigh          | Eastleigh parkrun              | United Kingdom |
| eastney-juniors    | Eastney junior parkrun         | United Kingdom |
| eastpark           | East Park parkrun              | United Kingdom |
| eastrichmond       | East Richmond parkrun          | Australia      |
| eastville          | Eastville parkrun              | United Kingdom |
| eastville-juniors  | Eastville junior parkrun       | United Kingdom |
| reynellaeast       | Reynella East parkrun          | Australia      |
| somerseteast       | Somerset East parkrun          | South Africa   |
+--------------------+--------------------------------+----------------+
```

### parkrun-milestones

Determine possible milestone candidates for the next run of a given event.
A milestone candidate is a runner or volunteer, who will probably have a milestone number of runs or volunteerings (25, 50, 100, 250, 500) at the upcoming run, and who was active (running or volunteering) in at least 30% (parameter `-active`) the last 10 runs of the event (parameter `-runs`).

Example:

````
$ ./parkrun-milestones eastville
+---------------------+------+------+--------+
| NAME                | RUNS | VOLS | ACTIVE |
+---------------------+------+------+--------+
| Darren CLINTON      | *49* | 6    | 4/10   |
| Elena THODE MINGUET | *99* | 9    | 5/10   |
| Helen SAWYER        | 193  | *49* | 4/10   |
| James HARRISON      | *99* | 134  | 6/10   |
| James RODLIFF       | *99* | 13   | 3/10   |
| Joseph BRAZIER      | *49* | 0    | 7/10   |
| Philip SIM          | *49* | 0    | 4/10   |
| Rosie BURROWS       | *24* | 114  | 5/10   |
+---------------------+------+------+--------+
```