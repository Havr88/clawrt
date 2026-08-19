package ubus

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type EventCallback func(eventType string, data map[string]interface{})

type UBUSListener struct {
	mu        sync.RWMutex
	callbacks []EventCallback
	running   bool
}

var (
	listenerInstance *UBUSListener
	once             sync.Once
)

func GetListener() *UBUSListener {
	once.Do(func() {
		listenerInstance = &UBUSListener{
			callbacks: make([]EventCallback, 0),
		}
	})
	return listenerInstance
}

func (u *UBUSListener) Subscribe(cb EventCallback) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.callbacks = append(u.callbacks, cb)
}

func (u *UBUSListener) Start(ctx context.Context) {
	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		return
	}
	u.running = true
	u.mu.Unlock()

	log.Println("[UBUS] Iniciando listener de eventos reactivos del sistema (ubus listen)...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[UBUS] UBUS listener finalizado.")
				return
			default:
				u.listenStream(ctx)
				// Backoff before reconnect if ubus listen crashes
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (u *UBUSListener) listenStream(ctx context.Context) {
	// ubus listen captures network.interface, hostapd, service, dhcp
	cmd := exec.CommandContext(ctx, "ubus", "listen")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[UBUS] Error al abrir pipe de ubus listen: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[UBUS] Error al iniciar ubus listen: %v", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	var currentEventType string
	var jsonBuilder strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// ubus listen outputs in format:
		// { "network.interface": {"action":"ifdown","interface":"wan"} }
		// or lines starting with event type
		if strings.HasPrefix(line, "{") {
			var rawMap map[string]interface{}
			if err := json.Unmarshal([]byte(line), &rawMap); err == nil {
				for evtType, evtData := range rawMap {
					if dataMap, ok := evtData.(map[string]interface{}); ok {
						u.dispatch(evtType, dataMap)
					} else {
						u.dispatch(evtType, map[string]interface{}{"raw": evtData})
					}
				}
				continue
			}
		}

		// Fallback block parser
		if !strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "}") && !strings.Contains(line, ":") {
			currentEventType = line
			jsonBuilder.Reset()
			continue
		}

		if currentEventType != "" {
			jsonBuilder.WriteString(line)
			if strings.HasSuffix(line, "}") {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(jsonBuilder.String()), &data); err == nil {
					u.dispatch(currentEventType, data)
				}
				currentEventType = ""
				jsonBuilder.Reset()
			}
		}
	}

	_ = cmd.Wait()
}

func (u *UBUSListener) dispatch(eventType string, data map[string]interface{}) {
	u.mu.RLock()
	cbs := make([]EventCallback, len(u.callbacks))
	copy(cbs, u.callbacks)
	u.mu.RUnlock()

	for _, cb := range cbs {
		go func(callback EventCallback) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[UBUS] Panic recuperado en callback: %v", r)
				}
			}()
			callback(eventType, data)
		}(cb)
	}
}
