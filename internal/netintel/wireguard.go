package netintel

import (
	"bufio"
	"bytes"
	"clawrt/internal/sys"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type WireGuardPeer struct {
	PublicKey           string    `json:"public_key"`
	Endpoint            string    `json:"endpoint"`
	AllowedIPs          []string  `json:"allowed_ips"`
	LatestHandshakeSec  int64     `json:"latest_handshake_sec"`
	LatestHandshakeTime time.Time `json:"latest_handshake_time"`
	TransferRxBytes     int64     `json:"transfer_rx_bytes"`
	TransferTxBytes     int64     `json:"transfer_tx_bytes"`
	Status              string    `json:"status"` // ACTIVE, STALLED, NEVER_CONNECTED
}

type WireGuardInterfaceStatus struct {
	InterfaceName  string          `json:"interface_name"`
	ListenPort     int             `json:"listen_port"`
	PublicKey      string          `json:"public_key"`
	Peers          []WireGuardPeer `json:"peers"`
	HealthyCount   int             `json:"healthy_count"`
	StalledCount   int             `json:"stalled_count"`
	Recommendation string          `json:"recommendation"`
}

func ManageWireGuard(iface string, autoReconnect bool) (*WireGuardInterfaceStatus, error) {
	if iface == "" {
		iface = "wg0"
	}

	res := &WireGuardInterfaceStatus{
		InterfaceName: iface,
		Peers:         make([]WireGuardPeer, 0),
	}

	// 1. Run wg show <iface> dump
	out, err := exec.Command("wg", "show", iface, "dump").Output()
	if err != nil || len(out) == 0 {
		res.Recommendation = fmt.Sprintf("⚠️ No se pudo consultar la interfaz WireGuard '%s' (posiblemente inactiva o no configurada).", iface)
		return res, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	isFirstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if isFirstLine {
			// Interface line: private-key public-key listen-port fwmark
			if len(fields) >= 3 {
				res.PublicKey = fields[1]
				res.ListenPort, _ = strconv.Atoi(fields[2])
			}
			isFirstLine = false
			continue
		}

		// Peer line: public-key preshared-key endpoint allowed-ips latest-handshake transfer-rx transfer-tx persistent-keepalive
		if len(fields) >= 8 {
			hsSec, _ := strconv.ParseInt(fields[4], 10, 64)
			rxBytes, _ := strconv.ParseInt(fields[5], 10, 64)
			txBytes, _ := strconv.ParseInt(fields[6], 10, 64)

			peer := WireGuardPeer{
				PublicKey:          fields[0],
				Endpoint:           fields[2],
				AllowedIPs:         strings.Split(fields[3], ","),
				LatestHandshakeSec: hsSec,
				TransferRxBytes:    rxBytes,
				TransferTxBytes:    txBytes,
			}

			nowSec := time.Now().Unix()
			if hsSec == 0 {
				peer.Status = "NEVER_CONNECTED"
				res.StalledCount++
			} else if nowSec-hsSec > 180 {
				peer.Status = "STALLED"
				peer.LatestHandshakeTime = time.Unix(hsSec, 0)
				res.StalledCount++
			} else {
				peer.Status = "ACTIVE"
				peer.LatestHandshakeTime = time.Unix(hsSec, 0)
				res.HealthyCount++
			}

			res.Peers = append(res.Peers, peer)
		}
	}

	if res.StalledCount > 0 {
		res.Recommendation = fmt.Sprintf("⚠️ Se detectaron %d peer(s) WireGuard con handshake expirado (> 180s).", res.StalledCount)
		if autoReconnect {
			_ = exec.Command("/sbin/ifup", iface).Run()
			_, _ = sys.ExecuteTypedServiceRestart("network")
			res.Recommendation += " (Se ejecutó ifup/reinicio de túnel para forzar nuevo handshake)."
		}
	} else if len(res.Peers) > 0 {
		res.Recommendation = fmt.Sprintf("✅ Todos los peers WireGuard (%d activos) mantienen un handshake saludable.", res.HealthyCount)
	} else {
		res.Recommendation = "ℹ️ No hay peers activos configurados en esta interfaz WireGuard."
	}

	return res, nil
}
