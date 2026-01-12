package main

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

// TestCreateResolver verifies resolver creation.
func TestCreateResolver(t *testing.T) {
	// Define the test IP.
	ip := "8.8.8.8"

	// Build the expected resolver type.
	expected := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Millisecond * time.Duration(10000)}
			return d.DialContext(ctx, "udp", ip+":53")
		},
	}

	// Create the resolver.
	resolver, err := createResolver(ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Compare resolver types.
	if reflect.TypeOf(*resolver) != reflect.TypeOf(*expected) {
		t.Fatalf("expected resolver type %T got %T", expected, resolver)
	}
}
