package main

import (
	"clawrt/internal/config"
	"clawrt/internal/hotplug"
	"clawrt/internal/llm"
	"clawrt/internal/skills"
	"clawrt/internal/telegram"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const Version = "1.0.0-openwrt"

func main() {
	configPath := flag.String("config", "", "Ruta al archivo de configuración UCI o JSON")
	showVersion := flag.Bool("version", false, "Muestra la versión de ClawRT")
	testLLM := flag.Bool("test-llm", false, "Prueba de conectividad con el proveedor de LLM")
	fetchModels := flag.Bool("fetch-models", false, "Consulta y obtiene la lista de modelos disponibles en vivo del proveedor")
	testTelegram := flag.Bool("test-telegram", false, "Prueba del Bot Token de Telegram")
	clearFacts := flag.Bool("clear-facts", false, "Limpia los hechos aprendidos en /tmp/clawrt_facts.json")
	notifyHotplug := flag.Bool("notify-hotplug", false, "Procesa y envía alertas de eventos hotplug del sistema OpenWrt")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ClawRT OpenWrt Agent v%s\n", Version)
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

	log.Printf("[START] Iniciando ClawRT OpenWrt Agent v%s...", Version)

	if !cfg.Enabled {
		log.Println("[INFO] ClawRT está deshabilitado en la configuración (/etc/config/clawrt). Saliendo.")
		os.Exit(0)
	}

	// Initialize skills registry
	registry := skills.NewRegistry()

	// Initialize Telegram Bot & Polling
	bot := telegram.NewBot(cfg, registry)

	stopChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go bot.StartPolling(stopChan)

	log.Println("[RUNNING] ClawRT está en ejecución. Presione Ctrl+C para detener.")

	sig := <-sigChan
	log.Printf("[SIGNAL] Señal recibida (%v). Deteniendo ClawRT...", sig)

	close(stopChan)
	log.Println("[STOP] ClawRT detenido correctamente.")
}
