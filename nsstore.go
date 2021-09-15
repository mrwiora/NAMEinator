package main

import (
	"sync"
	"time"
)

type NInfo struct {
	IPAddr           string
	Name             string
	Country          string
	Count            int
	ErrorsConnection int
	ErrorsValidation int
	ID               int64
	rtt              []time.Duration
	rttAvg           time.Duration
	rttMin           time.Duration
	rttMax           time.Duration
}

type nsInfoMap struct {
	ns    map[string]NInfo
	mutex sync.RWMutex
}

// Get IP address entry // DEBUG
func nsStoreGetRecord(nsStore *nsInfoMap, ipAddr string) NInfo {
	nsStore.mutex.RLock()
	defer nsStore.mutex.RUnlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	return entry
}

// Get nameserver average time
func nsStoreGetMeasurement(nsStore *nsInfoMap, ipAddr string) NInfo {
	var nsMeasurement = NInfo{}
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	var total time.Duration = 0
	var min time.Duration = 10000000
	var max time.Duration = 0
	for _, value := range entry.rtt {
		// check for new min record
		if value < min {
			min = value
		}
		// check for new max record
		if value > max {
			max = value
		}
		// add for total time
		total += value
	}
	nsMeasurement.rttAvg = total / time.Duration(len(entry.rtt))
	nsMeasurement.rttMin = min
	nsMeasurement.rttMax = max
	return nsMeasurement
}

// add rtt to the nameserver slice
func nsStoreSetRTT(nsStore *nsInfoMap, ipAddr string, rtt time.Duration) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	entry.rtt = append(entry.rtt, rtt)
	entry.Count++
	nsStore.ns[ipAddr] = entry
}

// add rtt to the nameserver slice
func nsStoreAddNS(nsStore *nsInfoMap, ipAddr string, name string, country string) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	entry.Name = name
	entry.Country = country
	nsStore.ns[ipAddr] = entry
}
