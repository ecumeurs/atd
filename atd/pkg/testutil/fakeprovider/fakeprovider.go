// Package fakeprovider is the WP-6 deliverable (test_atd_07_26.md §3.5, §6):
// a deterministic, no-network stand-in for the LLM boundary so
// audit/dissect/map/recon logic and the webui chat path can be tested
// without a model, network, or GPU.
//
// The product actually has two separate LLM seams, and this package covers
// both:
//
//  1. chat.Provider (pkg/chat/chat.go) — used by pkg/llmservice's webui chat
//     path. llmservice.Service already has a real injection point for this:
//     (*Service).AddProvider populates s.providers directly, and
//     (*Service).getProviders() returns early ("if len(s.providers) > 0")
//     without touching env vars or config, so a test can do:
//
//     svc := llmservice.NewService(cfg)
//     svc.AddProvider(&fakeprovider.ChatProvider{...})
//     resp, err := svc.Chat(ctx, req, nil)
//
//  2. ollama.Query / ollama.QueryEmbed (pkg/ollama) — the seam actually used
//     by every CLI/pkg command that talks to an LLM: audit, dissect, map,
//     reconcile, congruence, compare, fix, generate, coverage's semantic
//     check, exploration's assemble/search, and the indexer. None of these
//     go through chat.Provider or llmservice at all — they call
//     ollama.Query(taskType, prompt, format) directly. That function
//     resolves a provider via ollama.ResolveProvider (config-driven, already
//     a seam) and then calls the package-level ollama.Generate/ollama.Embed
//     functions to actually hit the network.
//
//     Prior to WP-6, Generate/Embed were plain functions — no seam existed
//     to fake the actual LLM call, only provider *resolution*
//     (pkg/ollama/provider_test.go swaps ListModels to fake availability,
//     but a resolved call still made a real HTTP POST). WP-6's product-side
//     change (pkg/ollama/client.go) turns Generate and Embed into package
//     vars, mirroring the ListModels var immediately above them in the same
//     file — the smallest possible change that makes the real call site
//     fakeable. InstallOllama below is the seam's test-side counterpart: it
//     points config.ActiveConfig.LLM at a fake provider that "has" every
//     model needed for every task type, then swaps ListModels/Generate/Embed
//     so no HTTP ever leaves the process.
package fakeprovider

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/chat"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/testutil"
)

// ─────────────────────────────────────────────────────────────────────────
// 1. chat.Provider fake — for llmservice.Service.AddProvider injection.
// ─────────────────────────────────────────────────────────────────────────

// ChatProvider is a canned chat.Provider. Every field is optional: a zero
// value ChatProvider.Chat returns a canned success response, and
// ChatProvider.ListModels returns Models (nil is fine — an empty list).
// Set ChatFunc/ModelsFunc for per-test canned/malformed behavior.
type ChatProvider struct {
	NameStr    string
	ChatFunc   func(ctx context.Context, req chat.ChatRequest) (*chat.ChatResponse, error)
	ModelsFunc func(ctx context.Context) ([]chat.ModelInfo, error)
	Models     []chat.ModelInfo

	mu    sync.Mutex
	Calls []chat.ChatRequest // every request seen, in order — assert prompt assembly against this
}

// Name implements chat.Provider.
func (f *ChatProvider) Name() string {
	if f.NameStr == "" {
		return "fake"
	}
	return f.NameStr
}

// Chat implements chat.Provider. Records the request (for assertions) before
// dispatching to ChatFunc, or a default canned response if unset.
func (f *ChatProvider) Chat(ctx context.Context, req chat.ChatRequest) (*chat.ChatResponse, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, req)
	f.mu.Unlock()

	if f.ChatFunc != nil {
		return f.ChatFunc(ctx, req)
	}
	return &chat.ChatResponse{Message: "fake response", Proposals: []chat.Proposal{}}, nil
}

// ListModels implements chat.Provider.
func (f *ChatProvider) ListModels(ctx context.Context) ([]chat.ModelInfo, error) {
	if f.ModelsFunc != nil {
		return f.ModelsFunc(ctx)
	}
	return f.Models, nil
}

// CallCount returns the number of Chat calls seen so far (race-safe).
func (f *ChatProvider) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Calls)
}

// ─────────────────────────────────────────────────────────────────────────
// 2. ollama.Generate/ollama.Embed fake — the seam audit/dissect/map/recon/
//    search actually run on. InstallOllama is the one-call setup: it snapshots
//    and restores config.ActiveConfig (via testutil.SnapshotConfig — the T-1
//    idiom from incident I-1, test_atd_07_26.md §2.1/§3.6) and restores the
//    three swapped ollama package vars on t.Cleanup, so tests using this
//    never leak state into a sibling test in the same binary.
// ─────────────────────────────────────────────────────────────────────────

// GenerateCall records one ollama.Generate invocation the fake observed.
type GenerateCall struct {
	Prompt string
	Format interface{}
}

// Ollama is the fake LLM backend wired into the ollama package's swappable
// Generate/Embed/ListModels vars. Configure behavior with SetJSON,
// SetSequence, SetGenerateFunc, SetError, SetEmbedVector, or SetEmbedTable;
// the zero value errors loudly on any call rather than silently returning a
// zero-value success (the exact failure mode WP-6/S12 tests guard against
// in the *product* parsers — the fake itself must not paper over a
// misconfigured test).
type Ollama struct {
	t testing.TB

	mu           sync.Mutex
	generateFunc func(prompt string, format interface{}) (*ollama.GenerateResponse, error)
	embedFunc    func(text string) ([]float32, error)
	calls        []GenerateCall
}

// InstallOllama points config.ActiveConfig.LLM at a synthetic provider that
// "has" a model for every task type (code_analysis, text_analysis,
// text_generation, embed — the task strings used across cmd/atd/cmd and
// pkg/audit, pkg/coverage, pkg/exploration), then swaps
// ollama.ListModels/Generate/Embed so ollama.Query/ollama.QueryEmbed never
// make an HTTP call. Call one of the Set* methods before exercising product
// code, or every call fails loudly with "no GenerateFunc configured".
func InstallOllama(t testing.TB) *Ollama {
	t.Helper()

	testutil.SnapshotConfig(t)

	origListModels := ollama.ListModels
	origGenerate := ollama.Generate
	origEmbed := ollama.Embed
	t.Cleanup(func() {
		ollama.ListModels = origListModels
		ollama.Generate = origGenerate
		ollama.Embed = origEmbed
	})

	f := &Ollama{t: t}

	config.ActiveConfig.LLM = config.LLMConfig{
		Providers: []config.LLMProvider{
			{Name: "fakeprovider", BaseURL: "fake://ollama"},
		},
		Models: map[string]config.ModelConfig{
			"fake-model": {Tasks: []string{"code_analysis", "text_analysis", "text_generation"}},
			"fake-embed": {Tasks: []string{"embed"}},
		},
	}

	ollama.ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
		return []string{"fake-model", "fake-embed"}, nil
	}

	ollama.Generate = func(baseURL, model, prompt string, format interface{}, opts *ollama.Options) (*ollama.GenerateResponse, error) {
		f.mu.Lock()
		f.calls = append(f.calls, GenerateCall{Prompt: prompt, Format: format})
		fn := f.generateFunc
		f.mu.Unlock()

		if fn == nil {
			return nil, fmt.Errorf("fakeprovider: Ollama.Generate called with no GenerateFunc/SetJSON/SetSequence configured (prompt: %.120s)", prompt)
		}
		return fn(prompt, format)
	}

	ollama.Embed = func(baseURL, model, text string) ([]float32, error) {
		f.mu.Lock()
		fn := f.embedFunc
		f.mu.Unlock()

		if fn == nil {
			return nil, fmt.Errorf("fakeprovider: Ollama.Embed called with no EmbedFunc/SetEmbedVector/SetEmbedTable configured")
		}
		return fn(text)
	}

	return f
}

// SetJSON makes every subsequent Generate call return resp verbatim as the
// GenerateResponse.Response field — the common case of "one canned
// schema-valid (or deliberately malformed) response for this test".
func (f *Ollama) SetJSON(resp string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.generateFunc = func(prompt string, format interface{}) (*ollama.GenerateResponse, error) {
		return &ollama.GenerateResponse{Response: resp}, nil
	}
}

// SetSequence configures a queue of canned responses consumed in order —
// for commands that make more than one LLM call per invocation (e.g. `atd
// map` discover mode: intent extraction, then link recommendation). Calling
// past the end of the sequence errors loudly rather than panicking or
// reusing the last response, so a test that assumes N calls and gets N+1
// fails instead of silently passing on stale data.
func (f *Ollama) SetSequence(resps ...string) {
	var mu sync.Mutex
	idx := 0

	f.mu.Lock()
	f.generateFunc = func(prompt string, format interface{}) (*ollama.GenerateResponse, error) {
		mu.Lock()
		defer mu.Unlock()
		if idx >= len(resps) {
			return nil, fmt.Errorf("fakeprovider: SetSequence exhausted after %d call(s), got call %d", len(resps), idx+1)
		}
		r := resps[idx]
		idx++
		return &ollama.GenerateResponse{Response: r}, nil
	}
	f.mu.Unlock()
}

// SetGenerateFunc installs a custom responder, e.g. to branch on prompt
// content or to return distinct responses for distinct task-shaped prompts
// in a single test.
func (f *Ollama) SetGenerateFunc(fn func(prompt string, format interface{}) (*ollama.GenerateResponse, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.generateFunc = fn
}

// SetError makes every subsequent Generate call fail with err — for testing
// the ollama.ErrIDEFallback / generic-error branches product code takes when
// the LLM call itself fails (as opposed to succeeding with malformed JSON).
func (f *Ollama) SetError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.generateFunc = func(prompt string, format interface{}) (*ollama.GenerateResponse, error) {
		return nil, err
	}
}

// SetEmbedVector makes every subsequent Embed call return the same fixed
// vector regardless of input text — enough for tests that only need *an*
// embedding to exist (e.g. driving the audit collision path's shape without
// caring about specific similarity values).
func (f *Ollama) SetEmbedVector(vec []float32) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.embedFunc = func(text string) ([]float32, error) { return vec, nil }
}

// SetEmbedTable configures fixed vectors keyed by exact input text, for
// tests that need controlled cosine similarity between specific inputs
// (e.g. S12/search embedding-mode ranking, or pkg/audit's collision-map
// path). A lookup miss errors loudly rather than falling back to a zero
// vector, since a silent zero vector would quietly produce a similarity of
// 0 instead of failing the test that forgot to seed it.
func (f *Ollama) SetEmbedTable(table map[string][]float32) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.embedFunc = func(text string) ([]float32, error) {
		if v, ok := table[text]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("fakeprovider: no embed vector configured for text (len %d): %.80s", len(text), text)
	}
}

// Calls returns every Generate invocation observed so far, in order.
func (f *Ollama) Calls() []GenerateCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]GenerateCall, len(f.calls))
	copy(out, f.calls)
	return out
}
