package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

// TestGetIpsFromEndpoint verifies IP extraction from endpoints.
func TestGetIpsFromEndpoint(t *testing.T) {
	// Build a sample endpoints list.
	endpoints := &v1.EndpointsList{
		Items: []v1.Endpoints{
			{
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{IP: "10.0.0.1"},
							{IP: "10.0.0.2"},
						},
					},
				},
			},
		},
	}

	// Extract IPs from the endpoints.
	ips, err := getIpsFromEndpoint(endpoints)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ensure IPs were returned.
	if len(ips) < 1 {
		t.Fatalf("no ips found from endpoint list")
	}
}
