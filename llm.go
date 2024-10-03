package convogen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ---
// --- Base types
// ---

type ChatModel interface {
	Generate(userPrompt string) (string, error)
}

// ---
// --- OpenAI
// ---

type OpenAIChatModel struct {
	model          string
	apiKey         string
	systemMessages []string
}

var _ ChatModel = &OpenAIChatModel{}

type openAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAiChatResponse struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens            int `json:"prompt_tokens"`
		CompletionTokens        int `json:"completion_tokens"`
		TotalTokens             int `json:"total_tokens"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

func (oai *OpenAIChatModel) Generate(userPrompt string) (string, error) {
	// Create request
	var openAIMessages []openAIMessage
	for _, msg := range oai.systemMessages {
		openAIMessages = append(openAIMessages, openAIMessage{
			Role:    "system",
			Content: msg,
		})
	}
	openAIMessages = append(openAIMessages, openAIMessage{
		Role:    "user",
		Content: userPrompt,
	})
	openAIReq := openAIChatRequest{
		Model:    oai.model,
		Messages: openAIMessages,
	}

	// Make the request
	var openAIRes openAiChatResponse
	err := bearerHttpRequest(oai.apiKey, "https://api.openai.com/v1/chat/completions", http.MethodPost, openAIReq, &openAIRes)
	if err != nil {
		return "", err
	}

	// Get the first choice
	return openAIRes.Choices[0].Message.Content, nil
}

func NewGPT4oModel(apiKey string, systemMessages ...string) *OpenAIChatModel {
	return &OpenAIChatModel{
		model:          "gpt-4o",
		apiKey:         apiKey,
		systemMessages: systemMessages,
	}
}

// ---
// --- Utility functions
// ---

func bearerHttpRequest(token, url, method string, jsonReq any, jsonResPtr any) error {
	// Construct request
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(jsonReq)
	if err != nil {
		log.Err(err).
			Str("url", url).
			Interface("obj", jsonReq).
			Msg("could not marshal json for request")
		return err
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		log.Err(err).
			Str("url", url).
			Msg("bearer http request failed")
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Err(err).
			Str("url", url).
			Msg("request init failed")
		return err
	}

	// Non 2xx we take as an error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		err = fmt.Errorf("non 2xx status: %d", res.StatusCode)
		log.Err(err).
			Str("url", url).
			Msg("non 2xx response")
		return err
	}

	// Read the response
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Err(err).
			Str("url", url).
			Msg("failed to read response body")
		return err
	}
	err = json.Unmarshal(resBody, jsonResPtr)
	if err != nil {
		log.Err(err).
			Str("url", url).
			Str("response_body", string(resBody)).
			Msg("failed to unmarshal response body")
		return err
	}

	return nil
}
