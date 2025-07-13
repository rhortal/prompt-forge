package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"promptforge/internal/config"
	"promptforge/internal/models"
)

type OpenAIService struct {
	client *http.Client
}

func NewOpenAIService() *OpenAIService {
	return &OpenAIService{
		client: &http.Client{},
	}
}

func (s *OpenAIService) CallAzureOpenAI(messages []models.Message, temperature float64, maxTokens int, model string) (string, error) {
	endpoint := config.GetEndpointURL(model)

	// o3 model only supports temperature = 1
	if model == "o3" {
		temperature = 1
	}

	requestBody := models.OpenAIRequest{
		Messages:    messages,
		Temperature: temperature,
	}

	// Initialize both to 0 to ensure only the relevant one is set
	requestBody.MaxTokens = 0
	requestBody.MaxCompletionTokens = 0

	if maxTokens > 0 {
		// o3 model uses max_completion_tokens instead of max_tokens
		if model == "o3" {
			requestBody.MaxCompletionTokens = maxTokens
		} else {
			requestBody.MaxTokens = maxTokens
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.AppConfig.AzureOpenAI.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp models.OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
