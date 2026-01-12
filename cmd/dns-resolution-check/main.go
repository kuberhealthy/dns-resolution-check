package main

import (
	"context"
	"time"

	nodecheck "github.com/kuberhealthy/kuberhealthy/v3/pkg/nodecheck"
	log "github.com/sirupsen/logrus"
)

// main wires configuration, dependencies, and executes the DNS check.
func main() {
	// Parse configuration from environment variables.
	cfg, err := parseConfig()
	if err != nil {
		log.Errorln("Configuration error:", err.Error())
		return
	}

	// Create a Kubernetes client.
	client, err := createKubeClient()
	if err != nil {
		log.Fatalln("Unable to create kubernetes client", err)
	}

	// Create the checker instance.
	dc := New(cfg)

	// Enable nodecheck debug output for parity with v2.
	nodecheck.EnableDebugOutput()

	// Create context for node readiness checks.
	checkTimeLimit := time.Minute * 1
	ctx, _ := context.WithTimeout(context.Background(), checkTimeLimit)

	// Wait for node age.
	minNodeAge := time.Minute * 3
	err = nodecheck.WaitForNodeAge(ctx, client, cfg.NodeName, minNodeAge)
	if err != nil {
		log.Errorln("Error waiting for node to reach minimum age:", err)
	}

	// Wait for Kuberhealthy endpoint readiness.
	err = nodecheck.WaitForKuberhealthy(ctx)
	if err != nil {
		log.Errorln("Error waiting for Kuberhealthy to be ready:", err)
	}

	// Run the DNS status check.
	err = dc.Run(client)
	if err != nil {
		log.Errorln("Error running DNS Status check for hostname:", cfg.Hostname)
	}
	log.Infoln("Done running DNS Status check for hostname:", cfg.Hostname)
}

// New returns a new DNS Checker.
func New(cfg *CheckConfig) *Checker {
	// Build a checker with configuration inputs.
	return &Checker{
		Hostname:         cfg.Hostname,
		MaxTimeInFailure: maxTimeInFailure,
		CheckTimeout:     cfg.CheckTimeout,
		Namespace:        cfg.Namespace,
		LabelSelector:    cfg.LabelSelector,
		NodeName:         cfg.NodeName,
	}
}
