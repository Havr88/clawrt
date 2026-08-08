package netintel

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type DeviceInfo struct {
	IP            string    `json:"ip"`
	MAC           string    `json:"mac"`
	Hostname      string    `json:"hostname"`
	LeaseExpiry   time.Time `json:"lease_expiry"`
	IsRandomMAC   bool      `json:"is_random_mac"`
	Vendor        string    `json:"vendor"`
	WiFiInterface string    `json:"wifi_interface,omitempty"`
	SignalRSSI    int       `json:"signal_rssi,omitempty"` // dBm
	PHYRate       int       `json:"phy_rate,omitempty"`   // Mbps
	ProbableOS    string    `json:"probable_os"`
	OpenPorts     []int     `json:"open_ports,omitempty"`
}

var ouiVendorMap = map[string]string{
	"00:03:93": "Apple",
	"00:05:02": "Apple",
	"00:0a:95": "Apple",
	"00:10:fa": "Apple",
	"00:11:24": "Apple",
	"00:1c:b3": "Apple",
	"00:1e:c2": "Apple",
	"00:21:e9": "Apple",
	"00:23:12": "Apple",
	"00:25:00": "Apple",
	"00:26:08": "Apple",
	"00:26:b0": "Apple",
	"00:00:f0": "Samsung",
	"00:02:78": "Samsung",
	"00:07:ab": "Samsung",
	"00:12:fb": "Samsung",
	"00:15:99": "Samsung",
	"00:18:af": "Samsung",
	"00:1c:43": "Samsung",
	"00:21:d2": "Samsung",
	"00:24:90": "Samsung",
	"00:1a:2b": "Cisco",
	"00:27:0d": "Cisco",
	"00:04:4b": "NVIDIA",
	"00:09:6b": "IBM",
	"00:15:5d": "Microsoft",
	"00:50:56": "VMware",
	"00:0c:29": "VMware",
	"00:16:3e": "Xen",
	"00:1e:67": "Intel",
	"00:21:6b": "Intel",
	"00:22:fb": "Intel",
	"00:24:d7": "Intel",
	"00:27:0e": "Intel",
	"00:ec:0a": "Xiaomi",
	"00:9e:c8": "Xiaomi",
	"18:1e:78": "Xiaomi",
	"28:6c:07": "Xiaomi",
	"34:80:b5": "Xiaomi",
	"64:09:80": "Xiaomi",
	"b0:e2:35": "Xiaomi",
	"d4:61:9d": "Xiaomi",
	"f4:8b:32": "Xiaomi",
	"b8:27:eb": "Raspberry Pi",
	"dc:a6:32": "Raspberry Pi",
	"e4:5f:01": "Raspberry Pi",
	"24:0a:c4": "Espressif (ESP32/ESP8266)",
	"30:ae:a4": "Espressif (ESP32/ESP8266)",
	"54:5a:a6": "Espressif (ESP32/ESP8266)",
	"60:01:94": "Espressif (ESP32/ESP8266)",
	"84:0d:8e": "Espressif (ESP32/ESP8266)",
	"a0:20:a6": "Espressif (ESP32/ESP8266)",
	"bc:dd:c2": "Espressif (ESP32/ESP8266)",
	"cc:50:e3": "Espressif (ESP32/ESP8266)",
}

// IsRandomizedMAC checks if the MAC address has the Locally Administered Bit set (bit 1 of octet 0)
func IsRandomizedMAC(macStr string) bool {
	hw, err := net.ParseMAC(macStr)
	if err != nil || len(hw) < 1 {
		return false
	}
	// If bit 1 of byte 0 is 1 (0x02 mask), it is a randomized/private MAC address
	return (hw[0] & 0x02) != 0
}

func LookupVendor(macStr string) string {
	macLower := strings.ToLower(strings.TrimSpace(macStr))
	if len(macLower) >= 8 {
		prefix := macLower[:8]
		if vendor, ok := ouiVendorMap[prefix]; ok {
			return vendor
		}
	}
	if IsRandomizedMAC(macStr) {
		return "MAC Privada / Aleatoria (iOS/Android Privacy)"
	}
	return "Desconocido"
}

func EstimateOS(hostname string, openPorts []int) string {
	h := strings.ToLower(hostname)

	if strings.Contains(h, "iphone") || strings.Contains(h, "ipad") || strings.Contains(h, "macbook") || strings.Contains(h, "apple") {
		return "Apple iOS / macOS"
	}
	if strings.Contains(h, "android") || strings.Contains(h, "galaxy") || strings.Contains(h, "redmi") || strings.Contains(h, "pixel") || strings.Contains(h, "huawei") {
		return "Android OS"
	}
	if strings.Contains(h, "win") || strings.Contains(h, "desktop") || strings.Contains(h, "laptop") {
		return "Microsoft Windows"
	}
	if strings.Contains(h, "raspberry") || strings.Contains(h, "ubuntu") || strings.Contains(h, "debian") || strings.Contains(h, "arch") {
		return "Linux OS"
	}
	if strings.Contains(h, "esp8266") || strings.Contains(h, "esp32") || strings.Contains(h, "tasmota") || strings.Contains(h, "shelly") {
		return "IoT Firmware (ESP32/ESP8266)"
	}

	for _, p := range openPorts {
		if p == 5555 {
			return "Android OS (ADB Open)"
		}
		if p == 3389 || p == 445 {
			return "Microsoft Windows"
		}
	}

	return "Dispositivo Estándar LAN"
}

func GetEnrichedLeases() ([]DeviceInfo, error) {
	leasesFile := "/tmp/dhcp.leases"
	file, err := os.Open(leasesFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir %s: %w", leasesFile, err)
	}
	defer file.Close()

	var devices []DeviceInfo
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 {
			timestamp, _ := strconv.ParseInt(fields[0], 10, 64)
			mac := fields[1]
			ip := fields[2]
			hostname := fields[3]

			if hostname == "*" {
				hostname = "desconocido"
			}

			dev := DeviceInfo{
				IP:          ip,
				MAC:         mac,
				Hostname:    hostname,
				LeaseExpiry: time.Unix(timestamp, 0),
				IsRandomMAC: IsRandomizedMAC(mac),
				Vendor:      LookupVendor(mac),
			}

			dev.ProbableOS = EstimateOS(hostname, nil)
			devices = append(devices, dev)
		}
	}

	// Enrich with WiFi information if iwinfo or ubus is available
	enrichWiFiInfo(devices)

	return devices, nil
}

func enrichWiFiInfo(devices []DeviceInfo) {
	out, err := exec.Command("ubus", "call", "iwinfo", "assoclist", `{"device":"wlan0"}`).Output()
	if err != nil || len(out) == 0 {
		return
	}
	// Simple lookup in iwinfo output for signal strength
	outputStr := string(out)
	for i := range devices {
		if strings.Contains(outputStr, strings.ToLower(devices[i].MAC)) {
			devices[i].WiFiInterface = "wlan0"
		}
	}
}
