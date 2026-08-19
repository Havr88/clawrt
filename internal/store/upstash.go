package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type UpstashClient struct {
	RedisURL    string
	RedisToken  string
	VectorURL   string
	VectorToken string
	QStashToken string
	HTTPClient  *http.Client
}

func NewUpstashClient(redisURL, redisToken, vectorURL, vectorToken, qstashToken string) *UpstashClient {
	redisURL = strings.TrimSuffix(redisURL, "/")
	vectorURL = strings.TrimSuffix(vectorURL, "/")
	if !strings.HasPrefix(redisToken, "Bearer ") && redisToken != "" {
		redisToken = "Bearer " + redisToken
	}
	if !strings.HasPrefix(vectorToken, "Bearer ") && vectorToken != "" {
		vectorToken = "Bearer " + vectorToken
	}

	return &UpstashClient{
		RedisURL:    redisURL,
		RedisToken:  redisToken,
		VectorURL:   vectorURL,
		VectorToken: vectorToken,
		QStashToken: qstashToken,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// 1. Health Check (Ping)
func (u *UpstashClient) Ping() error {
	if u.RedisURL == "" || u.RedisToken == "" {
		return fmt.Errorf("URL o Token de Upstash Redis no configurados")
	}

	endpoint := fmt.Sprintf("%s/ping", u.RedisURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", u.RedisToken)
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Upstash Redis error status: %d", resp.StatusCode)
	}

	return nil
}

// 1. Upstash Redis REST API (Set Key Value)
func (u *UpstashClient) Set(key string, val string) error {
	if u.RedisURL == "" {
		return nil
	}

	endpoint := fmt.Sprintf("%s/set/%s", u.RedisURL, key)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(val))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", u.RedisToken)
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// 1. Upstash Redis REST API (Get Key Value)
func (u *UpstashClient) Get(key string) (string, error) {
	if u.RedisURL == "" {
		return "", fmt.Errorf("Upstash Redis no configurado")
	}

	endpoint := fmt.Sprintf("%s/get/%s", u.RedisURL, key)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", u.RedisToken)
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var resObj struct {
		Result string `json:"result"`
	}
	_ = json.Unmarshal(body, &resObj)
	return resObj.Result, nil
}

// 2. Upstash QStash (Serverless Message Queue & Delayed Alerts)
func (u *UpstashClient) PublishQStash(destinationURL string, payload map[string]interface{}, delaySeconds int) error {
	if u.QStashToken == "" {
		return fmt.Errorf("QStash Token no configurado")
	}

	endpoint := fmt.Sprintf("https://qstash.upstash.io/v2/publish/%s", destinationURL)
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.QStashToken)
	if delaySeconds > 0 {
		req.Header.Set("Upstash-Delay", fmt.Sprintf("%ds", delaySeconds))
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Upstash QStash error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
