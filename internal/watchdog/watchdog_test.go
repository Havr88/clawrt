package watchdog

import (
	"testing"
)

func TestWatchdogDiagnosticStructure(t *testing.T) {
	wd := GetWatchdog()
	if wd == nil {
		t.Fatalf("GetWatchdog devolvió nil")
	}

	diag := wd.RunDiagnostic()
	if diag == nil {
		t.Fatalf("RunDiagnostic devolvió nil")
	}

	if diag.OverallStatus != StatusHealthy && diag.OverallStatus != StatusDegraded && diag.OverallStatus != StatusDown {
		t.Errorf("Estado de watchdog no esperado: %s", diag.OverallStatus)
	}
}
