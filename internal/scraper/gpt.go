package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiBaseURL      = "https://nexra.aryahcr.cc/api/chat"
	pollInterval    = 1 * time.Second
	maxPollAttempts = 60
	defaultModel    = "GPT-4"
	requestTimeout  = 10 * time.Second
	pollTimeout     = 5 * time.Second
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Prompt   string    `json:"prompt"`
	Model    string    `json:"model"`
	Markdown bool      `json:"markdown"`
}

type TaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	GPT    string `json:"gpt"`
	Error  string `json:"error,omitempty"`
}

type ChatResult struct {
	Status       bool        `json:"status"`
	Message      string      `json:"message"`
	FullResponse interface{} `json:"full_response,omitempty"`
	Error        error       `json:"error,omitempty"`
}

type GPTScraper struct {
	client *http.Client
}

func NewGPTScraper() *GPTScraper {
	return &GPTScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *GPTScraper) Chat(prompt string, messages []Message, model string) (ChatResult, error) {
	if prompt == "" {
		return ChatResult{}, errors.New("prompt cannot be empty")
	}

	if model == "" {
		model = "GPT-4"
	}

	conversation := append(messages, Message{
		Role:    "user",
		Content: prompt,
	})

	requestBody := ChatRequest{
		Messages: conversation,
		Prompt:   prompt,
		Model:    model,
		Markdown: false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return ChatResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := g.client.Post(apiBaseURL+"/gpt", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return ChatResult{}, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChatResult{}, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var startResponse TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&startResponse); err != nil {
		return ChatResult{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if startResponse.ID == "" {
		return ChatResult{}, errors.New("no task ID received")
	}

	result, err := g.pollTaskResult(startResponse.ID)
	if err != nil {
		return ChatResult{}, fmt.Errorf("polling failed: %w", err)
	}

	return ChatResult{
		Status:       true,
		Message:      result.GPT,
		FullResponse: result,
	}, nil
}

func (g *GPTScraper) pollTaskResult(taskID string) (*TaskResponse, error) {
	client := &http.Client{Timeout: pollTimeout}

	for attempt := 0; attempt < maxPollAttempts; attempt++ {
		resp, err := client.Get(fmt.Sprintf("%s/task/%s", apiBaseURL, taskID))
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var taskResponse TaskResponse
		if err := json.Unmarshal(body, &taskResponse); err != nil {
			return nil, err
		}

		switch taskResponse.Status {
		case "completed":
			return &taskResponse, nil
		case "error":
			return nil, errors.New(taskResponse.Error)
		case "not_found":
			return nil, errors.New("task not found")
		case "pending":
			time.Sleep(pollInterval)
			continue
		default:
			return nil, fmt.Errorf("unknown status: %s", taskResponse.Status)
		}
	}

	return nil, errors.New("max polling attempts reached")
}
