package main

import (
	"strings"
	"testing"
)

func TestReadNameserversToleratesShortAndCommentLines(t *testing.T) {
	nsStore := &nsInfoMap{ns: make(map[string]NInfo)}
	input := "8.8.8.8,google,US\n# comment\n\n1.1.1.1\n 9.9.9.9 , quad9\n"
	if err := readNameservers(nsStore, strings.NewReader(input)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nsStore.ns) != 3 {
		t.Fatalf("expected 3 nameservers, got %d: %v", len(nsStore.ns), nsStore.ns)
	}
	if nsStore.ns["9.9.9.9"].Name != "quad9" {
		t.Errorf("expected name quad9, got %q", nsStore.ns["9.9.9.9"].Name)
	}
}

func TestReadNameserversEmpty(t *testing.T) {
	nsStore := &nsInfoMap{ns: make(map[string]NInfo)}
	if err := readNameservers(nsStore, strings.NewReader("\n# nothing\n")); err == nil {
		t.Fatal("expected an error for an empty nameserver list")
	}
}

func TestReadDomainsMakesNamesFullyQualified(t *testing.T) {
	domains, err := readDomains(strings.NewReader("example.com\n\nexample.org.\n# skipped\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 || domains[0] != "example.com." || domains[1] != "example.org." {
		t.Fatalf("unexpected domains: %v", domains)
	}
}

func TestPrepareBenchmarkDomainsClampsToAvailable(t *testing.T) {
	appConfiguration = AppConfig{numberOfDomains: 1000000}
	dStore := &dInfoMap{d: make(map[string]DInfo)}
	if err := prepareBenchmarkDomains(dStore); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dStore.d) == 0 || len(dStore.d) >= 1000000 {
		t.Fatalf("expected the built-in domain list to be used in full, got %d", len(dStore.d))
	}
}

func TestEmbeddedNameserversLoad(t *testing.T) {
	appConfiguration = AppConfig{}
	nsStore := &nsInfoMap{ns: make(map[string]NInfo)}
	if err := prepareBenchmarkNameservers(nsStore); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nsStore.ns) == 0 {
		t.Fatal("expected nameservers from the embedded list")
	}
}
