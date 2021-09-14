package main

import (
	"sync"
)

type DInfo struct {
	FQDN             string
	ErrorsResolution int
}

type dInfoMap struct {
	d     map[string]DInfo
	mutex sync.RWMutex
}

// add rtt to the nameserver slice
func dStoreAddFQDN(dStore *dInfoMap, dns []string) {
	dStore.mutex.Lock()
	defer dStore.mutex.Unlock()
	for _, domain := range dns {
		entry := dStore.d[domain]
		entry.FQDN = domain
		dStore.d[domain] = entry
	}
}
