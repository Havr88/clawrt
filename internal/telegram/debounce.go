package telegram

import (
	"strings"
	"sync"
	"time"
)

type Debouncer struct {
	mu       sync.Mutex
	buffers  map[int64]*messageBuffer
	delay    time.Duration
	callback func(chatID int64, mergedText string)
}

type messageBuffer struct {
	messages []string
	timer    *time.Timer
}

func NewDebouncer(delay time.Duration, callback func(chatID int64, mergedText string)) *Debouncer {
	return &Debouncer{
		buffers:  make(map[int64]*messageBuffer),
		delay:    delay,
		callback: callback,
	}
}

// IsControlCommand identifies commands that MUST BYPASS debounce immediately
func IsControlCommand(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	controlCommands := []string{
		"/stop", "/abort", "/cancel", "/status", "/sysinfo", "/reboot", "/help", "/wifi",
	}
	for _, cmd := range controlCommands {
		if strings.HasPrefix(t, cmd) {
			return true
		}
	}
	return false
}

func (d *Debouncer) PushMessage(chatID int64, text string) {
	// 1. Control commands bypass debounce instantly
	if IsControlCommand(text) {
		d.callback(chatID, text)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	buf, exists := d.buffers[chatID]
	if !exists {
		buf = &messageBuffer{
			messages: []string{text},
		}
		d.buffers[chatID] = buf
	} else {
		buf.messages = append(buf.messages, text)
		if buf.timer != nil {
			buf.timer.Stop()
		}
	}

	buf.timer = time.AfterFunc(d.delay, func() {
		d.flush(chatID)
	})
}

func (d *Debouncer) flush(chatID int64) {
	d.mu.Lock()
	buf, exists := d.buffers[chatID]
	if !exists {
		d.mu.Unlock()
		return
	}
	delete(d.buffers, chatID)
	d.mu.Unlock()

	merged := strings.Join(buf.messages, "\n")
	d.callback(chatID, merged)
}
