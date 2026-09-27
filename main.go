package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/miekg/dns"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

var VERSION = "custom"
var appConfiguration AppConfig

type AppConfig struct {
	numberOfDomains int
	debug           bool
	contest         bool
	nameserver      []string
	nameserversFile string
	domainsFile     string
}

// the default data files are embedded so the binary works from any directory
//
//go:embed datasrc
var datasrc embed.FS

// process flags
func processFlags() {
	var appConfig AppConfig
	flagNumberOfDomains := flag.Int("domains", 100, "number of domains to be tested")
	flagNameserver := flag.String("nameserver", "", "specify one or more nameservers (comma separated, e.g. 1.1.1.1,9.9.9.9) instead of using defaults")
	flagNameserversFile := flag.String("nameservers-file", "", "path to a CSV file (ip,name,country) with nameservers to test (default: built-in list)")
	flagDomainsFile := flag.String("domains-file", "", "path to a text file with one domain per line (default: built-in list)")
	flagContest := flag.Bool("contest", true, "contest=true/false : enable or disable a contest against your locally configured DNS server (default true)")
	flagDebug := flag.Bool("debug", false, "debug=true/false : enable or disable debugging (default false)")
	flag.Parse()
	appConfig.numberOfDomains = *flagNumberOfDomains
	appConfig.debug = *flagDebug
	appConfig.contest = *flagContest
	appConfig.nameserver = parseNameservers(*flagNameserver, flag.Args())
	appConfig.nameserversFile = *flagNameserversFile
	appConfig.domainsFile = *flagDomainsFile
	appConfiguration = appConfig
}

// parseNameservers splits the -nameserver value on commas/whitespace; remaining
// positional arguments are accepted as well, e.g. "-nameserver 1.1.1.1 9.9.9.9"
func parseNameservers(value string, args []string) []string {
	return strings.FieldsFunc(strings.Join(append([]string{value}, args...), " "), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
}

// return the IP of the DNS used by the operating system
func getOSdns() string {
	// get local dns ip
	out, err := exec.Command("nslookup", ".").Output()
	if appConfiguration.debug {
		fmt.Println("DEBUG: nslookup output")
		fmt.Printf("%s\n", out)
	}
	var errorCode = fmt.Sprint(err)
	if err != nil {
		if errorCode == "exit status 1" {
			// newer versions of nslookup return error code 1 when executing "nslookup ." - but that's fine for us
			_ = err
		} else {
			log.Print("Something went wrong obtaining the local DNS Server - is \"nslookup\" available?")
			log.Fatal(err)
		}
	}

	// fmt.Printf("%s\n", out)
	re := regexp.MustCompile("([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})|(([a-f0-9:]+:+)+[a-f0-9]+)")
	// fmt.Printf("%q\n", re.FindString(string(out)))
	var localDNS = re.FindString(string(out))
	if appConfiguration.debug {
		fmt.Println("DEBUG: dns server")
		fmt.Printf("%s\n", localDNS)
	}
	return localDNS
}

// prints welcome messages
func printWelcome() {
	fmt.Println("starting NAMEinator - version " + VERSION)
	fmt.Printf("understood the following configuration: %+v\n", appConfiguration)
	fmt.Println("-------------")
	fmt.Println("NOTE: as this is an alpha - we rely on feedback - please report bugs and feature requests to https://github.com/mrwiora/NAMEinator/issues and provide this output")
	fmt.Println("OS: " + runtime.GOOS + " ARCH: " + runtime.GOARCH)
	fmt.Println("-------------")
}

func processResults(nsStore *nsInfoMap) []NInfo {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	var nsStoreSorted []NInfo
	for _, entry := range nsStore.ns {
		nsResults := nsStoreGetMeasurement(nsStore, entry.IPAddr)
		entry.rttAvg = nsResults.rttAvg
		entry.rttMin = nsResults.rttMin
		entry.rttMax = nsResults.rttMax
		entry.rttMedian = nsResults.rttMedian
		entry.rttP95 = nsResults.rttP95
		entry.ID = int64(nsResults.rttAvg)
		nsStore.ns[entry.IPAddr] = entry
		nsStoreSorted = append(nsStoreSorted, entry)
	}
	// nameservers without a single successful answer go last
	sort.Slice(nsStoreSorted, func(i, j int) bool {
		iOK, jOK := len(nsStoreSorted[i].rtt) > 0, len(nsStoreSorted[j].rtt) > 0
		if iOK != jOK {
			return iOK
		}
		return nsStoreSorted[i].ID < nsStoreSorted[j].ID
	})
	return nsStoreSorted
}

// prints results
func printResults(nsStore *nsInfoMap, nsStoreSorted []NInfo) {
	fmt.Println("")
	fmt.Println("finished - presenting results: ") // TODO: Colorful representation in a table PLEASE

	for _, nameserver := range nsStoreSorted {
		fmt.Println("")
		fmt.Println(nameserver.IPAddr + ": ")
		if len(nameserver.rtt) == 0 {
			fmt.Print("WARNING: no successful answers - this nameserver is unreachable or failing ")
		} else {
			fmt.Printf("Avg. [%v], Median [%v], 95th perc. [%v], Min. [%v], Max. [%v] ", nameserver.rttAvg, nameserver.rttMedian, nameserver.rttP95, nameserver.rttMin, nameserver.rttMax)
		}
		if errors := nameserver.ErrorsConnection + nameserver.ErrorsValidation; errors > 0 {
			fmt.Printf("\nErrors: %d of %d queries failed (%d unreachable/timeout, %d error responses e.g. SERVFAIL) ", errors, nameserver.Count, nameserver.ErrorsConnection, nameserver.ErrorsValidation)
		}
		if appConfiguration.debug {
			fmt.Println(nsStoreGetRecord(nsStore, nameserver.IPAddr))
		}
		fmt.Println("")
	}
}

// prints bye messages
func printBye() {
	fmt.Println("")
	fmt.Println("Au revoir!")
}

func prepareBenchmark(nsStore *nsInfoMap, dStore *dInfoMap) {
	if appConfiguration.contest {
		// we need to know who we are testing
		var localDNS = getOSdns()
		loadNameserver(nsStore, localDNS, "localhost")
	}
	if err := prepareBenchmarkNameservers(nsStore); err != nil {
		log.Fatal(err)
	}
	if err := prepareBenchmarkDomains(dStore); err != nil {
		log.Fatal(err)
	}
}

func performBenchmark(nsStore *nsInfoMap, dStore *dInfoMap) {
	// create progress bar
	bar := pb.Full.Start(len(nsStore.ns) * len(dStore.d))
	// initialize DNS client
	c := new(dns.Client)
	// to avoid overload against one server we will test all defined nameservers with one domain before proceeding
	for _, domain := range dStore.d {

		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{Name: domain.FQDN, Qtype: dns.TypeA, Qclass: dns.ClassINET}

		// iterate through all given nameservers
		for _, nameserver := range nsStore.ns {
			in, rtt, err := c.Exchange(m1, "["+nameserver.IPAddr+"]"+":53")
			switch {
			case err != nil:
				nsStoreAddConnectionError(nsStore, nameserver.IPAddr)
				if appConfiguration.debug {
					log.Printf("DEBUG: %s query for %s failed: %v", nameserver.IPAddr, domain.FQDN, err)
				}
			case in.Rcode != dns.RcodeSuccess && in.Rcode != dns.RcodeNameError:
				// NXDOMAIN is a valid answer, SERVFAIL/REFUSED/... are not
				nsStoreAddValidationError(nsStore, nameserver.IPAddr)
				if appConfiguration.debug {
					log.Printf("DEBUG: %s answered %s for %s", nameserver.IPAddr, dns.RcodeToString[in.Rcode], domain.FQDN)
				}
			default:
				nsStoreSetRTT(nsStore, nameserver.IPAddr, rtt)
			}
			// increment progress bar
			bar.Increment()
		}
		//fmt.Print(".")
	}
	bar.Finish()
}

func main() {
	// process startup parameters and welcome
	processFlags()
	printWelcome()

	// prepare storage for nameservers and domains
	var nsStore = &nsInfoMap{ns: make(map[string]NInfo)}
	var dStore = &dInfoMap{d: make(map[string]DInfo)}
	// var nsStoreSorted []NInfo

	// based on startup configuration we have to do some preparation
	prepareBenchmark(nsStore, dStore)

	// let's go benchmark - iterate through all domains
	fmt.Println("LETS GO")

	performBenchmark(nsStore, dStore)

	// benchmark has been completed - now we have to tell the results and say good bye
	var nsStoreSorted = processResults(nsStore)
	printResults(nsStore, nsStoreSorted)
	printBye()
}
