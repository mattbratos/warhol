package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
)

type openAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type openAIImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type openAIImageResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newOpenAIClient() (*openAIClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &openAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 180 * time.Second},
	}, nil
}

func (c *openAIClient) generateImage(model string, prompt string, size string, quality string) ([]byte, error) {
	reqBody, err := json.Marshal(openAIImageRequest{
		Model:          model,
		Prompt:         prompt,
		Size:           size,
		Quality:        quality,
		ResponseFormat: "b64_json",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/images/generations", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload openAIImageResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if payload.Error != nil && payload.Error.Message != "" {
			return nil, fmt.Errorf("openai error: %s", payload.Error.Message)
		}
		return nil, fmt.Errorf("openai request failed with status %d", resp.StatusCode)
	}

	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("openai response did not include image data")
	}

	if payload.Data[0].B64JSON != "" {
		imageBytes, err := base64.StdEncoding.DecodeString(payload.Data[0].B64JSON)
		if err != nil {
			return nil, fmt.Errorf("decode image bytes: %w", err)
		}
		return imageBytes, nil
	}

	if payload.Data[0].URL != "" {
		return c.downloadImage(payload.Data[0].URL)
	}

	return nil, fmt.Errorf("openai response had no supported image payload")
}

func (c *openAIClient) downloadImage(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
