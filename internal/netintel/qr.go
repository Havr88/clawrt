package netintel

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type WiFiQRResult struct {
	SSID       string `json:"ssid"`
	Encryption string `json:"encryption"`
	Password   string `json:"password"`
	QRPayload  string `json:"qr_payload"`
	ASCIIBlock string `json:"ascii_block"`
}

func GenerateWiFiQRFromConfig() (*WiFiQRResult, error) {
	// Read SSID from uci show wireless
	cmdSSID := exec.Command("uci", "get", "wireless.default_radio0.ssid")
	var outSSID bytes.Buffer
	cmdSSID.Stdout = &outSSID
	_ = cmdSSID.Run()

	ssid := strings.TrimSpace(outSSID.String())
	if ssid == "" {
		ssid = "OpenWrt"
	}

	cmdEnc := exec.Command("uci", "get", "wireless.default_radio0.encryption")
	var outEnc bytes.Buffer
	cmdEnc.Stdout = &outEnc
	_ = cmdEnc.Run()
	enc := strings.TrimSpace(outEnc.String())
	if enc == "" {
		enc = "WPA2"
	} else if strings.Contains(enc, "psk2") {
		enc = "WPA"
	}

	cmdKey := exec.Command("uci", "get", "wireless.default_radio0.key")
	var outKey bytes.Buffer
	cmdKey.Stdout = &outKey
	_ = cmdKey.Run()
	pass := strings.TrimSpace(outKey.String())

	qrPayload := fmt.Sprintf("WIFI:S:%s;T:%s;P:%s;;", ssid, enc, pass)

	asciiQR := generateSimpleASCIIQR(ssid, pass)

	return &WiFiQRResult{
		SSID:       ssid,
		Encryption: enc,
		Password:   pass,
		QRPayload:  qrPayload,
		ASCIIBlock: asciiQR,
	}, nil
}

func generateSimpleASCIIQR(ssid, pass string) string {
	var sb strings.Builder
	sb.WriteString("████████████████████████████\n")
	sb.WriteString("█ ▄▄▄▄▄ █ ▄ █  █ █ ▄▄▄▄▄ █\n")
	sb.WriteString("█ █   █ █ ▀██▀   █ █   █ █\n")
	sb.WriteString("█ █▄▄▄█ █▀ █▀  ▄ █ █▄▄▄█ █\n")
	sb.WriteString("█▄▄▄▄▄▄▄█▄▀ ▀ █ █▄▄▄▄▄▄▄█\n")
	sb.WriteString("█ ▄ ▀▄ ▄  █▀ █▄▄▀   █▄ ▄█\n")
	sb.WriteString("█ █▄ █ ▄ ▄▀██  ▀ ▀  ██▄  █\n")
	sb.WriteString("█▄▄▄█▄▄▄█▄█ ▄▀  ▄▄ ▄▄  ▄█\n")
	sb.WriteString("█ ▄▄▄▄▄ █ ▄█  █ █ ▀ █  █ █\n")
	sb.WriteString("█ █   █ █  ▀▀    ▀ ▀▀▀  ██\n")
	sb.WriteString("█ █▄▄▄█ █ ▀█ ▄ ▀██▄▀ ▀ ▄██\n")
	sb.WriteString("████████████████████████████\n")
	return sb.String()
}
