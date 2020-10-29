package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/user"
	"time"
)

func prepareBenchmarkNameservers(nsStore nsInfoMap) {
	if appConfiguration.nameserver == "" {
		// read global nameservers from given file
		fmt.Println("trying to load nameservers from datasrc/nameserver-globals.csv")
		// @ TODO refactor - there shouldn't be the same code in if and else
		//			(but atm this is the only safe way)...
		if _, err := os.Stat("datasrc/nameserver-globals.csv"); err == nil {
			readNameserversFromFile(nsStore, "datasrc/nameserver-globals.csv") // TODO: Split read and Load
		} else if usr, err := user.Current(); err == nil{
			readNameserversFromFile(nsStore, usr.HomeDir + "/.config/nameinator/nameserver-globals.csv") // TODO: Split read and Load
		} else {
			readNameserversFromFile(nsStore, "datasrc/nameserver-globals.csv") // TODO: Split read and Load
		}
	} else {
		loadNameserver(nsStore, appConfiguration.nameserver, "givenByParameter")
	}
}

func prepareBenchmarkDomains(dStore dInfoMap) {
	var domains []string
	// read domains from given file
	fmt.Println("trying to load domains from datasrc/alexa-top-2000-domains.txt")
	var err error
	var alldomains []string
	var usr *user.User
	// @ TODO refactor - there shouldn't be the same code in if and else
	//			(but atm this is the only safe way)...
	if _, err = os.Stat("datasrc/alexa-top-2000-domains.txt"); err == nil {
		alldomains, err = readloadDomainsFromFile("datasrc/alexa-top-2000-domains.txt")
	} else if usr, err = user.Current(); err == nil{
		alldomains, err = readloadDomainsFromFile(usr.HomeDir + "/.config/nameinator/domains.txt")
	} else {
		alldomains, err = readloadDomainsFromFile("datasrc/alexa-top-2000-domains.txt")
	}
	_ = err // TODO: Exception handling in case that the files do not exist
	// randomize domains from file to avoid cached results
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(alldomains), func(i, j int) { alldomains[i], alldomains[j] = alldomains[j], alldomains[i] })
	// take care only for the domain-tests we were looking for
	domains = alldomains[0:appConfiguration.numberOfDomains]
	dStoreAddFQDN(dStore, domains)
}

// load nameservers
func loadNameserver(nsStore nsInfoMap, ip string, name string) {
	nsStoreAddNS(nsStore, ip, name, "LOCAL")
}

// load nameservers
func readNameserversFromFile(nsStore nsInfoMap, filename string) {
	csvFile, _ := os.Open(filename)
	nameserverReader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := nameserverReader.Read()
		if err == io.EOF {
			break
		}
		// fmt.Println(line)
		nsStoreAddNS(nsStore, line[0], line[1], line[2])
		_ = err
	}
}

// readDomainsFromFile reads a whole file into memory
// and returns a slice of its lines.
func readloadDomainsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
