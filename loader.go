package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	defaultNameserversFile = "datasrc/nameserver-globals.csv"
	defaultDomainsFile     = "datasrc/alexa-top-2000-domains.txt"
)

// openDataFile opens a user supplied file from disk or, if no path is given, the embedded default
func openDataFile(path string, embeddedPath string) (io.ReadCloser, error) {
	if path != "" {
		fmt.Println("trying to load " + path)
		return os.Open(path)
	}
	fmt.Println("trying to load built-in " + embeddedPath)
	return datasrc.Open(embeddedPath)
}

func prepareBenchmarkNameservers(nsStore *nsInfoMap) error {
	if appConfiguration.nameserver != "" {
		loadNameserver(nsStore, appConfiguration.nameserver, "givenByParameter")
		return nil
	}
	// read global nameservers from given file
	file, err := openDataFile(appConfiguration.nameserversFile, defaultNameserversFile)
	if err != nil {
		return fmt.Errorf("unable to open nameserver list: %w", err)
	}
	defer file.Close()
	return readNameservers(nsStore, file)
}

func prepareBenchmarkDomains(dStore *dInfoMap) error {
	// read domains from given file
	file, err := openDataFile(appConfiguration.domainsFile, defaultDomainsFile)
	if err != nil {
		return fmt.Errorf("unable to open domain list: %w", err)
	}
	defer file.Close()
	allDomains, err := readDomains(file)
	if err != nil {
		return fmt.Errorf("unable to read domain list: %w", err)
	}
	if len(allDomains) == 0 {
		return fmt.Errorf("domain list is empty")
	}
	// randomize domains from file to avoid cached results
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allDomains), func(i, j int) { allDomains[i], allDomains[j] = allDomains[j], allDomains[i] })
	// take care only for the domain-tests we were looking for
	numberOfDomains := appConfiguration.numberOfDomains
	if numberOfDomains <= 0 || numberOfDomains > len(allDomains) {
		fmt.Printf("requested %d domains, but %d are available - using %d\n", numberOfDomains, len(allDomains), len(allDomains))
		numberOfDomains = len(allDomains)
	}
	dStoreAddFQDN(dStore, allDomains[:numberOfDomains])
	return nil
}

// load nameservers
func loadNameserver(nsStore *nsInfoMap, ip string, name string) {
	nsStoreAddNS(nsStore, ip, name, "LOCAL")
}

// readNameservers reads "ip,name,country" records; name and country are optional
func readNameservers(nsStore *nsInfoMap, r io.Reader) error {
	nameserverReader := csv.NewReader(bufio.NewReader(r))
	nameserverReader.FieldsPerRecord = -1
	nameserverReader.TrimLeadingSpace = true
	nameserverReader.Comment = '#'
	loaded := 0
	for {
		line, err := nameserverReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to read nameserver list: %w", err)
		}
		fields := make([]string, 3)
		copy(fields, line)
		ip := strings.TrimSpace(fields[0])
		if ip == "" {
			continue
		}
		nsStoreAddNS(nsStore, ip, strings.TrimSpace(fields[1]), strings.TrimSpace(fields[2]))
		loaded++
	}
	if loaded == 0 {
		return fmt.Errorf("nameserver list is empty")
	}
	return nil
}

// readDomains returns the non-empty lines of the given reader
func readDomains(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// the resolver expects fully qualified names, e.g. "example.com."
		lines = append(lines, dns.Fqdn(line))
	}
	return lines, scanner.Err()
}
