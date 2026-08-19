package netintel

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/skip2/go-qrcode"
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

	asciiQR := generateNativeASCIIQR(qrPayload)

	return &WiFiQRResult{
		SSID:       ssid,
		Encryption: enc,
		Password:   pass,
		QRPayload:  qrPayload,
		ASCIIBlock: asciiQR,
	}, nil
}

func generateNativeASCIIQR(payload string) string {
	q, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return "████████████ (QR Error) ████████████"
	}
	return q.ToSmallString(false)
}
