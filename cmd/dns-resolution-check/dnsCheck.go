package main

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// maxTimeInFailure controls how long a failure is tolerated.
	maxTimeInFailure = 60 * time.Second
)

// Checker validates that DNS is functioning correctly.
type Checker struct {
	// client is the Kubernetes client used to query endpoints.
	client *kubernetes.Clientset
	// MaxTimeInFailure is the tolerated failure duration.
	MaxTimeInFailure time.Duration
	// Hostname is the DNS hostname to resolve.
	Hostname string
	// CheckTimeout is the overall timeout for the check.
	CheckTimeout time.Duration
	// Namespace for endpoint lookup.
	Namespace string
	// LabelSelector to locate DNS endpoints.
	LabelSelector string
	// NodeName where the checker is scheduled.
	NodeName string
}

// Run executes DNS checks with a timeout.
func (dc *Checker) Run(client *kubernetes.Clientset) error {
	// Attach the client to the checker.
	dc.client = client

	// Run the check in a goroutine.
	doneChan := runChecksAsync(dc)

	// Wait for timeout or completion.
	select {
	case <-time.After(dc.CheckTimeout):
		message := "Failed to complete DNS Status check in time! Timeout was reached."
		err := checkclient.ReportFailure([]string{message})
		if err != nil {
			log.Println("Error reporting failure to Kuberhealthy servers:", err)
			return err
		}
		return errors.New(message)
	case err := <-doneChan:
		if err != nil {
			return reportKHFailure(err.Error())
		}
		return reportKHSuccess()
	}
}

// runChecksAsync runs doChecks in a goroutine.
func runChecksAsync(dc *Checker) <-chan error {
	// Create a channel to signal completion.
	doneChan := make(chan error, 1)

	// Run checks in a goroutine.
	go runChecksWorker(dc, doneChan)

	return doneChan
}

// runChecksWorker executes the check and reports completion.
func runChecksWorker(dc *Checker, doneChan chan<- error) {
	// Execute checks and send the result.
	doneChan <- dc.doChecks()
}

// createResolver creates a Resolver object to use for DNS queries.
func createResolver(ip string) (*net.Resolver, error) {
	// Validate IP input.
	if len(ip) < 1 {
		return &net.Resolver{}, errors.New("need a valid ip to create Resolver")
	}

	// Build the resolver for a specific IP.
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Millisecond * time.Duration(10000)}
			return d.DialContext(ctx, "udp", ip+":53")
		},
	}

	return resolver, nil
}

// getIpsFromEndpoint extracts IPs from the endpoints list.
func getIpsFromEndpoint(endpoints *v1.EndpointsList) ([]string, error) {
	// Initialize result list.
	var ipList []string
	if len(endpoints.Items) == 0 {
		return ipList, errors.New("no endpoints found")
	}

	// Walk endpoints, subsets, and addresses.
	for ep := 0; ep < len(endpoints.Items); ep++ {
		for sub := 0; sub < len(endpoints.Items[ep].Subsets); sub++ {
			for address := 0; address < len(endpoints.Items[ep].Subsets[sub].Addresses); address++ {
				ipList = append(ipList, endpoints.Items[ep].Subsets[sub].Addresses[address].IP)
			}
		}
	}

	if len(ipList) != 0 {
		return ipList, nil
	}

	return ipList, errors.New("no ips found in endpoints list")
}

// dnsLookup resolves a hostname with the provided resolver.
func dnsLookup(resolver *net.Resolver, host string) error {
	// Execute the DNS lookup.
	_, err := resolver.LookupHost(context.Background(), host)
	if err != nil {
		message := "DNS Status check determined that " + host + " is DOWN: " + err.Error()
		return errors.New(message)
	}

	return nil
}

// checkEndpoints resolves the hostname via DNS endpoints.
func (dc *Checker) checkEndpoints() error {
	// Query endpoints with the label selector.
	endpoints, err := dc.client.CoreV1().Endpoints(dc.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: dc.LabelSelector})
	if err != nil {
		message := "DNS status check unable to get dns endpoints from cluster: " + err.Error()
		log.Errorln(message)
		return errors.New(message)
	}

	// Get ips from endpoint list to check.
	ips, err := getIpsFromEndpoint(endpoints)
	if err != nil {
		return err
	}

	// Create resolvers and run lookups.
	if len(ips) > 0 {
		for ip := 0; ip < len(ips); ip++ {
			resolver, resolverErr := createResolver(ips[ip])
			if resolverErr != nil {
				return resolverErr
			}
			lookupErr := dnsLookup(resolver, dc.Hostname)
			if lookupErr != nil {
				return lookupErr
			}
		}

		log.Infoln("DNS Status check from service endpoint determined that", dc.Hostname, "was OK.")
		return nil
	}

	return errors.New("no ips found in endpoint with label: " + dc.LabelSelector)
}

// doChecks validates DNS resolution.
func (dc *Checker) doChecks() error {
	// Log the hostname under test.
	log.Infoln("DNS Status check testing hostname:", dc.Hostname)

	// Check endpoints if a selector is provided.
	if len(dc.LabelSelector) > 0 {
		err := dc.checkEndpoints()
		if err != nil {
			return err
		}
		return nil
	}

	// Otherwise do lookup against service endpoint.
	_, err := net.LookupHost(dc.Hostname)
	if err != nil {
		message := "DNS Status check determined that " + dc.Hostname + " is DOWN: " + err.Error()
		log.Errorln(message)
		return errors.New(message)
	}

	log.Infoln("DNS Status check from service endpoint determined that", dc.Hostname, "was OK.")
	return nil
}

// reportKHSuccess reports success to Kuberhealthy servers.
func reportKHSuccess() error {
	// Report success and confirm delivery.
	err := checkclient.ReportSuccess()
	if err != nil {
		log.Println("Error reporting success to Kuberhealthy servers:", err)
		return err
	}
	log.Println("Successfully reported success to Kuberhealthy servers")
	return nil
}

// reportKHFailure reports failure to Kuberhealthy servers.
func reportKHFailure(errorMessage string) error {
	// Report failure and confirm delivery.
	err := checkclient.ReportFailure([]string{errorMessage})
	if err != nil {
		log.Println("Error reporting failure to Kuberhealthy servers:", err)
		return err
	}
	log.Println("Successfully reported failure to Kuberhealthy servers")
	return nil
}
