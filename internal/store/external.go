package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ExternalDBType string

const (
	DBTypeNone     ExternalDBType = "none"
	DBTypeRedis    ExternalDBType = "redis"
	DBTypeD1       ExternalDBType = "d1"
	DBTypeSupabase ExternalDBType = "supabase"
)

type ExternalDBClient struct {
	Provider   ExternalDBType
	URL        string
	Token      string
	HTTPClient *http.Client
}

func NewExternalDBClient(providerStr, url, token string) *ExternalDBClient {
	p := DBTypeNone
	switch strings.ToLower(providerStr) {
	case "redis":
		p = DBTypeRedis
	case "d1":
		p = DBTypeD1
	case "supabase":
		p = DBTypeSupabase
	}

	return &ExternalDBClient{
		Provider: p,
		URL:      url,
		Token:    token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (db *ExternalDBClient) IsConfigured() bool {
	return db.Provider != DBTypeNone && db.URL != ""
}

func (db *ExternalDBClient) SaveEventAsync(eventType string, data map[string]interface{}) {
	if !db.IsConfigured() {
		return
	}
	go func() {
		_ = db.SaveEvent(eventType, data)
	}()
}

func (db *ExternalDBClient) SaveEvent(eventType string, data map[string]interface{}) error {
	if !db.IsConfigured() {
		return nil
	}

	payload := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Format(time.RFC3339),
		"payload":    data,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", db.URL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if db.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", db.Token))
	}

	resp, err := db.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("external DB devolvió HTTP %d", resp.StatusCode)
	}

	return nil
}
