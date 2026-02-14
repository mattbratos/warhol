package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"
)

const defaultGoogleBaseURL = "https://generativelanguage.googleapis.com/v1beta"

type googleClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type googleGenerateRequest struct {
	Contents []googleContent `json:"contents"`
}

type googleContent struct {
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text,omitempty"`
}

type googleGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mimeType,omitempty"`
					Data     string `json:"data,omitempty"`
				} `json:"inlineData,omitempty"`
				InlineDataSnake *struct {
					MimeType string `json:"mime_type,omitempty"`
					Data     string `json:"data,omitempty"`
				} `json:"inline_data,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newGoogleClient() (*googleClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY (or GOOGLE_API_KEY) is required")
	}

	baseURL := strings.TrimSpace(os.Getenv("GEMINI_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultGoogleBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &googleClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 180 * time.Second},
	}, nil
}

func (c *googleClient) generateImage(model string, prompt string) ([]byte, error) {
	reqBody, err := json.Marshal(googleGenerateRequest{
		Contents: []googleContent{
			{
				Parts: []googlePart{
					{Text: prompt},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"%s/models/%s:generateContent?key=%s",
		c.baseURL,
		neturl.PathEscape(model),
		neturl.QueryEscape(c.apiKey),
	)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
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

	var payload googleGenerateResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if payload.Error != nil && payload.Error.Message != "" {
			return nil, fmt.Errorf("google error: %s", payload.Error.Message)
		}
		return nil, fmt.Errorf("google request failed with status %d", resp.StatusCode)
	}

	for _, candidate := range payload.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				return decodeBase64Image(part.InlineData.Data)
			}
			if part.InlineDataSnake != nil && part.InlineDataSnake.Data != "" {
				return decodeBase64Image(part.InlineDataSnake.Data)
			}
		}
	}

	return nil, fmt.Errorf("google response did not include image data")
}

func decodeBase64Image(data string) ([]byte, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decode image bytes: %w", err)
	}
	return imageBytes, nil
}
