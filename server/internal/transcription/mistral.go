package transcription

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const endpoint = "https://api.mistral.ai/v1/audio/transcriptions"

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *Client) Transcribe(ctx context.Context, filename string, source io.Reader) (string, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return "", errors.New("Mistral API key is not configured")
	}

	reader, writer := io.Pipe()
	form := multipart.NewWriter(writer)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", form.FormDataContentType())

	go func() {
		part, err := form.CreateFormFile("file", filename)
		if err == nil {
			_, err = io.Copy(part, source)
		}
		if err == nil {
			err = form.WriteField("model", c.model)
		}
		if err == nil {
			err = form.Close()
		}
		_ = writer.CloseWithError(err)
	}()

	response, err := c.httpClient.Do(request)
	if err != nil {
		_ = reader.CloseWithError(err)
		return "", err
	}
	defer reader.Close()
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		if strings.TrimSpace(string(message)) == "" {
			return "", fmt.Errorf("Mistral transcription failed with status %d", response.StatusCode)
		}
		return "", fmt.Errorf("Mistral transcription failed: %s", strings.TrimSpace(string(message)))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Mistral transcription: %w", err)
	}
	return result.Text, nil
}
