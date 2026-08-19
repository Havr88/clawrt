package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CloudflareClient struct {
	AccountID    string
	APIToken     string
	D1DatabaseID string
	R2BucketName string
	HTTPClient   *http.Client
}

func NewCloudflareClient(accountID, apiToken, d1DatabaseID, r2Bucket string) *CloudflareClient {
	if !strings.HasPrefix(apiToken, "Bearer ") && apiToken != "" {
		apiToken = "Bearer " + apiToken
	}

	return &CloudflareClient{
		AccountID:    accountID,
		APIToken:     apiToken,
		D1DatabaseID: d1DatabaseID,
		R2BucketName: r2Bucket,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *CloudflareClient) Ping() error {
	if c.AccountID == "" || c.APIToken == "" {
		return fmt.Errorf("Account ID o API Token de Cloudflare no configurados")
	}

	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s", c.AccountID, c.D1DatabaseID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.APIToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 401 {
		return fmt.Errorf("Cloudflare D1 HTTP error: %d", resp.StatusCode)
	}

	return nil
}

// 1. Cloudflare D1 (SQL Query execution via REST API)
func (c *CloudflareClient) ExecuteD1Query(sql string, params []interface{}) ([]map[string]interface{}, error) {
	if c.AccountID == "" || c.D1DatabaseID == "" {
		return nil, fmt.Errorf("D1 Database ID o Account ID no configurados")
	}

	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query", c.AccountID, c.D1DatabaseID)
	bodyMap := map[string]interface{}{
		"sql":    sql,
		"params": params,
	}

	data, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Cloudflare D1 error %d: %s", resp.StatusCode, string(respBody))
	}

	var resObj struct {
		Result []struct {
			Results []map[string]interface{} `json:"results"`
			Success bool                     `json:"success"`
		} `json:"result"`
		Success bool `json:"success"`
	}

	_ = json.Unmarshal(respBody, &resObj)
	if len(resObj.Result) > 0 {
		return resObj.Result[0].Results, nil
	}

	return nil, nil
}

// 2. Cloudflare R2 Storage API (S3 compatible file upload)
func (c *CloudflareClient) UploadR2Backup(localFilePath string) (string, error) {
	if c.AccountID == "" || c.R2BucketName == "" {
		return "", fmt.Errorf("Cloudflare R2 Bucket o Account ID no configurados")
	}

	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo local: %v", err)
	}

	filename := filepath.Base(localFilePath)
	targetObject := fmt.Sprintf("backup-%s-%s", time.Now().Format("20060102-150405"), filename)
	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/r2/buckets/%s/objects/%s", c.AccountID, c.R2BucketName, targetObject)

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(fileData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("Authorization", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Cloudflare R2 error %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Sprintf("R2://%s/%s", c.R2BucketName, targetObject), nil
}
