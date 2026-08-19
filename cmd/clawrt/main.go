package main

import (
	"clawrt/internal/config"
	"clawrt/internal/hotplug"
	"clawrt/internal/llm"
	"clawrt/internal/netintel"
	"clawrt/internal/skills"
	"clawrt/internal/sys"
	"clawrt/internal/telegram"
	"clawrt/internal/ubus"
	"clawrt/internal/watchdog"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const Version = "1.1.0-openwrt"

func main() {
	configPath := flag.String("config", "", "Ruta al archivo de configuración UCI o JSON")
	showVersion := flag.Bool("version", false, "Muestra la versión de ClawRT")
	testLLM := flag.Bool("test-llm", false, "Prueba de conectividad con el proveedor de LLM")
	fetchModels := flag.Bool("fetch-models", false, "Consulta y obtiene la lista de modelos disponibles en vivo del proveedor")
	testTelegram := flag.Bool("test-telegram", false, "Prueba del Bot Token de Telegram")
	clearFacts := flag.Bool("clear-facts", false, "Limpia los hechos aprendidos en /tmp/clawrt_facts.json")
	notifyHotplug := flag.Bool("notify-hotplug", false, "Procesa y envía alertas de eventos hotplug del sistema OpenWrt")
	runDiagnose := flag.Bool("diagnose", false, "Ejecuta diagnóstico completo de conectividad y auto-sanación")
	optimizeWiFi := flag.Bool("optimize-wifi", false, "Escanea espectro WiFi y calcula el canal óptimo")
	queryPrompt := flag.String("query", "", "Ejecuta una consulta directa al agente (LuCI Web Copilot / CLI)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ClawRT OpenWrt Agent v%s (Autonomous Network Agent)\n", Version)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[FATAL] Error al cargar la configuración: %v", err)
	}

	if *notifyHotplug {
		err := hotplug.ProcessHotplugEvent(cfg, flag.Args())
		if err != nil {
			log.Printf("[ERROR] Hotplug notification failed: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *clearFacts {
		_ = os.Remove("/tmp/clawrt_facts.json")
		fmt.Println("✅ Hechos dinámicos eliminados correctamente de /tmp/clawrt_facts.json")
		os.Exit(0)
	}

	if *runDiagnose {
		wd := watchdog.GetWatchdog()
		diag := wd.RunDiagnostic()
		if diag.OverallStatus != watchdog.StatusHealthy {
			recovered, actions := wd.AutoHeal(diag)
			diag.RecoveryActions = actions
			if recovered {
				diag.OverallStatus = watchdog.StatusHealthy
			}
		}
		b, _ := json.MarshalIndent(diag, "", "  ")
		fmt.Println(string(b))
		os.Exit(0)
	}

	if *optimizeWiFi {
		report, err := netintel.OptimizeWiFiChannels("wlan0", false)
		if err != nil {
			fmt.Printf("❌ ERROR_WIFI_OPT: %v\n", err)
			os.Exit(1)
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
		os.Exit(0)
	}

	if *queryPrompt != "" {
		registry := skills.NewRegistry()
		bot := telegram.NewBot(cfg, registry)
		response := bot.ProcessDirectQuery(*queryPrompt)
		fmt.Println(response)
		os.Exit(0)
	}

	if *fetchModels {
		client := llm.NewClient(cfg.Provider, cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.FallbackModel)
		models, err := client.FetchModels()
		if err != nil {
			fmt.Printf("❌ ERROR_FETCH_MODELS: %v\n", err)
			os.Exit(1)
		}
		b, _ := json.MarshalIndent(models, "", "  ")
		fmt.Println(string(b))
		os.Exit(0)
	}

	if *testLLM {
		client := llm.NewClient(cfg.Provider, cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.FallbackModel)
		start := time.Now()
		msgs := []llm.ChatMessage{
			{Role: "user", Content: "Responder unicamente 'OK'"},
		}
		resp, err := client.ChatCompletion(msgs, nil)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("❌ ERROR_LLM: %v (Latencia: %v)\n", err, elapsed)
			os.Exit(1)
		}
		fmt.Printf("✅ SUCCESS_LLM: Modelo '%s' respondió en %v: %s\n", cfg.LLMModel, elapsed, resp.Content)
		os.Exit(0)
	}

	if *testTelegram {
		if cfg.BotToken == "" {
			fmt.Println("❌ ERROR_TELEGRAM: Bot Token no configurado en /etc/config/clawrt")
			os.Exit(1)
		}
		registry := skills.NewRegistry()
		bot := telegram.NewBot(cfg, registry)
		if len(cfg.ChatIDs) > 0 {
			err := bot.SendMessage(cfg.ChatIDs[0], "🧪 Mensaje de prueba enviado desde la interfaz LuCI de ClawRT AI.")
			if err != nil {
				fmt.Printf("❌ ERROR_TELEGRAM: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ SUCCESS_TELEGRAM: Mensaje de prueba enviado al Chat ID %d\n", cfg.ChatIDs[0])
		} else {
			fmt.Println("✅ SUCCESS_TELEGRAM: Token válido (sin Chat ID configurado para probar envío)")
		}
		os.Exit(0)
	}

	log.Printf("[START] Iniciando ClawRT OpenWrt Agent v%s (Autonomous Mode)...", Version)

	// Tune Go Runtime Memory Limit dynamically based on detected RAM
	sysInfo := sys.GetSystemInfo()
	var memLimitBytes int64 = 8 * 1024 * 1024 // Default 8MB limit for 64MB RAM
	if sysInfo.MemoryTotalMB <= 32 {
		memLimitBytes = 4 * 1024 * 1024 // 4MB soft limit for 32MB RAM
	} else if sysInfo.MemoryTotalMB >= 256 {
		memLimitBytes = 64 * 1024 * 1024 // 64MB limit for high RAM
	} else if sysInfo.MemoryTotalMB >= 128 {
		memLimitBytes = 16 * 1024 * 1024 // 16MB limit for 128MB RAM
	}
	sys.ApplyRuntimeMemoryLimit(memLimitBytes)

	if !cfg.Enabled {
		log.Println("[INFO] ClawRT está deshabilitado en la configuración (/etc/config/clawrt). Saliendo.")
		os.Exit(0)
	}

	// Initialize skills registry
	registry := skills.NewRegistry()

	// Initialize Telegram Bot
	bot := telegram.NewBot(cfg, registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start Self-Healing Watchdog in background
	wd := watchdog.GetWatchdog()
	wd.SetAlertCallback(func(alertMsg string) {
		for _, cid := range cfg.ChatIDs {
			_ = bot.SendMessage(cid, alertMsg)
		}
	})
	go wd.Start(ctx)

	// 2. Start UBUS Reactive Listener in background
	uListener := ubus.GetListener()
	uListener.Subscribe(func(eventType string, data map[string]interface{}) {
		if eventType == "network.interface" {
			if action, ok := data["action"].(string); ok && action == "ifdown" {
				if iface, ok := data["interface"].(string); ok && iface == "wan" {
					log.Println("[UBUS_EVENT] Caída de interfaz WAN detectada en tiempo real. Disparando Watchdog...")
					diag := wd.RunDiagnostic()
					if diag.OverallStatus != watchdog.StatusHealthy {
						wd.AutoHeal(diag)
					}
				}
			}
		}
	})
	go uListener.Start(ctx)

	// 3. Start Telegram Polling in background
	stopChan := make(chan struct{})
	go bot.StartPolling(stopChan)

	log.Println("[RUNNING] ClawRT Agent activo en modo autónomo continuo (Watchdog + UBUS + Telegram).")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[SIGNAL] Señal recibida (%v). Deteniendo ClawRT...", sig)

	close(stopChan)
	cancel()
	log.Println("[STOP] ClawRT detenido correctamente.")
}
