package sys

import (
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type MemoryManager struct {
	mu               sync.Mutex
	lastActivityTime time.Time
	idleTimeout      time.Duration
	isSleeping       bool
	stopChan         chan struct{}
}

func ApplyRuntimeMemoryLimit(limitBytes int64) {
	if limitBytes > 0 {
		debug.SetMemoryLimit(limitBytes)
		log.Printf("[MEMORY_MANAGER] ⚙️ Límite suave de memoria Go (debug.SetMemoryLimit) establecido en: %d MB", limitBytes/(1024*1024))
	}
}

var globalMemManager *MemoryManager
var once sync.Once

func GetMemoryManager() *MemoryManager {
	once.Do(func() {
		globalMemManager = &MemoryManager{
			lastActivityTime: time.Now(),
			idleTimeout:      5 * time.Minute, // Default 5 minutes idle timeout
			isSleeping:       false,
			stopChan:         make(chan struct{}),
		}
		go globalMemManager.startIdleMonitor()
	})
	return globalMemManager
}

func (m *MemoryManager) RecordActivity() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastActivityTime = time.Now()
	if m.isSleeping {
		m.isSleeping = false
		log.Println("[MEMORY_MANAGER] ⚡ Reanudado desde modo reposo. Actividad detectada.")
	}
}

func (m *MemoryManager) ForceGC() (uint64, uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	runtime.GC()
	debug.FreeOSMemory()

	runtime.ReadMemStats(&mAfter)

	allocBeforeMB := mBefore.Alloc / (1024 * 1024)
	allocAfterMB := mAfter.Alloc / (1024 * 1024)

	log.Printf("[MEMORY_MANAGER] 🧹 Limpieza manual de RAM realizada: %d MB -> %d MB (Alloc)", allocBeforeMB, allocAfterMB)
	return mBefore.Alloc, mAfter.Alloc
}

func (m *MemoryManager) startIdleMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mu.Lock()
			idleDuration := time.Since(m.lastActivityTime)
			if idleDuration >= m.idleTimeout && !m.isSleeping {
				m.isSleeping = true
				var mBefore, mAfter runtime.MemStats
				runtime.ReadMemStats(&mBefore)

				runtime.GC()
				debug.FreeOSMemory()

				runtime.ReadMemStats(&mAfter)

				log.Printf("[MEMORY_MANAGER] 💤 Modo Reposo activado tras %v de inactividad. Memoria liberada: %d KB -> %d KB",
					idleDuration.Round(time.Second), mBefore.Alloc/1024, mAfter.Alloc/1024)
			}
			m.mu.Unlock()
		}
	}
}

func (m *MemoryManager) Stop() {
	close(m.stopChan)
}
