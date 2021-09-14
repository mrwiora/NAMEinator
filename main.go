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
)

var VERSION = "custom"
var appConfiguration AppConfig

type AppConfig struct {
	numberOfDomains int
	debug           bool
	contest         bool
	nameserver      string
}

//go:embed datasrc

var datasrc embed.FS

// process flags
func processFlags() {
	var appConfig AppConfig
	flagNumberOfDomains := flag.Int("domains", 100, "number of domains to be tested")
	flagNameserver := flag.String("nameserver", "", "specify a nameserver instead of using defaults")
	flagContest := flag.Bool("contest", true, "contest=true/false : enable or disable a contest against your locally configured DNS server (default true)")
	flagDebug := flag.Bool("debug", false, "debug=true/false : enable or disable debugging (default false)")
	flag.Parse()
	appConfig.numberOfDomains = *flagNumberOfDomains
	appConfig.debug = *flagDebug
	appConfig.contest = *flagContest
	appConfig.nameserver = *flagNameserver
	appConfiguration = appConfig
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
	fmt.Println("NOTE: as this is an alpha - we rely on feedback - please report bugs and feature requests to https://github.com/mwiora/NAMEinator/issues and provide this output")
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
		entry.ID = int64(nsResults.rttAvg)
		nsStore.ns[entry.IPAddr] = entry
		nsStoreSorted = append(nsStoreSorted, NInfo{entry.IPAddr, entry.Name, entry.Country, entry.Count, entry.ErrorsConnection, entry.ErrorsValidation, entry.ID, entry.rtt, entry.rttAvg, entry.rttMin, entry.rttMax})
	}
	sort.Slice(nsStoreSorted, func(i, j int) bool {
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
		fmt.Printf("Avg. [%v], Min. [%v], Max. [%v] ", nameserver.rttAvg, nameserver.rttMin, nameserver.rttMax)
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
	prepareBenchmarkNameservers(nsStore)
	prepareBenchmarkDomains(dStore)
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
			_ = in
			nsStoreSetRTT(nsStore, nameserver.IPAddr, rtt)
			// increment progress bar
			bar.Increment()
			_ = err // TODO: Take care about errors during queries against the DNS - we will accept X fails
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
