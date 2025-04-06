package scraper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type YouTubeScraper struct {
	client *http.Client
}

func NewYouTubeScraper() *YouTubeScraper {
	return &YouTubeScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type VideoInfo struct {
	Title     string  `json:"title"`
	Duration  float64 `json:"duration"`
	Thumbnail string  `json:"thumbnail"`
	Author    string  `json:"author"`
}

type DownloadResult struct {
	Status    bool    `json:"status"`
	Title     string  `json:"title"`
	Time      float64 `json:"time"`
	URL       string  `json:"url"`
	Message   string  `json:"message,omitempty"`
	Thumbnail string  `json:"thumbnail"`
}

func (y *YouTubeScraper) Info(url string) (*VideoInfo, error) {
	reqBody := map[string]string{
		"url":      url,
		"platform": "youtube",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "https://ytb2mp4.com/api/youtube-video-info", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %v", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("referer", "https://ytb2mp4.com/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var apiResponse struct {
		Data struct {
			Title     string  `json:"title"`
			Duration  float64 `json:"duration"`
			Author    string  `json:"author"`
			Thumbnail string  `json:"thumbnail"`
		} `json:"data"`
		Status  bool   `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &VideoInfo{
		Title:     apiResponse.Data.Title,
		Duration:  apiResponse.Data.Duration,
		Thumbnail: apiResponse.Data.Thumbnail,
		Author:    apiResponse.Data.Author,
	}, nil
}

func (y *YouTubeScraper) decryptData(encrypted string) (map[string]interface{}, error) {
	key, _ := hex.DecodeString("C5D58EF67A7584E4A29F6C35BBC4EB12")
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	iv := data[:16]
	content := data[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(content, content)

	padLen := int(content[len(content)-1])
	content = content[:len(content)-padLen]

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (y *YouTubeScraper) makeRequest(method, url string, data interface{}) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if method == "POST" {
		jsonBody, _ := json.Marshal(data)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
		q := req.URL.Query()
		for k, v := range data.(map[string]string) {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://yt.savetube.me")
	req.Header.Set("referer", "https://yt.savetube.me/")
	req.Header.Set("user-agent", "Postify/1.0.0")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (y *YouTubeScraper) extractYouTubeID(url string) (string, error) {
	re := regexp.MustCompile(`(?:youtu\.be\/|youtube\.com\/(?:watch\?v=|embed\/|v\/|shorts\/))([a-zA-Z0-9_-]{11})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract video ID from URL")
	}
	return matches[1], nil
}

func (y *YouTubeScraper) Download(link, format string) (string, error) {
	apiBase := "https://media.savetube.me/api"
	apiCDN := "/random-cdn"
	apiInfo := "/v2/info"
	apiDownload := "/download"

	videoID, err := y.extractYouTubeID(link)
	if err != nil {
		return "", err
	}

	cdnRes, err := y.makeRequest("GET", apiBase+apiCDN, map[string]string{})
	if err != nil {
		return "", err
	}
	cdn := cdnRes["cdn"].(string)

	infoRes, err := y.makeRequest("POST", "https://"+cdn+apiInfo, map[string]string{
		"url": "https://www.youtube.com/watch?v=" + videoID,
	})
	if err != nil {
		return "", err
	}

	decrypted, err := y.decryptData(infoRes["data"].(string))
	if err != nil {
		return "", err
	}

	downloadType := "video"
	quality := format
	if format == "mp3" {
		downloadType = "audio"
		quality = "128"
	}

	downloadRes, err := y.makeRequest("POST", "https://"+cdn+apiDownload, map[string]string{
		"id":           videoID,
		"downloadType": downloadType,
		"quality":      quality,
		"key":          decrypted["key"].(string),
	})
	if err != nil {
		return "", err
	}

	return downloadRes["data"].(map[string]interface{})["downloadUrl"].(string), nil
}

func (y *YouTubeScraper) Audio(url string) (*DownloadResult, error) {
	info, err := y.Info(url)
	if err != nil {
		return nil, err
	}

	downloadURL, err := y.Download(url, "mp3")
	if err != nil {
		return &DownloadResult{
			Status:  false,
			Message: "failed to download audio: " + err.Error(),
		}, nil
	}

	return &DownloadResult{
		Status:    true,
		Title:     info.Title,
		Time:      info.Duration / 60,
		URL:       downloadURL,
		Thumbnail: info.Thumbnail,
	}, nil
}

func (y *YouTubeScraper) Video(url, quality string) (*DownloadResult, error) {
	info, err := y.Info(url)
	if err != nil {
		return nil, err
	}

	downloadURL, err := y.Download(url, quality)
	if err != nil {
		return &DownloadResult{
			Status:  false,
			Message: "failed to download video: " + err.Error(),
		}, nil
	}

	return &DownloadResult{
		Status: true,
		Title:  info.Title,
		Time:   info.Duration / 60,
		URL:    downloadURL,
	}, nil
}
