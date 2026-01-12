package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultCheckTimeout is used when KH_CHECK_RUN_DEADLINE is unavailable.
	defaultCheckTimeout = 5 * time.Minute
	// deadlinePadding keeps a small buffer before the Kuberhealthy deadline.
	deadlinePadding = 5 * time.Second
)

// CheckConfig stores configuration for the DNS resolution check.
type CheckConfig struct {
	// Hostname is the DNS hostname to resolve.
	Hostname string
	// NodeName is the node where the pod is scheduled.
	NodeName string
	// Namespace is where DNS endpoints are located when provided.
	Namespace string
	// LabelSelector is used to select DNS endpoints.
	LabelSelector string
	// CheckTimeout controls how long the check waits before timing out.
	CheckTimeout time.Duration
}

// parseConfig loads environment variables into a CheckConfig.
func parseConfig() (*CheckConfig, error) {
	// Start with default settings.
	cfg := &CheckConfig{}
	cfg.CheckTimeout = defaultCheckTimeout

	// Apply deadline from Kuberhealthy when available.
	deadline, err := checkclient.GetDeadline()
	if err == nil {
		remaining := deadline.Sub(time.Now().Add(deadlinePadding))
		if remaining > 0 {
			cfg.CheckTimeout = remaining
		}
	}
	if err != nil {
		log.Infoln("There was an issue getting the check deadline:", err.Error())
	}

	// Read required hostname.
	cfg.Hostname = os.Getenv("HOSTNAME")
	if len(cfg.Hostname) == 0 {
		return nil, fmt.Errorf("HOSTNAME environment variable has not been set")
	}

	// Read required node name.
	cfg.NodeName = os.Getenv("NODE_NAME")
	if len(cfg.NodeName) == 0 {
		return nil, fmt.Errorf("failed to retrieve NODE_NAME environment variable")
	}

	// Optional namespace and label selector.
	cfg.Namespace = os.Getenv("NAMESPACE")
	cfg.LabelSelector = os.Getenv("DNS_POD_SELECTOR")

	return cfg, nil
}
