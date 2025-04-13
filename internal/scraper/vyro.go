package scraper

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type VyroScraper struct {
	client *http.Client
}

func NewVyroScraper() *VyroScraper {
	return &VyroScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (v *VyroScraper) EnhanceImage(imageData []byte, action string) ([]byte, error) {
	const baseURL = "https://inferenceengine.vyro.ai/"

	validActions := map[string]bool{
		"enhance": true,
		"recolor": true,
		"dehaze":  true,
	}

	if !validActions[action] {
		action = "enhance"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := v.prepareRequest(writer, imageData); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+action, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "okhttp/4.9.3")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (v *VyroScraper) prepareRequest(writer *multipart.Writer, imageData []byte) error {
	writer.WriteField("model_version", "1")

	part, err := writer.CreateFormFile("image", "enhance_image.jpg")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
