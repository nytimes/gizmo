package config_test

import (
	"testing"

	"github.com/NYTimes/gizmo/config"
)

func TestServerVanillaConfig(t *testing.T) {
	server := config.Server{}
	portBound := server.PortBound()
	logsPopulated := server.LogsPopulated()
	healthCheckSet := server.HealthCheckSet()

	if portBound != false {
		t.Errorf("Expected portBound to be false, got %v", portBound)
	}

	if logsPopulated != false {
		t.Errorf("Expected logsPopulated to be false, got %v", logsPopulated)
	}

	if healthCheckSet != false {
		t.Errorf("Expected healhCheckSet to be false, got %v", healthCheckSet)
	}
}

func TestServerVanillaConfig(t *testing.T) {
	server := new(config.Server)
	server.HTTPPort = 80
	server.HTTPAccessLog = "LOG"
	server.HealthCheckType = "PINGDOM"
	portBound := server.PortBound()
	logsPopulated := server.LogsPopulated()
	healthCheckSet := server.HealthCheckSet()

	if portBound != true {
		t.Errorf("Expected portBound to be true, got %v", portBound)
	}

	if logsPopulated != true {
		t.Errorf("Expected logsPopulated to be true, got %v", logsPopulated)
	}

	if healthCheckSet != true {
		t.Errorf("Expected healhCheckSet to be true, got %v", healthCheckSet)
	}
}
