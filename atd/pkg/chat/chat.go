package chat

import (
	"context"
)

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ActionRecord tracks user decisions on previous proposals.
type ActionRecord struct {
	ProposalID string `json:"proposal_id"`
	AtomID     string `json:"atom_id"`
	Action     string `json:"action"` // "ACCEPTED" or "REJECTED"
	Summary    string `json:"summary"`
}

// ChatRequest is the generalized request for LLM chat.
type ChatRequest struct {
	Messages    []Message                `json:"messages"`
	Model       string                   `json:"model"`
	System      string                   `json:"system,omitempty"`
	AtdContext  []map[string]interface{} `json:"atd_context,omitempty"`
	Actions     []ActionRecord           `json:"actions,omitempty"`
	OmitHistory bool                     `json:"omit_history,omitempty"`
	APIKey      string                   `json:"-"` // Not serialized, injected by backend
}

// Proposal represents a suggested change to an ATD atom.
type Proposal struct {
	Action        string                 `json:"action"` // CREATE, UPDATE, DELETE
	AtomID        string                 `json:"atom_id"`
	Content       map[string]interface{} `json:"content"`
	ImpactSummary string                 `json:"impact_summary"`
}

// UsageRecord tracks token consumption.
type UsageRecord struct {
	PromptTokens     int `json:"prompt_tokens"`
	CandidatesTokens int `json:"candidates_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse is the generalized response from an LLM.
type ChatResponse struct {
	Message   string      `json:"message"`
	Proposals []Proposal  `json:"proposals"`
	Usage     UsageRecord `json:"usage"`
}

// ModelInfo describes an available LLM model.
type ModelInfo struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Provider    string   `json:"provider"`
}

// Provider is the interface for different LLM backends.
type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
