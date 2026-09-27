package main

import (
	"reflect"
	"testing"
	"time"
)

func TestParseNameservers(t *testing.T) {
	got := parseNameservers("1.1.1.1, 9.9.9.9", []string{"8.8.8.8"})
	want := []string{"1.1.1.1", "9.9.9.9", "8.8.8.8"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := parseNameservers("", nil); len(got) != 0 {
		t.Fatalf("expected no nameservers, got %v", got)
	}
}

func TestMeasurement(t *testing.T) {
	nsStore := &nsInfoMap{ns: make(map[string]NInfo)}
	for i := 1; i <= 20; i++ {
		nsStoreSetRTT(nsStore, "ns", time.Duration(i)*time.Millisecond)
	}
	m := nsStoreGetMeasurement(nsStore, "ns")
	if m.rttMin != 1*time.Millisecond || m.rttMax != 20*time.Millisecond {
		t.Errorf("unexpected min/max: %v/%v", m.rttMin, m.rttMax)
	}
	if m.rttMedian != 10*time.Millisecond {
		t.Errorf("unexpected median: %v", m.rttMedian)
	}
	if m.rttP95 != 19*time.Millisecond {
		t.Errorf("unexpected 95th percentile: %v", m.rttP95)
	}
	if m.rttAvg != 10500*time.Microsecond {
		t.Errorf("unexpected average: %v", m.rttAvg)
	}
}

func TestMeasurementWithoutAnswers(t *testing.T) {
	nsStore := &nsInfoMap{ns: make(map[string]NInfo)}
	nsStoreAddConnectionError(nsStore, "ns")
	nsStoreAddValidationError(nsStore, "ns")
	m := nsStoreGetMeasurement(nsStore, "ns")
	if m.rttAvg != 0 {
		t.Errorf("expected zero average, got %v", m.rttAvg)
	}
	if e := nsStore.ns["ns"]; e.ErrorsConnection != 1 || e.ErrorsValidation != 1 || e.Count != 2 {
		t.Errorf("unexpected error counters: %+v", e)
	}
}
