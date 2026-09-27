package main

import (
	"sort"
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
	rttMedian        time.Duration
	rttP95           time.Duration
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

// Get nameserver statistics (average, min, max, median and 95th percentile) of successful queries
func nsStoreGetMeasurement(nsStore *nsInfoMap, ipAddr string) NInfo {
	var nsMeasurement = NInfo{}
	entry := nsStore.ns[ipAddr]
	if len(entry.rtt) == 0 {
		return nsMeasurement
	}
	sorted := make([]time.Duration, len(entry.rtt))
	copy(sorted, entry.rtt)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	var total time.Duration = 0
	for _, value := range sorted {
		total += value
	}
	nsMeasurement.rttAvg = total / time.Duration(len(sorted))
	nsMeasurement.rttMin = sorted[0]
	nsMeasurement.rttMax = sorted[len(sorted)-1]
	nsMeasurement.rttMedian = percentile(sorted, 50)
	nsMeasurement.rttP95 = percentile(sorted, 95)
	return nsMeasurement
}

// percentile returns the p-th percentile (nearest-rank method) of an ascending sorted slice
func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := (p*len(sorted) + 99) / 100 // ceil(p/100 * n)
	if rank < 1 {
		rank = 1
	}
	return sorted[rank-1]
}

// count a query that failed because the nameserver could not be reached (timeout, refused connection, ...)
func nsStoreAddConnectionError(nsStore *nsInfoMap, ipAddr string) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry := nsStore.ns[ipAddr]
	entry.IPAddr = ipAddr
	entry.ErrorsConnection++
	entry.Count++
	nsStore.ns[ipAddr] = entry
}

// count a query that was answered with an error code (SERVFAIL, REFUSED, ...)
func nsStoreAddValidationError(nsStore *nsInfoMap, ipAddr string) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry := nsStore.ns[ipAddr]
	entry.IPAddr = ipAddr
	entry.ErrorsValidation++
	entry.Count++
	nsStore.ns[ipAddr] = entry
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
