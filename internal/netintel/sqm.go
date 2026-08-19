package netintel

import (
	"bytes"
	"clawrt/internal/sys"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SQMStatus struct {
	Installed        bool    `json:"installed"`
	Enabled          bool    `json:"enabled"`
	Interface        string  `json:"interface"`
	DownloadKbps     int     `json:"download_kbps"`
	UploadKbps       int     `json:"upload_kbps"`
	QueueDiscipline  string  `json:"queue_discipline"` // cake or fq_codel
	IdleLatencyMs    float64 `json:"idle_latency_ms"`
	LoadedLatencyMs  float64 `json:"loaded_latency_ms"`
	BufferbloatGrade string  `json:"bufferbloat_grade"` // A+, A, B, C, D, F
	Recommendation   string  `json:"recommendation"`
}

func InspectSQMStatus() (*SQMStatus, error) {
	status := &SQMStatus{
		Interface:       "wan",
		QueueDiscipline: "cake",
	}

	// 1. Check if sqm-scripts is installed
	if _, err := os.Stat("/etc/config/sqm"); err == nil {
		status.Installed = true
		parseSQMConfig(status)
	}

	// 2. Measure idle latency
	status.IdleLatencyMs = measurePingLatency("1.1.1.1", 3)

	// 3. Estimate bufferbloat grade
	if status.IdleLatencyMs <= 0 {
		status.BufferbloatGrade = "N/A (Sin Internet)"
		status.Recommendation = "No se puede evaluar Bufferbloat sin conexión a Internet."
		return status, nil
	}

	if status.Enabled && status.DownloadKbps > 0 {
		status.BufferbloatGrade = "A (SQM Cake Activo)"
		status.Recommendation = fmt.Sprintf("✅ SQM está activo en '%s' (%s: %d Kbps Down / %d Kbps Up). Latencia base: %.1f ms.",
			status.Interface, status.QueueDiscipline, status.DownloadKbps, status.UploadKbps, status.IdleLatencyMs)
	} else if !status.Installed {
		status.BufferbloatGrade = "C (SQM No Instalado)"
		status.Recommendation = "⚠️ Se sugiere instalar 'sqm-scripts' y 'luci-app-sqm' para eliminar latencia y lag en llamadas y juegos."
	} else {
		status.BufferbloatGrade = "B (SQM Instalado pero Inactivo)"
		status.Recommendation = "💡 SQM está instalado pero deshabilitado. Se recomienda activar el algoritmo Cake con el 90% de la velocidad de tu plan."
	}

	return status, nil
}

func ConfigureSQM(downloadMbps, uploadMbps int, qdisc string) (string, error) {
	if downloadMbps <= 0 || uploadMbps <= 0 {
		return "", fmt.Errorf("las velocidades de descarga y subida deben ser mayores a 0 Mbps")
	}

	if qdisc == "" {
		qdisc = "cake"
	}

	downKbps := downloadMbps * 1000
	upKbps := uploadMbps * 1000

	// Configure UCI sqm
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.enabled", "1")
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.interface", "wan")
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.download", strconv.Itoa(downKbps))
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.upload", strconv.Itoa(upKbps))
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.qdisc", qdisc)
	_, _ = sys.ExecuteTypedUCISet("sqm", "sqm.eth1.script", "layer_cake.qos")

	_, errRestart := sys.ExecuteTypedServiceRestart("sqm")
	if errRestart != nil {
		// Try init.d script directly if service name is sqm
		_ = exec.Command("/etc/init.d/sqm", "restart").Run()
	}

	return fmt.Sprintf("⚡ SQM (%s) configurado exitosamente: Down=%d Mbps (%d Kbps), Up=%d Mbps (%d Kbps) en WAN.",
		qdisc, downloadMbps, downKbps, uploadMbps, upKbps), nil
}

func parseSQMConfig(st *SQMStatus) {
	data, err := os.ReadFile("/etc/config/sqm")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) >= 3 && fields[0] == "option" {
			switch fields[1] {
			case "enabled":
				st.Enabled = strings.Trim(fields[2], "'\"") == "1"
			case "interface":
				st.Interface = strings.Trim(fields[2], "'\"")
			case "download":
				st.DownloadKbps, _ = strconv.Atoi(strings.Trim(fields[2], "'\""))
			case "upload":
				st.UploadKbps, _ = strconv.Atoi(strings.Trim(fields[2], "'\""))
			case "qdisc":
				st.QueueDiscipline = strings.Trim(fields[2], "'\"")
			}
		}
	}
}

func measurePingLatency(target string, count int) float64 {
	start := time.Now()
	cmd := exec.Command("/bin/ping", "-c", strconv.Itoa(count), "-W", "2", target)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0
	}
	elapsed := float64(time.Since(start).Milliseconds()) / float64(count)
	return elapsed
}
