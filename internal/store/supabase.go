package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SupabaseClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

type SupabaseLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

type SupabaseFactEntry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

type SupabaseDHCPClient struct {
	MAC         string `json:"mac"`
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	Vendor      string `json:"vendor"`
	IsRandomMAC bool   `json:"is_random_mac"`
	LastSeen    string `json:"last_seen"`
}

type VectorMatchResult struct {
	ID         int64                  `json:"id"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata"`
	Similarity float64                `json:"similarity"`
}

func NewSupabaseClient(url, token string) *SupabaseClient {
	url = strings.TrimSuffix(url, "/")
	if !strings.HasPrefix(token, "Bearer ") && token != "" {
		token = "Bearer " + token
	}

	return &SupabaseClient{
		BaseURL:   url,
		AuthToken: token,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// 1. Health Check (Ping)
func (s *SupabaseClient) Ping() error {
	if s.BaseURL == "" {
		return fmt.Errorf("URL de Supabase no configurada")
	}

	req, err := http.NewRequest("GET", s.BaseURL+"/rest/v1/", nil)
	if err != nil {
		return err
	}

	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 401 {
		return fmt.Errorf("HTTP error status: %d", resp.StatusCode)
	}

	return nil
}

// 1. PostgREST API: Registrar Evento del Router
func (s *SupabaseClient) LogEvent(eventType, severity, message string, payload map[string]interface{}) error {
	if s.BaseURL == "" {
		return nil
	}

	entry := SupabaseLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Severity:  severity,
		Message:   message,
		Payload:   payload,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	endpoint := s.BaseURL + "/rest/v1/clawrt_events"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Supabase LogEvent %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 1. PostgREST API: Guardar Hecho en Memoria
func (s *SupabaseClient) SaveFact(key, value string) error {
	if s.BaseURL == "" {
		return nil
	}

	entry := SupabaseFactEntry{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	endpoint := s.BaseURL + "/rest/v1/clawrt_facts"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// 2. Supabase Realtime Broadcast (Event Streaming)
func (s *SupabaseClient) BroadcastRealtimeEvent(topic, event string, payload map[string]interface{}) error {
	if s.BaseURL == "" {
		return nil
	}

	endpoint := s.BaseURL + "/realtime/v1/api/broadcast"
	bodyMap := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"topic":   "realtime:" + topic,
				"event":   event,
				"payload": payload,
			},
		},
	}

	data, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// 3. Supabase Edge Functions (Serverless Webhooks & Heavy Computing)
func (s *SupabaseClient) TriggerEdgeFunction(functionName string, payload map[string]interface{}) (string, error) {
	if s.BaseURL == "" {
		return "", fmt.Errorf("URL de Supabase no configurada")
	}

	// Host URL: https://xyz.supabase.co -> https://xyz.supabase.co/functions/v1/functionName
	endpoint := fmt.Sprintf("%s/functions/v1/%s", s.BaseURL, functionName)

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("Edge Function %s error %d: %s", functionName, resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}

// 4. Supabase Storage API: Subir Respaldos (.tar.gz) de OpenWrt
func (s *SupabaseClient) UploadBackupFile(localFilePath string) (string, error) {
	if s.BaseURL == "" {
		return "", fmt.Errorf("URL de Supabase no configurada")
	}

	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo local: %v", err)
	}

	filename := filepath.Base(localFilePath)
	targetObject := fmt.Sprintf("backup-%s-%s", time.Now().Format("20060102-150405"), filename)
	endpoint := fmt.Sprintf("%s/storage/v1/object/router-backups/%s", s.BaseURL, targetObject)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(fileData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("x-upsert", "true")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("Supabase Storage error %d: %s", resp.StatusCode, string(body))
	}

	fileURL := fmt.Sprintf("%s/storage/v1/object/public/router-backups/%s", s.BaseURL, targetObject)
	return fileURL, nil
}

// 5. Supabase Vector (pgvector RAG Semantic Match)
func (s *SupabaseClient) MatchVectorDocs(queryEmbedding []float64, matchCount int) ([]VectorMatchResult, error) {
	if s.BaseURL == "" {
		return nil, fmt.Errorf("URL de Supabase no configurada")
	}

	if matchCount <= 0 {
		matchCount = 3
	}

	endpoint := s.BaseURL + "/rest/v1/rpc/match_openwrt_docs"
	reqMap := map[string]interface{}{
		"query_embedding": queryEmbedding,
		"match_threshold": 0.70,
		"match_count":     matchCount,
	}

	data, err := json.Marshal(reqMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", s.AuthToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Supabase pgvector error %d: %s", resp.StatusCode, string(respBody))
	}

	var results []VectorMatchResult
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Multipart helper if needed
func createMultipartBody(fieldName, filename string, fileData []byte) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, "", err
	}
	_, err = part.Write(fileData)
	if err != nil {
		return nil, "", err
	}
	_ = writer.Close()
	return body, writer.FormDataContentType(), nil
}
