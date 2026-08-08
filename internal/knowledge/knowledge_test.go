package knowledge

import (
	"os"
	"testing"
)

func TestKnowledgeEngineEmbeddedData(t *testing.T) {
	tmpFile := "/tmp/test_clawrt_facts.json"
	defer os.Remove(tmpFile)

	ke := NewKnowledgeEngine(512, tmpFile)

	// Test Tier detection
	if ke.Tier != TierFull {
		t.Errorf("Esperado TierFull para 512MB RAM, se obtuvo: %s", ke.Tier)
	}

	// Test Program lookup (dnsmasq)
	prog, found := ke.GetProgramInfo("dnsmasq")
	if !found {
		t.Fatalf("Programa 'dnsmasq' no encontrado en la base de datos embebida")
	}

	if prog.Package != "dnsmasq" {
		t.Errorf("Paquete esperado 'dnsmasq', se obtuvo '%s'", prog.Package)
	}

	if prog.Config != "/etc/config/dhcp" {
		t.Errorf("Config esperada '/etc/config/dhcp', se obtuvo '%s'", prog.Config)
	}

	// Test Program lookup (firewall)
	fwProg, found := ke.GetProgramInfo("firewall")
	if !found {
		t.Fatalf("Programa 'firewall' no encontrado en la base de datos embebida")
	}

	if fwProg.Package != "fw4" {
		t.Errorf("Paquete de firewall esperado 'fw4', se obtuvo '%s'", fwProg.Package)
	}

	// Test Mutable Facts & Recipe Saving
	ke.SetFact("wan_interface", "eth1")
	val, ok := ke.GetFact("wan_interface")
	if !ok || val != "eth1" {
		t.Errorf("Fallo al guardar o recuperar hecho dinámico. Obtenido: %s", val)
	}

	ke.SaveRecipe("restart_wifi", "wifi reload", "wireless.default_radio0.disabled=0")
	if len(ke.Recipes) == 0 {
		t.Errorf("Fallo al guardar receta de refuerzo")
	}
}
