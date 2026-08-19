package netintel

import (
	"net"
	"strconv"
	"sync"
	"time"
)

type PortScanResult struct {
	IP        string   `json:"ip"`
	OpenPorts []int    `json:"open_ports"`
	Warnings  []string `json:"warnings"`
}

var DefaultCriticalPorts = []int{22, 23, 80, 443, 445, 1900, 5555, 6379, 1883}

func ScanLANPorts(ip string, ports []int) PortScanResult {
	if len(ports) == 0 {
		ports = DefaultCriticalPorts
	}

	result := PortScanResult{
		IP:        ip,
		OpenPorts: []int{},
		Warnings:  []string{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, p := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := net.JoinHostPort(ip, strconv.Itoa(port))
			conn, err := net.DialTimeout("tcp", address, 300*time.Millisecond)
			if err == nil {
				conn.Close()
				mu.Lock()
				result.OpenPorts = append(result.OpenPorts, port)
				switch port {
				case 23:
					result.Warnings = append(result.Warnings, "⚠️ Puerto Telnet (23) Abierto: Tráfico en texto claro no seguro")
				case 5555:
					result.Warnings = append(result.Warnings, "⚠️ Puerto ADB (5555) Abierto: Depuración Android expuesta sin clave")
				case 6379:
					result.Warnings = append(result.Warnings, "⚠️ Puerto Redis (6379) Abierto: Posible base de datos desprotegida")
				case 445:
					result.Warnings = append(result.Warnings, "ℹ️ Puerto SMB (445) Abierto: Compartición de archivos Windows/Samba")
				}
				mu.Unlock()
			}
		}(p)
	}

	wg.Wait()
	return result
}
