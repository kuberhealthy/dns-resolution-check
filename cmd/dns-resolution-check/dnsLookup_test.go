package main

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// TestDnsLookup verifies DNS lookup behavior for valid and invalid hosts.
func TestDnsLookup(t *testing.T) {
	// Build a resolver that uses a public DNS server.
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Millisecond * time.Duration(10000)}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	// Define expected outcomes.
	testCase := make(map[string]error)
	testCase["bad.host.com"] = errors.New("DNS Status check determined that bad.host.com is DOWN")
	testCase["google.com"] = nil

	// Iterate each case.
	for host, expectedValue := range testCase {
		err := dnsLookup(resolver, host)
		if err == nil {
			if host != "google.com" {
				t.Fatalf("expected failure for host %s", host)
			}
			continue
		}

		if expectedValue == nil {
			t.Fatalf("expected success for host %s, got error %s", host, err.Error())
		}

		if !strings.Contains(err.Error(), expectedValue.Error()) {
			t.Fatalf("expected error to contain %s, got %s", expectedValue.Error(), err.Error())
		}
	}
}
