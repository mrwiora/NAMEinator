package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

func prepareBenchmarkNameservers(nsStore *nsInfoMap) {
	if appConfiguration.nameserver == "" {
		// read global nameservers from given file
		fmt.Println("trying to load nameservers from nameserver-globals")
		readNameserversFromFile(nsStore, "datasrc/nameserver-globals.csv") // TODO: Split read and Load
	} else {
		loadNameserver(nsStore, appConfiguration.nameserver, "givenByParameter")
	}
}

func prepareBenchmarkDomains(dStore *dInfoMap) {
	var domains []string
	// read domains from given file
	fmt.Println("trying to load domains from cisco-umbrella-top-1m")
	allDomains, err := readLoadDomainsFromFile("datasrc/cisco-umbrella-top-1m.txt")
	if err != nil {
		fmt.Println("File not found")
		return
	}
	_ = err // TODO: Exception handling in case that the files do not exist
	// randomize domains from file to avoid cached results
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allDomains), func(i, j int) { allDomains[i], allDomains[j] = allDomains[j], allDomains[i] })
	// take care only for the domain-tests we were looking for
	domains = allDomains[0:appConfiguration.numberOfDomains]
	dStoreAddFQDN(dStore, domains)
}

// load nameservers
func loadNameserver(nsStore *nsInfoMap, ip string, name string) {
	nsStoreAddNS(nsStore, ip, name, "LOCAL")
}

// load nameservers
func readNameserversFromFile(nsStore *nsInfoMap, filename string) {
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
func readLoadDomainsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
