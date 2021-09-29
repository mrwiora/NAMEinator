NAMEinator [![Build Status](https://travis-ci.org/mwiora/NAMEinator.svg)](https://travis-ci.org/mwiora/NAMEinator) [![Code Climate](https://codeclimate.com/github/mwiora/NAMEinator/badges/gpa.svg)](https://codeclimate.com/github/mwiora/NAMEinator)
=========

Are you a power-user with 5 minutes to spare? Do you want a faster internet experience?

Try out NAMEinator. It hunts down the fastest DNS servers available for your
computer to use. NAMEinator runs a fair and thorough benchmark using standardized
datasets in order to provide an individualized recommendation. NAMEinator is completely
free and does not modify your system in any way.
This project began as a 20% project at Google.

NAMEinator runs on Windows (10), Linux (tested on Ubuntu 16.04) and is available with a
a command-line interface and in the near future with a graphical user interface.

This project has been forked from google/namebench. While it seems that the initial Author wanted to move the application towards GO with another application use - I just wanted to continue and maintain the python variant as it did what it should.
After investigation the python code showed up some very frustrating complications, which were leading to my decision to reimplement the initial application idea - also in GO.

requirements and getting started
---------------

prerequisite: the system running this software must have the tool "nslookup" installed.
windows users are fine :)

* download and run the corresponding compiled files from releases. execute them on the desired platform

or

* compile yourself (requirement: install go (minimum 1.10.x) as described here https://golang.org/doc/install)
```
go get github.com/mrwiora/NAMEinator
cd $GOPATH/src/github.com/mrwiora/NAMEinator/
go build
./NAMEinator
```

help & start parameters
---------------

```
$ ./NAMEinator.exe -h
Usage of C:\Users\Administrator\go\src\NAMEinator\NAMEinator.exe:
  -contest=false
        enable or disable a contest against your locally configured DNS server (default true)
  -debug=true
        enable or disable debugging (default false)
  -domains int
        number of domains to be tested (default 100)
  -nameserver string
        specify a nameserver instead of using defaults
```

sample output of current version (debug off)
---------------

```
$ ./NAMEinator.exe
starting NAMEinator - version 0.2 alpha
understood the following configuration: {numberOfDomains:100 debug:false contest:true nameserver:}
-------------
NOTE: as this is an alpha - we rely on feedback - please report bugs and featurerequests to https://github.com/mrwiora/NAMEinator/issues and provide this output
OS: windows ARCH: amd64
-------------
trying to load nameservers from datasrc/nameserver-globals.csv
trying to load domains from datasrc/alexa-top-2000-domains.txt
LETS GO - each dot is a completed domain request against all nameservers
....................................................................................................
finished - presenting results:

172.31.0.2:
Avg. [60.109104ms], Min. [313.4µs], Max. [899.4558ms]

8.8.8.8:
Avg. [46.839882ms], Min. [266.8µs], Max. [1.3698924s]

8.8.4.4:
Avg. [38.543425ms], Min. [409.2µs], Max. [555.3005ms]

208.67.222.222:
Avg. [48.621088ms], Min. [217.6µs], Max. [614.7506ms]

2001:470:20::2:
Avg. [82.29074ms], Min. [280.1µs], Max. [1.7591384s]

156.154.71.1:
Avg. [56.059446ms], Min. [269.4µs], Max. [1.2002193s]

216.146.35.35:
Avg. [93.600686ms], Min. [441.6µs], Max. [664.3157ms]

Au revoir!
```

checklist
---------------
basics
- [x] perform DNS Requests
- [x] iterate through given nameservers (basic set) and measure time
- [x] randomly select domain names from alexa top 2000 list
- [x] produce cli report
- [x] test on windows and linux
- [x] release cli version

nice2have
- [ ] implement test driven development

to subsitute namebench 1.3.1
- [ ] support localization of execution
- [ ] select the best suitable DNS server
- [ ] provide basic GUI which has the CLI version as its base
- [ ] produce html/pdf report
- [ ] test on windows and linux
- [ ] release gui version

reimplement functions that were planned, but did not work with namebench 1.3.1
- [ ] detect censorship and manipulated dns entries
- [ ] optional upload of results

new features
- [ ] perform identification of best usable dns server not only based on location, but based also on network path traces
- [ ] test caching - disable cache if this option is selected and ask for the domains a second time (increasing number of domains)
