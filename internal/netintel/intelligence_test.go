package netintel

import (
	"testing"
)

func TestVendorLookup(t *testing.T) {
	appleMAC := "00:03:93:11:22:33"
	vendor := LookupVendor(appleMAC)
	if vendor != "Apple" {
		t.Errorf("Se esperaba Apple, se obtuvo %s", vendor)
	}

	randomMAC := "02:00:00:00:00:00"
	if !IsRandomizedMAC(randomMAC) {
		t.Errorf("02:00:00:00:00:00 debería identificarse como MAC aleatoria")
	}
}

func TestSQMInspect(t *testing.T) {
	st, err := InspectSQMStatus()
	if err != nil {
		t.Fatalf("InspectSQMStatus falló: %v", err)
	}
	if st == nil {
		t.Fatalf("SQMStatus es nil")
	}
}
