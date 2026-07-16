package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GenerateRequest is the Ollama /api/generate request body.
type GenerateRequest struct {
	Model   string      `json:"model"`
	Prompt  string      `json:"prompt"`
	Stream  bool        `json:"stream"`
	Format  interface{} `json:"format,omitempty"` // string "json" or JSON schema object
	Options *Options    `json:"options,omitempty"`
}

type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}

// GenerateResponse is the Ollama /api/generate response.
type GenerateResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

// EmbeddingRequest is the Ollama /api/embeddings request body.
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse is the Ollama /api/embeddings response.
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// TagsResponse is the Ollama /api/tags response.
type TagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// Generate sends a generation request to an Ollama endpoint. Package var
// (mirrors the ListModels seam below) so tests can swap in a deterministic
// fake responder with no network — see pkg/testutil/fakeprovider, built for
// WP-6 (test_atd_07_26.md §3.5/§6) to unlock audit/dissect/map/recon parser
// tests without a live model.
var Generate = generateHTTP

func generateHTTP(baseURL, model, prompt string, format interface{}, opts *Options) (*GenerateResponse, error) {
	req := GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Format:  format,
		Options: opts,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/generate", baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(raw))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gen GenerateResponse
	if err := json.Unmarshal(raw, &gen); err != nil {
		return nil, err
	}
	return &gen, nil
}

// Embed sends an embedding request to an Ollama endpoint. Package var for
// the same reason as Generate above (see pkg/testutil/fakeprovider).
var Embed = embedHTTP

func embedHTTP(baseURL, model, text string) ([]float32, error) {
	req := EmbeddingRequest{
		Model:  model,
		Prompt: text,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/embeddings", baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(raw))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var emb EmbeddingResponse
	if err := json.Unmarshal(raw, &emb); err != nil {
		return nil, err
	}
	return emb.Embedding, nil
}

// ListModels queries /api/tags to get available models on an endpoint.
var ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
	client := http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
	resp, err := client.Get(fmt.Sprintf("%s/api/tags", baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama tags error: status %d", resp.StatusCode)
	}

	var tags TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range tags.Models {
		models = append(models, m.Name)
	}
	return models, nil
}
