# rsummary

## Purpose
CLI Utility to produce a report of all hosts that have written to rsyslog log files.

## Features
1. Specify an rsyslog directory
2. Specify number of threads. Can process multiple servers at the same time
3. Figures out the lengths of hostnames and numbers to make the report look nice
4. Calculates the top talkers and comes up with an overall percentage for each host, keeps track of counts
5. Attempts an IP lookup

## Usage
```
Usage: ./program_name --dir=<log directory path> --T=<number of threads>
Or: ./program_name -dir=<log directory path> -T=<number of threads>
```

## Output
```
root@rsyslogserver1:~/go# go run rsummary.go --dir=logs
Processing: /syslog/messages.log (0.00% completed)

SUMMARY:
Hostname                                 Count    Percentage  IP Address
------------------------------------------------------------------------
servera                                  465937   81.82%    192.168.2.1
serverb                                  45714    8.03%     192.168.2.2
serverc                                  21034    3.69%     192.168.2.3
serverd                                  18194    3.19%     192.168.2.4
servere                                  18194    3.19%     192.168.2.5
serverf                                  183      0.03%     N/A
```
