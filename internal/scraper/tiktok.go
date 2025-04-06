package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type TikTokScraper struct {
	client *http.Client
}

func NewTikTokScraper() *TikTokScraper {
	return &TikTokScraper{
		client: http.DefaultClient,
	}
}

type tikWMResponse struct {
	Data struct {
		Play   string   `json:"play"`
		Wmplay string   `json:"wmplay"`
		Music  string   `json:"music"`
		Images []string `json:"images"`
	} `json:"data"`
}

func (t *TikTokScraper) DownloadVideo(tiktokURL string) (map[string]interface{}, error) {
	formData := url.Values{
		"url":    {tiktokURL},
		"count":  {"12"},
		"cursor": {"0"},
		"web":    {"1"},
		"hd":     {"1"},
	}

	req, err := http.NewRequest(
		"POST",
		"https://www.tikwm.com/api/",
		bytes.NewBufferString(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header = http.Header{
		"Accept":       {"application/json"},
		"Content-Type": {"application/x-www-form-urlencoded"},
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"},
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // Limit to 10MB
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result tikWMResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	response := make(map[string]interface{}, 4)
	response["status"] = true
	response["wm"] = "https://www.tikwm.com" + result.Data.Wmplay
	response["music"] = "https://www.tikwm.com" + result.Data.Music
	response["video"] = "https://www.tikwm.com" + result.Data.Play

	if len(result.Data.Images) > 0 {
		response["images"] = result.Data.Images
	}

	return response, nil
}
